package applets

import (
	"gobox/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

type goboxCompleter struct{}

func shellCompleter() readline.AutoCompleter {
	return goboxCompleter{}
}

func (goboxCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])
	parts := utils.SplitArgs(input)

	// Completion pierwszego słowa: komendy.
	if len(parts) <= 1 && !strings.HasSuffix(input, " ") {
		prefix := ""

		if len(parts) == 1 {
			prefix = parts[0]
		}

		var suggestions [][]rune

		for name := range Registry {
			if strings.HasPrefix(name, prefix) {
				suggestions = append(suggestions, []rune(name[len(prefix):]))
			}
		}

		for _, builtin := range []string{"cd", "exit", "help"} {
			if strings.HasPrefix(builtin, prefix) {
				suggestions = append(suggestions, []rune(builtin[len(prefix):]))
			}
		}

		return suggestions, len(prefix)
	}

	// Completion argumentów: pliki/foldery.
	current := ""

	if strings.HasSuffix(input, " ") {
		current = ""
	} else if len(parts) > 0 {
		current = parts[len(parts)-1]
	}

	suggestions := completePath(current)

	return suggestions, len(current)
}

func completePath(prefix string) [][]rune {
	dir := "."
	base := prefix

	if strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		base = filepath.Base(prefix)
	}

	if strings.HasPrefix(prefix, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dir = filepath.Join(home, strings.TrimPrefix(filepath.Dir(prefix), "~"))
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var suggestions [][]rune

	for _, entry := range entries {
		name := entry.Name()

		if !strings.HasPrefix(name, base) {
			continue
		}

		suffix := name[len(base):]

		if entry.IsDir() {
			suffix += "/"
		} else {
			suffix += " "
		}

		suggestions = append(suggestions, []rune(suffix))
	}

	return suggestions
}
