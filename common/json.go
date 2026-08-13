package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func UnmarshalJsonStr(data string, v any) error {
	return json.Unmarshal(StringToByteSlice(data), v)
}

func DecodeJson(reader io.Reader, v any) error {
	return json.NewDecoder(reader).Decode(v)
}

func rejectDuplicateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			member, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := member.(string)
			if !ok {
				return errors.New("request object member name is invalid")
			}
			if _, duplicate := seen[name]; duplicate {
				return errors.New("request object contains a duplicate member")
			}
			seen[name] = struct{}{}
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("request object is incomplete")
		}
	case '[':
		for decoder.More() {
			if err := rejectDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("request array is incomplete")
		}
	default:
		return errors.New("request contains an invalid delimiter")
	}
	return nil
}

// DecodeStrictJSONObject rejects oversized bodies, duplicate members at any
// depth, unknown struct fields, and trailing JSON values.
func DecodeStrictJSONObject(reader io.Reader, destination any, maxBytes int64) error {
	if maxBytes <= 0 {
		return errors.New("request body limit must be positive")
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return err
	}
	if int64(len(raw)) > maxBytes {
		return errors.New("request body is too large")
	}

	keys := json.NewDecoder(bytes.NewReader(raw))
	opening, err := keys.Token()
	if err != nil || opening != json.Delim('{') {
		return errors.New("request body must be a JSON object")
	}
	seen := make(map[string]struct{})
	for keys.More() {
		member, err := keys.Token()
		if err != nil {
			return err
		}
		key, ok := member.(string)
		if !ok {
			return errors.New("request object member name is invalid")
		}
		if _, duplicate := seen[key]; duplicate {
			return errors.New("request object contains a duplicate member")
		}
		seen[key] = struct{}{}
		if err := rejectDuplicateJSONValue(keys); err != nil {
			return err
		}
	}
	closing, err := keys.Token()
	if err != nil || closing != json.Delim('}') {
		return errors.New("request body must be a JSON object")
	}
	if _, err := keys.Token(); err != io.EOF {
		return errors.New("request body must contain exactly one JSON object")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func GetJsonType(data json.RawMessage) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "unknown"
	}
	firstChar := trimmed[0]
	switch firstChar {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}

// JsonRawMessageToString returns JSON strings as their decoded value and other JSON values as raw text.
func JsonRawMessageToString(data json.RawMessage) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return ""
	}
	if trimmed[0] != '"' {
		return string(trimmed)
	}
	var value string
	if err := Unmarshal(trimmed, &value); err != nil {
		return string(trimmed)
	}
	return value
}
