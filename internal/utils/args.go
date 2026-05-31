package utils

import (
	"fmt"
	"strings"
)

func SplitArgs(line string) ([]string, error) {
	var args []string
	var current strings.Builder

	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range line {
		if escaped {
			switch r {
			case 'n':
				current.WriteRune('\n')
			case 't':
				current.WriteRune('\t')
			case 'r':
				current.WriteRune('\r')
			case '\\', '"', '\'':
				current.WriteRune(r)
			default:
				current.WriteRune('\\')
				current.WriteRune(r)
			}

			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
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

	if escaped {
		return nil, fmt.Errorf("unfinished escape")
	}

	if inSingle {
		return nil, fmt.Errorf("unterminated single quote")
	}

	if inDouble {
		return nil, fmt.Errorf("unterminated double quote")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}
