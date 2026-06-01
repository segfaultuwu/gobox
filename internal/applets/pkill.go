package applets

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	Register("pkill", "kill processes by name", Pkill)
}

func Pkill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pkill [-SIGNAL] [-f] <pattern>")
	}

	sig := syscall.SIGTERM
	matchCmdline := false

	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "-f" {
			matchCmdline = true
			i++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			parsedSig, err := parsePkillSignal(arg)
			if err != nil {
				return err
			}

			sig = parsedSig
			i++
			continue
		}

		break
	}

	if i >= len(args) {
		return fmt.Errorf("missing pattern")
	}

	pattern := args[i]
	if pattern == "" {
		return fmt.Errorf("empty pattern")
	}

	self := os.Getpid()

	matches, err := findMatchingPIDs(pattern, matchCmdline, self)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no process matched: %s", pattern)
	}

	var failed int

	for _, pid := range matches {
		if err := syscall.Kill(pid, sig); err != nil {
			fmt.Fprintf(os.Stderr, "pkill: kill %d: %v\n", pid, err)
			failed++
			continue
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d process(es) failed", failed)
	}

	return nil
}

func findMatchingPIDs(pattern string, matchCmdline bool, self int) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var matches []int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		if pid <= 0 || pid == self {
			continue
		}

		text := ""

		if matchCmdline {
			text = readProcCmdline(pid)
		} else {
			text = readProcComm(pid)
		}

		if text == "" {
			continue
		}

		if strings.Contains(text, pattern) {
			matches = append(matches, pid)
		}
	}

	return matches, nil
}

func readProcComm(pid int) string {
	path := filepath.Join("/proc", strconv.Itoa(pid), "comm")

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func readProcCmdline(pid int) string {
	path := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	text := strings.ReplaceAll(string(data), "\x00", " ")
	return strings.TrimSpace(text)
}

func parsePkillSignal(text string) (syscall.Signal, error) {
	text = strings.TrimPrefix(text, "-")
	text = strings.ToUpper(text)

	switch text {
	case "1", "HUP", "SIGHUP":
		return syscall.SIGHUP, nil
	case "2", "INT", "SIGINT":
		return syscall.SIGINT, nil
	case "3", "QUIT", "SIGQUIT":
		return syscall.SIGQUIT, nil
	case "9", "KILL", "SIGKILL":
		return syscall.SIGKILL, nil
	case "15", "TERM", "SIGTERM":
		return syscall.SIGTERM, nil
	case "18", "CONT", "SIGCONT":
		return syscall.SIGCONT, nil
	case "19", "STOP", "SIGSTOP":
		return syscall.SIGSTOP, nil
	default:
		return 0, fmt.Errorf("unknown signal: %s", text)
	}
}
