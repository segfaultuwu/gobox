package applets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gobox/internal/utils"

	"github.com/chzyer/readline"
)

var defaultEnv = map[string]string{
	"PATH": "/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin",
}

func init() {
	initDefaultEnv()
	Register("sh", "simple gobox shell", Sh)
}

func initDefaultEnv() {
	for key, value := range defaultEnv {
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	if os.Getenv("SHELL") == "" {
		_ = os.Setenv("SHELL", "gobox")
	}
}

func Sh(args []string) error {
	if len(args) > 0 {
		return runShellLine(strings.Join(args, " "))
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          buildPrompt(),
		HistoryFile:     historyPath(),
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
			if strings.TrimSpace(line) == "" {
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

func buildPrompt() string {
	suffix := "$"
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

	return fmt.Sprintf("[ \033[33m%s\033[0m ] %s ", pwd, suffix)
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/gobox_history"
	}

	return filepath.Join(home, ".gobox_history")
}

func runShellLine(line string) error {
	args, err := utils.SplitArgs(line)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return nil
	}

	cmd := args[0]
	cmdArgs := args[1:]

	if err := RunApplet(cmd, cmdArgs); err == nil {
		return nil
	}

	return runExternal(cmd, cmdArgs)
}

func runExternal(cmd string, cmdArgs []string) error {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmd)
	}

	c := exec.Command(path, cmdArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()

	return c.Run()
}
