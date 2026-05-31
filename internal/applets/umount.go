package applets

import (
	"fmt"
	"syscall"
)

func init() {
	Register("umount", "unmount filesystems", Umount)
	Register("unmount", "unmount filesystems", Umount)
}

func Umount(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umount [-f] [-l] target")
	}

	var flags int
	var targets []string

	for _, arg := range args {
		switch arg {
		case "-f", "--force":
			flags |= syscall.MNT_FORCE

		case "-l", "--lazy":
			flags |= syscall.MNT_DETACH

		case "-r", "--read-only":
			// Linux util-linux umount has -r for "remount read-only on failure".
			// Skip cause os.Umount does not do this

		default:
			targets = append(targets, arg)
		}
	}

	if len(targets) == 0 {
		return fmt.Errorf("missing target")
	}

	for _, target := range targets {
		if err := syscall.Unmount(target, flags); err != nil {
			return fmt.Errorf("umount %s: %w", target, err)
		}
	}

	return nil
}
