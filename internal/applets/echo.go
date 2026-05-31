package applets

import (
	"fmt"
	"strconv"
	"strings"

	"gobox/internal/utils"
)

func init() {
	Register("echo", "print arguments", Echo)
}

func Echo(args []string) error {
	pArgs := utils.ParseArgs(args)

	enableEscape := pArgs.Has("e")
	disableEscape := pArgs.Has("E")
	noNewline := pArgs.Has("n")

	text := strings.Join(pArgs.Rest, " ")

	if enableEscape && !disableEscape {
		text = parseEchoEscapes(text)
	}

	if noNewline {
		fmt.Print(text)
		return nil
	}

	fmt.Println(text)
	return nil
}

func parseEchoEscapes(text string) string {
	var out strings.Builder

	for i := 0; i < len(text); i++ {
		ch := text[i]

		if ch != '\\' {
			out.WriteByte(ch)
			continue
		}

		if i+1 >= len(text) {
			out.WriteByte('\\')
			break
		}

		i++
		next := text[i]

		switch next {
		case 'n':
			out.WriteByte('\n')

		case 't':
			out.WriteByte('\t')

		case 'r':
			out.WriteByte('\r')

		case 'b':
			out.WriteByte('\b')

		case 'a':
			out.WriteByte('\a')

		case 'v':
			out.WriteByte('\v')

		case 'f':
			out.WriteByte('\f')

		case '\\':
			out.WriteByte('\\')

		case 'e', 'E':
			out.WriteByte(0x1b)

		case '0':
			j := i + 1

			for j < len(text) && j < i+4 && text[j] >= '0' && text[j] <= '7' {
				j++
			}

			if j == i+1 {
				out.WriteByte(0)
				continue
			}

			value, err := strconv.ParseInt(text[i+1:j], 8, 32)
			if err != nil {
				out.WriteByte('\\')
				out.WriteByte('0')
				continue
			}

			out.WriteByte(byte(value))
			i = j - 1

		default:
			out.WriteByte('\\')
			out.WriteByte(next)
		}
	}

	return out.String()
}
