package operation_setting

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

const ChatCompletionModelRetryOptionKey = "ChatCompletionModelRetryTimes"

// Match the existing RetryTimes UI range (routing-reliability-section.tsx).
// This is a project configuration bound, not an upstream protocol requirement.
const MaxChatCompletionModelRetryTimes = 10

var chatCompletionModelRetries atomic.Value // immutable map[string]int

func ParseChatCompletionModelRetryTimes(value string) (map[string]int, error) {
	var input map[string]*int
	// Reuse the option CAS strict object decoder and its 64 KiB option limit.
	if err := common.DecodeStrictJSONObject(strings.NewReader(value), &input, 64*1024); err != nil {
		return nil, err
	}
	result := make(map[string]int, len(input))
	for model, count := range input {
		if strings.TrimSpace(model) == "" || strings.TrimSpace(model) != model {
			return nil, fmt.Errorf("retry model must be a nonempty exact canonical model ID")
		}
		if count == nil || *count < 0 || *count > MaxChatCompletionModelRetryTimes {
			return nil, fmt.Errorf("retry count must be an integer from 0 to %d", MaxChatCompletionModelRetryTimes)
		}
		result[model] = *count
	}
	return result, nil
}

func UpdateChatCompletionModelRetryTimes(value string) error {
	policy, err := ParseChatCompletionModelRetryTimes(value)
	if err != nil {
		return err
	}
	chatCompletionModelRetries.Store(policy)
	return nil
}

// SnapshotRelayRetryBudget is called once at admission. Option updates hold the
// same lock, so the global fallback and model override are captured together.
// The returned scalar cannot change when an option is hot-updated in flight.
func SnapshotRelayRetryBudget(method, route, originalModel string, ordinaryOpenAI bool) int {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	budget := common.RetryTimes
	if ordinaryOpenAI && method == http.MethodPost && route == "/v1/chat/completions" {
		if snapshot := chatCompletionModelRetries.Load(); snapshot != nil {
			if count, ok := snapshot.(map[string]int)[originalModel]; ok {
				return count
			}
		}
	}
	return budget
}

// AdmitRelayRetryBudget shares one admission snapshot across initial selection
// and relay retries. The route guard excludes task, media and alternate formats.
func AdmitRelayRetryBudget(c *gin.Context, originalModel string, ordinaryOpenAI bool) int {
	const key = "admitted_relay_retry_budget"
	if budget, exists := c.Get(key); exists {
		return budget.(int)
	}
	budget := SnapshotRelayRetryBudget(c.Request.Method, c.FullPath(), originalModel, ordinaryOpenAI)
	c.Set(key, budget)
	return budget
}
