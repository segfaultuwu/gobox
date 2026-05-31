package applets

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func init() {
	Register("useradd", "add a user", Useradd)
}

func Useradd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: useradd [-m] [-s shell] [-u uid] [-g gid] name")
	}

	createHome := false
	shell := "/bin/sh"
	uid := -1
	gid := -1

	var name string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			createHome = true

		case "-s":
			if i+1 >= len(args) {
				return fmt.Errorf("missing shell after -s")
			}
			shell = args[i+1]
			i++

		case "-u":
			if i+1 >= len(args) {
				return fmt.Errorf("missing uid after -u")
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid uid: %s", args[i+1])
			}
			uid = v
			i++

		case "-g":
			if i+1 >= len(args) {
				return fmt.Errorf("missing gid after -g")
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid gid: %s", args[i+1])
			}
			gid = v
			i++

		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown option: %s", args[i])
			}
			name = args[i]
		}
	}

	if name == "" {
		return fmt.Errorf("missing username")
	}

	if strings.ContainsAny(name, ":\n/") {
		return fmt.Errorf("invalid username")
	}

	users, err := readPasswd()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == name {
			return fmt.Errorf("user already exists: %s", name)
		}
	}

	if uid < 0 {
		uid = nextUID(users)
	}

	if gid < 0 {
		gid = uid
	}

	home := "/home/" + name

	if err := ensureGroup(name, gid); err != nil {
		return err
	}

	passwdLine := fmt.Sprintf("%s:x:%d:%d:%s:%s:%s\n", name, uid, gid, name, home, shell)
	if err := appendFile("/etc/passwd", passwdLine, 0644); err != nil {
		return err
	}

	shadowLine := fmt.Sprintf("%s:*:0:0:99999:7:::\n", name)
	if err := appendFile("/etc/shadow", shadowLine, 0600); err != nil {
		return err
	}

	if createHome {
		if err := os.MkdirAll(home, 0755); err != nil {
			return err
		}
		_ = os.Chown(home, uid, gid)
	}

	return nil
}

type passwdUser struct {
	Name string
	UID  int
	GID  int
}

func readPasswd() ([]passwdUser, error) {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return nil, err
	}

	var users []passwdUser

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])

		users = append(users, passwdUser{
			Name: parts[0],
			UID:  uid,
			GID:  gid,
		})
	}

	return users, nil
}

func nextUID(users []passwdUser) int {
	maxUID := 999

	for _, user := range users {
		if user.UID >= maxUID {
			maxUID = user.UID
		}
	}

	return maxUID + 1
}

func ensureGroup(name string, gid int) error {
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[0] == name {
			return nil
		}
	}

	line := fmt.Sprintf("%s:x:%d:\n", name, gid)
	return appendFile("/etc/group", line, 0644)
}

func appendFile(path string, text string, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}
