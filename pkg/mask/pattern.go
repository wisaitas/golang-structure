package mask

import (
	"encoding/json"
	"strings"
)

func ParsePatternMap(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil || len(m) == 0 {
		return nil
	}
	return m
}
