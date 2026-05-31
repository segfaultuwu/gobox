package utils

import "strings"

func CharsToString(chars []int8) string {
	var b strings.Builder

	for _, c := range chars {
		if c == 0 {
			break
		}
		b.WriteByte(byte(c))
	}

	return b.String()
}

func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func Truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}

	if max <= 3 {
		return s[:max]
	}

	return s[:max-3] + "..."
}

func StripANSI(s string) string {
	var out strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if inEscape {
			if ch >= '@' && ch <= '~' {
				inEscape = false
			}
			continue
		}

		if ch == 0x1b {
			inEscape = true
			continue
		}

		out.WriteByte(ch)
	}

	return out.String()
}
