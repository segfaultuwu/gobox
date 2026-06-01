package applets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	Register("Yaslctl", "control yasld services", Yaslctl)
}

func Yaslctl(args []string) error {
	if len(args) == 0 {
		printYaslctlHelp()
		return nil
	}

	switch args[0] {
	case "status":
		return yasldStatus()

	case "logs":
		return catFile("/run/init.log")

	case "net":
		return yasldNet(args[1:])

	case "ssh":
		return yasldSSH(args[1:])

	case "reboot":
		return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)

	case "poweroff":
		return syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	case "help", "-h", "--help":
		printYaslctlHelp()
		return nil

	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printYaslctlHelp() {
	fmt.Println("yaslctl - Yet Another Shitty Linux Control TooL")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  yaslctl status")
	fmt.Println("  yaslctl logs")
	fmt.Println("  yaslctl < service > < status | start | stop >")
	fmt.Println("  yaslctl reboot")
	fmt.Println("  yaslctl poweroff")
}

func yasldStatus() error {
	fmt.Println(" YASLD status")
	fmt.Println("--------------")

	hostname := readTrim("/proc/sys/kernel/hostname")
	if hostname == "" {
		hostname = "unknown"
	}

	fmt.Println("hostname:", hostname)

	if iface := firstNetworkInterface(); iface != "" {
		fmt.Println("net:      interface", iface)
	} else {
		fmt.Println("net:      no interface")
	}

	if hasDefaultRoute() {
		fmt.Println("route:    default route present")
	} else {
		fmt.Println("route:    no default route")
	}

	if pid := findPidByName("dropbear"); pid > 0 {
		fmt.Println("ssh:      running pid", pid)
	} else {
		fmt.Println("ssh:      stopped")
	}

	if _, err := os.Stat("/usr/bin/bash"); err == nil {
		fmt.Println("bash:     installed")
	} else {
		fmt.Println("bash:     missing")
	}

	if _, err := os.Stat("/usr/bin/git"); err == nil {
		fmt.Println("git:      installed")
	} else {
		fmt.Println("git:      missing")
	}

	if _, err := os.Stat("/usr/bin/curl"); err == nil {
		fmt.Println("curl:     installed")
	} else {
		fmt.Println("curl:     missing")
	}

	return nil
}

func yasldNet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: Yaslctl net <status|start>")
	}

	switch args[0] {
	case "status":
		if iface := firstNetworkInterface(); iface != "" {
			fmt.Println("interface:", iface)
		} else {
			fmt.Println("interface: none")
		}

		if hasDefaultRoute() {
			fmt.Println("default route: yes")
		} else {
			fmt.Println("default route: no")
		}

		fmt.Println()
		fmt.Println("/etc/resolv.conf:")
		return catFile("/etc/resolv.conf")

	case "start":
		return runCommand("dhcp", "auto")

	default:
		return fmt.Errorf("unknown net command: %s", args[0])
	}
}

func yasldSSH(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: Yaslctl ssh <status|start|stop|restart>")
	}

	switch args[0] {
	case "status":
		pid := findPidByName("dropbear")
		if pid > 0 {
			fmt.Println("dropbear running:", pid)
		} else {
			fmt.Println("dropbear stopped")
		}
		return nil

	case "start":
		if pid := findPidByName("dropbear"); pid > 0 {
			fmt.Println("dropbear already running:", pid)
			return nil
		}

		if err := os.MkdirAll("/etc/dropbear", 0700); err != nil {
			return err
		}

		cmd := exec.Command("/usr/sbin/dropbear", "-R", "-E", "-s", "-p", "22")
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			return err
		}

		fmt.Println("dropbear started:", cmd.Process.Pid)
		return nil

	case "stop":
		pid := findPidByName("dropbear")
		if pid <= 0 {
			fmt.Println("dropbear already stopped")
			return nil
		}

		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			return err
		}

		fmt.Println("dropbear stopped:", pid)
		return nil

	case "restart":
		_ = yasldSSH([]string{"stop"})
		return yasldSSH([]string{"start"})

	default:
		return fmt.Errorf("unknown ssh command: %s", args[0])
	}
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func catFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

func readTrim(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func firstNetworkInterface() string {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "lo" {
			continue
		}

		return name
	}

	return ""
}

func hasDefaultRoute() bool {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[1] == "00000000" {
			return true
		}
	}

	return false
}

func findPidByName(name string) int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return -1
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commPath := filepath.Join("/proc", entry.Name(), "comm")
		comm := readTrim(commPath)

		if comm == name {
			return pid
		}
	}

	return -1
}
