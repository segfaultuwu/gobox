package utils

import "strings"

func SplitArgs(line string) []string {
	var args []string
	var current strings.Builder

	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}

		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}

		case ' ', '\t':
			if !inSingle && !inDouble {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
