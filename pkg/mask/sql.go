package mask

import (
	"regexp"
	"strings"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// MaskSQLLogLine masks literal values in INSERT ... VALUES for columns listed in colMask.
// Keys are column names (case-insensitive); values are MaskPlainString patterns.
// If colMask is empty or the line is not a recognizable INSERT, the input is returned unchanged (cheap path).
func MaskSQLLogLine(line string, colMask map[string]string) string {
	if len(colMask) == 0 {
		return line
	}
	plain := ansiSeq.ReplaceAllString(line, "")
	if !strings.Contains(strings.ToUpper(plain), "INSERT") || !strings.Contains(strings.ToUpper(plain), "VALUES") {
		return line
	}
	masked, ok := maskInsertValues(plain, colMask)
	if !ok {
		return line
	}
	return masked
}

func maskInsertValues(line string, colMask map[string]string) (string, bool) {
	upper := strings.ToUpper(line)
	valuesIdx := strings.Index(upper, "VALUES")
	if valuesIdx < 0 {
		return "", false
	}
	openCol := strings.LastIndex(line[:valuesIdx], "(")
	if openCol < 0 {
		return "", false
	}
	closeCol := findMatchingParen(line, openCol)
	if closeCol < 0 || closeCol >= valuesIdx {
		return "", false
	}
	cols := parseColumnList(line[openCol+1 : closeCol])
	if len(cols) == 0 {
		return "", false
	}

	rest := strings.TrimSpace(line[valuesIdx+len("VALUES"):])
	if len(rest) == 0 || rest[0] != '(' {
		return "", false
	}
	closeVals := findMatchingParen(rest, 0)
	if closeVals < 0 {
		return "", false
	}
	inner := rest[1:closeVals]
	vals, ok := splitCommaSeparatedSQLValues(inner)
	if !ok || len(vals) != len(cols) {
		return "", false
	}

	outVals := make([]string, len(vals))
	for i := range vals {
		pat := patternForColumn(cols[i], colMask)
		if pat == "" {
			outVals[i] = vals[i]
		} else {
			outVals[i] = maskSQLValueToken(vals[i], pat)
		}
	}
	newValues := "(" + strings.Join(outVals, ",") + ")"
	return line[:valuesIdx+len("VALUES")] + " " + newValues + rest[closeVals+1:], true
}

func patternForColumn(col string, colMask map[string]string) string {
	for k, v := range colMask {
		if strings.EqualFold(strings.TrimSpace(col), strings.TrimSpace(k)) {
			return v
		}
	}
	return ""
}

func parseColumnList(s string) []string {
	s = strings.TrimSpace(s)
	var cols []string
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"', '`':
			inQuote = !inQuote
			b.WriteByte(c)
		case ',':
			if !inQuote {
				cols = append(cols, normalizeSQLIdent(b.String()))
				b.Reset()
				continue
			}
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	if b.Len() > 0 {
		cols = append(cols, normalizeSQLIdent(b.String()))
	}
	return cols
}

func normalizeSQLIdent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	s = strings.Trim(s, "`")
	return s
}

func splitCommaSeparatedSQLValues(inner string) ([]string, bool) {
	var vals []string
	var b strings.Builder
	inString := false
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if c == '\'' {
			if inString {
				if i+1 < len(inner) && inner[i+1] == '\'' {
					b.WriteString("''")
					i++
					continue
				}
				inString = false
			} else {
				inString = true
			}
			b.WriteByte(c)
			continue
		}
		if c == ',' && !inString {
			vals = append(vals, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	if b.Len() > 0 {
		vals = append(vals, strings.TrimSpace(b.String()))
	}
	return vals, true
}

func maskSQLValueToken(tok, pattern string) string {
	tok = strings.TrimSpace(tok)
	if strings.HasPrefix(tok, "'") {
		inner := unquoteSQLString(tok)
		m := MaskPlainString(inner, pattern)
		return quoteSQLString(m)
	}
	return MaskPlainString(tok, pattern)
}

func unquoteSQLString(s string) string {
	if len(s) < 2 || s[0] != '\'' || s[len(s)-1] != '\'' {
		return s
	}
	inner := s[1 : len(s)-1]
	return strings.ReplaceAll(inner, "''", "'")
}

func quoteSQLString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func findMatchingParen(s string, start int) int {
	if start >= len(s) || s[start] != '(' {
		return -1
	}
	depth := 0
	inString := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if c == '\'' {
			if inString && i+1 < len(s) && s[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
