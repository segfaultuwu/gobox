package applets

import (
	"fmt"
	"gobox/internal/utils"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

func init() {
	Register("sh", "simple gobox shell", Sh)
}

func buildPrompt() string {
	var suffix string = "$"
	if utils.IsRoot() {
		suffix = "#"
	}
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "?"
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(pwd, home) {
		pwd = "~" + strings.TrimPrefix(pwd, home)
	}
	return fmt.Sprintf("[ \033[33m%s \033[0m] %s ", pwd, suffix)
}

func Sh(args []string) error {
	if len(args) > 0 {
		return runShellLine(strings.Join(args, " "))
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          buildPrompt(),
		HistoryFile:     "/tmp/gobox_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    shellCompleter(),
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		rl.SetPrompt(buildPrompt())

		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				continue
			}
		}

		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := runShellLine(line); err != nil {
			fmt.Fprintln(os.Stderr, "sh:", err)
		}
	}

	return nil
}

func runShellLine(line string) error {
	args := utils.SplitArgs(line)
	if len(args) == 0 {
		return nil
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "exit":
		os.Exit(0)

	case "cd":
		return shellCd(cmdArgs)

	case "help":
		shellHelp()
		return nil
	}

	return RunApplet(cmd, cmdArgs)
}

func shellCd(args []string) error {
	var dir string

	if len(args) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		dir = home
	} else {
		dir = args[0]
	}

	return os.Chdir(dir)
}

func shellHelp() {
	fmt.Println("gobox shell")
	fmt.Println()
	fmt.Println("Builtins:")
	fmt.Println("  cd [dir]")
	fmt.Println("  exit")
	fmt.Println("  help")
	fmt.Println()
	fmt.Println("Applets:")

	for name := range Registry {
		fmt.Println("  " + name)
	}
}
