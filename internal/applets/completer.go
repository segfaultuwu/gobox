package applets

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type goboxCompleter struct{}

func shellCompleter() readline.AutoCompleter {
	return goboxCompleter{}
}

func (goboxCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	parts := splitCompletionArgs(input)

	if len(parts) <= 1 && !strings.HasSuffix(input, " ") {
		prefix := ""

		if len(parts) == 1 {
			prefix = parts[0]
		}

		return completeCommand(prefix), len(prefix)
	}

	current := ""

	if strings.HasSuffix(input, " ") {
		current = ""
	} else if len(parts) > 0 {
		current = parts[len(parts)-1]
	}

	return completePath(current), len(current)
}

func completeCommand(prefix string) [][]rune {
	seen := map[string]bool{}
	var names []string

	for name := range Registry {
		if strings.HasPrefix(name, prefix) {
			seen[name] = true
			names = append(names, name)
		}
	}

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()

			if seen[name] {
				continue
			}

			if !strings.HasPrefix(name, prefix) {
				continue
			}

			path := filepath.Join(dir, name)

			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.IsDir() {
				continue
			}

			if info.Mode()&0111 == 0 {
				continue
			}

			seen[name] = true
			names = append(names, name)
		}
	}

	sort.Strings(names)

	suggestions := make([][]rune, 0, len(names))
	for _, name := range names {
		suggestions = append(suggestions, []rune(name[len(prefix):]+" "))
	}

	return suggestions
}

func completePath(prefix string) [][]rune {
	dir := "."
	base := prefix
	displayPrefix := ""

	if prefix == "" {
		dir = "."
		base = ""
		displayPrefix = ""
	} else if strings.HasPrefix(prefix, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}

		withoutHome := strings.TrimPrefix(prefix, "~/")

		if strings.Contains(withoutHome, "/") {
			dir = filepath.Join(home, filepath.Dir(withoutHome))
			base = filepath.Base(withoutHome)
			displayPrefix = filepath.Dir(withoutHome)

			if displayPrefix == "." {
				displayPrefix = ""
			}
		} else {
			dir = home
			base = withoutHome
			displayPrefix = ""
		}
	} else if strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		base = filepath.Base(prefix)
		displayPrefix = filepath.Dir(prefix)

		if displayPrefix == "." {
			displayPrefix = ""
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string

	for _, entry := range entries {
		name := entry.Name()

		if !strings.HasPrefix(name, base) {
			continue
		}

		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	var suggestions [][]rune

	for _, name := range names {
		fullPath := filepath.Join(dir, name)

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		suffix := name[len(base):]

		if info.IsDir() {
			suffix += "/"
		} else {
			suffix += " "
		}

		_ = displayPrefix
		suggestions = append(suggestions, []rune(suffix))
	}

	return suggestions
}

func splitCompletionArgs(line string) []string {
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

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
