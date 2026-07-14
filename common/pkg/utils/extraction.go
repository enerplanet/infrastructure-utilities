package utils

import "strconv"

// ExtractSessionID extracts session ID from various key formats
func ExtractSessionID(data map[string]interface{}) (int64, bool) {
	if data == nil {
		return 0, false
	}
	keys := []string{"session_id", "sessionId", "job_id", "jobId"}
	for _, key := range keys {
		if value, exists := data[key]; exists {
			if id, ok := tryConvertToInt64(value); ok {
				return id, true
			}
		}
	}
	return 0, false
}

func tryConvertToInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
		return 0, false
	}
	return 0, false
}


// ExtractCallbackURL extracts callback URL from various key formats
func ExtractCallbackURL(data map[string]interface{}) (string, bool) {
	if data == nil {
		return "", false
	}
	keys := []string{"callback_url", "callbackUrl"}
	for _, key := range keys {
		if value, exists := data[key]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str, true
			}
		}
	}
	return "", false
}
