package applets

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"gobox/internal/utils"
)

const (
	syslogActionReadAll       = 3
	syslogActionReadClear     = 4
	syslogActionSizeBuffer    = 10
	defaultKernelLogBufferCap = 1 << 20
)

func init() {
	Register("dmesg", "print kernel ring buffer", Dmesg)
}

func Dmesg(args []string) error {
	pArgs := utils.ParseArgs(args)

	clear := pArgs.Has("c", "clear")
	raw := pArgs.Has("r", "raw")
	noColor := pArgs.Has("no-color")
	levels := pArgs.Has("l", "level")

	size, err := kernelLogSize()
	if err != nil || size <= 0 {
		size = defaultKernelLogBufferCap
	}

	buffer := make([]byte, size+1)

	action := syslogActionReadAll
	if clear {
		action = syslogActionReadClear
	}

	n, err := syslog(action, buffer)
	if err != nil {
		return err
	}

	output := string(buffer[:n])

	if raw {
		fmt.Print(output)
		if output != "" && !strings.HasSuffix(output, "\n") {
			fmt.Println()
		}
		return nil
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		level, text := parseDmesgLine(line)

		if levels {
			text = fmt.Sprintf("[%s] %s", levelName(level), text)
		}

		if !noColor {
			text = colorForLevel(level) + text + colorReset
		}

		fmt.Println(text)
	}

	return nil
}

func kernelLogSize() (int, error) {
	n, err := syslog(syslogActionSizeBuffer, nil)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func syslog(action int, buffer []byte) (int, error) {
	var ptr uintptr
	var length uintptr

	if len(buffer) > 0 {
		ptr = uintptr(unsafe.Pointer(&buffer[0]))
		length = uintptr(len(buffer))
	}

	r1, _, errno := syscall.Syscall(
		syscall.SYS_SYSLOG,
		uintptr(action),
		ptr,
		length,
	)

	if errno != 0 {
		return 0, errno
	}

	return int(r1), nil
}

func parseDmesgLine(line string) (int, string) {
	level := 6

	if strings.HasPrefix(line, "<") {
		end := strings.Index(line, ">")
		if end > 1 {
			rawLevel := line[1:end]
			if len(rawLevel) > 0 && rawLevel[0] >= '0' && rawLevel[0] <= '7' {
				level = int(rawLevel[0] - '0')
			}
			line = line[end+1:]
		}
	}

	line = cleanKmsgLine(line)

	return level, line
}

func cleanKmsgLine(line string) string {
	index := strings.Index(line, ";")
	if index == -1 {
		return line
	}

	return line[index+1:]
}

func colorForLevel(level int) string {
	switch level {
	case 0, 1:
		return colorBold + colorRed
	case 2, 3:
		return colorRed
	case 4:
		return colorYellow
	case 5:
		return colorCyan
	case 6:
		return colorWhite
	case 7:
		return colorGray
	default:
		return colorWhite
	}
}

func levelName(level int) string {
	switch level {
	case 0:
		return "emerg"
	case 1:
		return "alert"
	case 2:
		return "crit"
	case 3:
		return "err"
	case 4:
		return "warn"
	case 5:
		return "notice"
	case 6:
		return "info"
	case 7:
		return "debug"
	default:
		return "unknown"
	}
}
