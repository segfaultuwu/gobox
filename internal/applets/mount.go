package applets

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func init() {
	Register("mount", "mount filesystems", Mount)
}

type mountConfig struct {
	fstype   string
	opts     string
	readOnly bool
	source   string
	target   string
}

func Mount(args []string) error {
	if len(args) == 0 {
		return printMounts()
	}

	cfg, err := parseMountArgs(args)
	if err != nil {
		return err
	}

	flags, data := parseMountOptions(cfg.opts)

	if cfg.readOnly {
		flags |= syscall.MS_RDONLY
	}

	if cfg.fstype == "" {
		cfg.fstype = "auto"
	}

	if err := os.MkdirAll(cfg.target, 0755); err != nil {
		return err
	}

	if err := syscall.Mount(cfg.source, cfg.target, cfg.fstype, uintptr(flags), data); err != nil {
		return fmt.Errorf("mount %s on %s: %w", cfg.source, cfg.target, err)
	}

	return nil
}

func parseMountArgs(args []string) (mountConfig, error) {
	var cfg mountConfig
	var rest []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-t":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing filesystem type after -t")
			}

			cfg.fstype = args[i+1]
			i++

		case "-o":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing options after -o")
			}

			cfg.opts = args[i+1]
			i++

		case "-r", "--read-only":
			cfg.readOnly = true

		case "--":
			rest = append(rest, args[i+1:]...)
			i = len(args)

		default:
			if strings.HasPrefix(arg, "-t=") {
				cfg.fstype = strings.TrimPrefix(arg, "-t=")
				continue
			}

			if strings.HasPrefix(arg, "--type=") {
				cfg.fstype = strings.TrimPrefix(arg, "--type=")
				continue
			}

			if strings.HasPrefix(arg, "-o=") {
				cfg.opts = strings.TrimPrefix(arg, "-o=")
				continue
			}

			if strings.HasPrefix(arg, "--options=") {
				cfg.opts = strings.TrimPrefix(arg, "--options=")
				continue
			}

			rest = append(rest, arg)
		}
	}

	if len(rest) != 2 {
		return cfg, fmt.Errorf("usage: mount [-r] [-t type] [-o opts] source target")
	}

	cfg.source = rest[0]
	cfg.target = rest[1]

	return cfg, nil
}

func parseMountOptions(opts string) (int, string) {
	if opts == "" {
		return 0, ""
	}

	var flags int
	var data []string

	for _, opt := range strings.Split(opts, ",") {
		opt = strings.TrimSpace(opt)
		if opt == "" {
			continue
		}

		switch opt {
		case "defaults", "rw":
			// Ignore.

		case "ro":
			flags |= syscall.MS_RDONLY

		case "nosuid":
			flags |= syscall.MS_NOSUID

		case "nodev":
			flags |= syscall.MS_NODEV

		case "noexec":
			flags |= syscall.MS_NOEXEC

		case "sync":
			flags |= syscall.MS_SYNCHRONOUS

		case "remount":
			flags |= syscall.MS_REMOUNT

		case "bind":
			flags |= syscall.MS_BIND

		case "rbind":
			flags |= syscall.MS_BIND | syscall.MS_REC

		case "dirsync":
			flags |= syscall.MS_DIRSYNC

		case "mand":
			flags |= syscall.MS_MANDLOCK

		case "noatime":
			flags |= syscall.MS_NOATIME

		case "nodiratime":
			flags |= syscall.MS_NODIRATIME

		default:
			data = append(data, opt)
		}
	}

	return flags, strings.Join(data, ",")
}

func printMounts() error {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return fmt.Errorf("cannot read /proc/mounts: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}

		source := fields[0]
		target := fields[1]
		fstype := fields[2]
		opts := fields[3]

		fmt.Printf("%s on %s type %s (%s)\n", source, target, fstype, opts)
	}

	return scanner.Err()
}
