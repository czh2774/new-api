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

// RelayRetryBudget separates the request-wide ceiling from the existing native
// global threshold for switching auto groups. A model override changes only the
// ceiling, not the routing policy within and between native groups.
type RelayRetryBudget struct {
	AdditionalAttempts   int
	GroupSwitchThreshold int
}

func snapshotRelayRetryBudget(method, route, originalModel string, ordinaryOpenAI bool) RelayRetryBudget {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	budget := RelayRetryBudget{AdditionalAttempts: common.RetryTimes, GroupSwitchThreshold: common.RetryTimes}
	if ordinaryOpenAI && method == http.MethodPost && route == "/v1/chat/completions" {
		if snapshot := chatCompletionModelRetries.Load(); snapshot != nil {
			if count, ok := snapshot.(map[string]int)[originalModel]; ok {
				budget.AdditionalAttempts = count
			}
		}
	}
	return budget
}

// SnapshotRelayRetryBudget captures the global fallback and model override under
// the option-update lock. The returned scalar is the request-wide ceiling.
func SnapshotRelayRetryBudget(method, route, originalModel string, ordinaryOpenAI bool) int {
	return snapshotRelayRetryBudget(method, route, originalModel, ordinaryOpenAI).AdditionalAttempts
}

// AdmitRelayRetryBudget shares one admission snapshot across initial selection
// and relay retries. The route guard excludes task, media and alternate formats.
func AdmitRelayRetryBudget(c *gin.Context, originalModel string, ordinaryOpenAI bool) RelayRetryBudget {
	const key = "admitted_relay_retry_budget"
	if budget, exists := c.Get(key); exists {
		return budget.(RelayRetryBudget)
	}
	budget := snapshotRelayRetryBudget(c.Request.Method, c.FullPath(), originalModel, ordinaryOpenAI)
	c.Set(key, budget)
	return budget
}
