package applets

import (
	"fmt"
	"os"
	"os/exec"
)

func init() {
	Register("init", "initramfs init process", Init)
}

func Init(args []string) error {
	writeConsole("[init] yasld init started\n")

	_ = os.MkdirAll("/proc", 0755)
	_ = os.MkdirAll("/sys", 0755)
	_ = os.MkdirAll("/dev", 0755)
	_ = os.MkdirAll("/tmp", 0777)

	run("mount", "-t", "proc", "proc", "/proc")
	run("mount", "-t", "sysfs", "sysfs", "/sys")
	run("mount", "-t", "devtmpfs", "devtmpfs", "/dev")

	writeConsole("[init] starting gobox shell\n")

	console, err := os.OpenFile("/dev/console", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer console.Close()

	cmd := exec.Command("/bin/gobox", "sh")
	cmd.Stdin = console
	cmd.Stdout = console
	cmd.Stderr = console
	cmd.Env = []string{
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin",
		"SHELL=/bin/gobox",
		"HOME=/",
		"TERM=linux",
	}

	if err := cmd.Run(); err != nil {
		writeConsole(fmt.Sprintf("[init] shell exited: %v\n", err))
	}

	for {
		run("sleep", "1")
	}
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func writeConsole(msg string) {
	f, err := os.OpenFile("/dev/console", os.O_WRONLY, 0)
	if err != nil {
		fmt.Print(msg)
		return
	}
	defer f.Close()

	_, _ = f.WriteString(msg)
}
