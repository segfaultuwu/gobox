package applets

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	vtActivate   = 0x5606
	vtWaitActive = 0x5607
)

func init() {
	Register("chvt", "change active virtual terminal", Chvt)
}

func Chvt(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: chvt <tty-number>")
	}

	n, err := strconv.Atoi(args[0])
	if err != nil || n <= 0 {
		return fmt.Errorf("invalid tty number: %s", args[0])
	}

	console, err := os.OpenFile("/dev/console", os.O_RDWR, 0)
	if err != nil {
		console, err = os.OpenFile("/dev/tty0", os.O_RDWR, 0)
		if err != nil {
			return err
		}
	}
	defer console.Close()

	if err := ioctl(console.Fd(), vtActivate, uintptr(n)); err != nil {
		return fmt.Errorf("VT_ACTIVATE %d: %w", n, err)
	}

	if err := ioctl(console.Fd(), vtWaitActive, uintptr(n)); err != nil {
		return fmt.Errorf("VT_WAITACTIVE %d: %w", n, err)
	}

	return nil
}

func ioctl(fd uintptr, req uintptr, arg uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg)
	if errno != 0 {
		return errno
	}

	_ = unsafe.Sizeof(arg)
	return nil
}
