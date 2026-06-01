package applets

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	Register("kill", "send signal to process", Kill)
}

func Kill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kill [-SIGNAL] <pid> [pid...]")
	}

	sig := syscall.SIGTERM
	pids := args

	if strings.HasPrefix(args[0], "-") {
		parsedSig, err := parseSignal(args[0])
		if err != nil {
			return err
		}

		sig = parsedSig
		pids = args[1:]

		if len(pids) == 0 {
			return fmt.Errorf("missing pid")
		}
	}

	for _, pidText := range pids {
		pid, err := strconv.Atoi(pidText)
		if err != nil {
			return fmt.Errorf("invalid pid: %s", pidText)
		}

		if pid <= 0 {
			return fmt.Errorf("invalid pid: %d", pid)
		}

		if err := syscall.Kill(pid, sig); err != nil {
			return fmt.Errorf("kill %d: %w", pid, err)
		}
	}

	return nil
}

func parseSignal(text string) (syscall.Signal, error) {
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
