package applets

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type passwdEntry struct {
	Name  string
	UID   string
	GID   string
	Home  string
	Shell string
}

func init() {
	Register("login", "simple login program", Login)
}

func Login(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s login: ", readHostname())

	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("empty username")
	}

	user, err := findUser(username)
	if err != nil {
		return err
	}

	if user.Name != "root" {
		return fmt.Errorf("only root login is supported for now")
	}

	if user.Home == "" {
		user.Home = "/root"
	}

	if user.Shell == "" {
		user.Shell = "/bin/bash"
	}

	_ = os.Chdir(user.Home)

	env := []string{
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin",
		"HOME=" + user.Home,
		"SHELL=" + user.Shell,
		"USER=" + user.Name,
		"LOGNAME=" + user.Name,
		"TERM=console",
	}

	shellName := "-" + baseName(user.Shell)

	cmd := exec.Command(user.Shell)
	cmd.Args = []string{shellName}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	return cmd.Run()
}

func findUser(name string) (passwdEntry, error) {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return passwdEntry{}, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		if parts[0] != name {
			continue
		}

		return passwdEntry{
			Name:  parts[0],
			UID:   parts[2],
			GID:   parts[3],
			Home:  parts[5],
			Shell: parts[6],
		}, nil
	}

	return passwdEntry{}, fmt.Errorf("unknown user: %s", name)
}

func baseName(path string) string {
	path = strings.TrimRight(path, "/")

	index := strings.LastIndex(path, "/")
	if index == -1 {
		return path
	}

	return path[index+1:]
}

func loginSetControllingTTY(fd uintptr) {
	_, _ = syscall.Setsid()
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(0x540E), 0) // TIOCSCTTY
}
