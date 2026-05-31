package applets

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func init() {
	Register("init", "initramfs init process", Init)
}

func Init(args []string) error {
	writeConsole("[init] yasld init started\n")

	mkdir("/proc", 0555)
	mkdir("/sys", 0555)
	mkdir("/dev", 0755)
	mkdir("/tmp", 0777)
	mkdir("/root", 0700)

	mount("proc", "/proc", "proc", 0, "")
	mount("sysfs", "/sys", "sysfs", 0, "")
	mount("devtmpfs", "/dev", "devtmpfs", 0, "")

	writeConsole("[init] mounted proc/sys/dev\n")
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
		"SHELL=/bin/sh",
		"HOME=/root",
		"TERM=linux",
	}

	if err := cmd.Run(); err != nil {
		writeConsole(fmt.Sprintf("[init] shell exited: %v\n", err))
	}

	for {
		_ = syscall.Pause()
	}
}

func mkdir(path string, mode os.FileMode) {
	if err := os.MkdirAll(path, mode); err != nil {
		writeConsole(fmt.Sprintf("[init] mkdir %s failed: %v\n", path, err))
	}
}

func mount(source, target, fstype string, flags uintptr, data string) {
	if err := syscall.Mount(source, target, fstype, flags, data); err != nil {
		writeConsole(fmt.Sprintf("[init] mount %s on %s failed: %v\n", fstype, target, err))
	}
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
