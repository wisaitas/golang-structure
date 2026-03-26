package mask

import (
	"strconv"
	"strings"
)

func MaskPlainString(s, pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return s
	}
	i := strings.IndexByte(pattern, ':')
	if i <= 0 {
		return pattern
	}
	preStr := strings.TrimSpace(pattern[:i])
	rest := strings.TrimSpace(pattern[i+1:])
	pre, err := strconv.Atoi(preStr)
	if err != nil || pre < 0 {
		return pattern
	}
	if rest != "" && isAllDecimalDigits(rest) {
		suf, err := strconv.Atoi(rest)
		if err != nil || suf < 0 {
			return pattern
		}
		return maskPrefixSuffix(s, pre, suf)
	}
	if rest == "" {
		return pattern
	}
	return maskWithMarker(s, pre, rest)
}

func maskWithMarker(s string, n int, marker string) string {
	if len(marker) == 0 {
		return maskPrefixSuffix(s, n, 0)
	}
	if len(s) >= len(marker) && strings.HasSuffix(s, marker) {
		core := s[:len(s)-len(marker)]
		if len(core) == 0 {
			return s
		}
		return maskPrefixSuffix(core, n, 0) + marker
	}
	if len(s) >= len(marker) && strings.HasPrefix(s, marker) {
		core := s[len(marker):]
		if len(core) == 0 {
			return s
		}
		return marker + maskPrefixSuffix(core, n, 0)
	}
	return maskPrefixSuffix(s, n, 0)
}

func isAllDecimalDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func maskPrefixSuffix(s string, prefix, suffix int) string {
	n := len(s)
	if n == 0 {
		return s
	}
	if prefix+suffix >= n {
		return strings.Repeat("*", n)
	}
	return s[:prefix] + strings.Repeat("*", n-prefix-suffix) + s[n-suffix:]
}
