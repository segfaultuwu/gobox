package utils

import (
	"fmt"
	"strings"
)

type ParsedArgs struct {
	Flags map[string]bool
	Opts  map[string]string
	Rest  []string
}

func ParseArgs(args []string) ParsedArgs {
	out := ParsedArgs{
		Flags: map[string]bool{},
		Opts:  map[string]string{},
		Rest:  []string{},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			out.Rest = append(out.Rest, args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "--") && len(arg) > 2 {
			name := strings.TrimPrefix(arg, "--")

			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				out.Opts[parts[0]] = parts[1]
				continue
			}

			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				out.Opts[name] = args[i+1]
				i++
				continue
			}

			out.Flags[name] = true
			continue
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			shorts := strings.TrimPrefix(arg, "-")

			for _, ch := range shorts {
				out.Flags[string(ch)] = true
			}

			continue
		}

		out.Rest = append(out.Rest, arg)
	}

	return out
}

func (p ParsedArgs) Has(names ...string) bool {
	for _, name := range names {
		if p.Flags[name] {
			return true
		}
	}
	return false
}

func (p ParsedArgs) Get(names ...string) (string, bool) {
	for _, name := range names {
		if value, ok := p.Opts[name]; ok {
			return value, true
		}
	}
	return "", false
}

func (p ParsedArgs) RequireOne(name string) (string, error) {
	if len(p.Rest) == 0 {
		return "", fmt.Errorf("missing argument: %s", name)
	}

	return p.Rest[0], nil
}
