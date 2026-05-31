package applets

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func init() {
	Register("init", "initramfs init process", Init)
}

func Init(args []string) error {
	logInit("yasld init started")

	setupDirs()
	mountCoreFilesystems()
	setupDevLinks()
	setHostname("yasldlive")

	configureDHCP()

	clearConsole()
	drawAscii()

	console, err := openConsole()
	if err != nil {
		return err
	}
	defer console.Close()

	for {
		err := runShell(console)
		if err != nil {
			logInit(fmt.Sprintf("shell exited: %v", err))
		} else {
			logInit("shell exited")
		}

		logInit("respawning shell in 1 second")
		time.Sleep(time.Second)
	}
}

func setupDirs() {
	mkdir("/", 0755)
	mkdir("/proc", 0555)
	mkdir("/sys", 0555)
	mkdir("/dev", 0755)
	mkdir("/dev/pts", 0755)
	mkdir("/tmp", 01777)
	mkdir("/root", 0700)
	mkdir("/etc", 0755)
	mkdir("/bin", 0755)
	mkdir("/sbin", 0755)
	mkdir("/usr", 0755)
	mkdir("/usr/bin", 0755)
	mkdir("/usr/sbin", 0755)
}

func mountCoreFilesystems() {
	mount("proc", "/proc", "proc", 0, "")
	mount("sysfs", "/sys", "sysfs", 0, "")
	mount("devtmpfs", "/dev", "devtmpfs", 0, "")
	mkdir("/dev/pts", 0755)
	mount("devpts", "/dev/pts", "devpts", 0, "gid=5,mode=620")
	mount("tmpfs", "/tmp", "tmpfs", 0, "mode=1777")
}

func setupDevLinks() {
	symlink("/proc/self/fd", "/dev/fd")
	symlink("/proc/self/fd/0", "/dev/stdin")
	symlink("/proc/self/fd/1", "/dev/stdout")
	symlink("/proc/self/fd/2", "/dev/stderr")
}

func runShell(console *os.File) error {
	cmd := exec.Command("/bin/gobox", "sh")

	cmd.Stdin = console
	cmd.Stdout = console
	cmd.Stderr = console

	cmd.Env = []string{
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin",
		"SHELL=/bin/sh",
		"HOME=/root",
		"TERM=linux",
		"USER=root",
		"LOGNAME=root",
	}

	return cmd.Run()
}

func openConsole() (*os.File, error) {
	console, err := os.OpenFile("/dev/console", os.O_RDWR, 0)
	if err == nil {
		return console, nil
	}

	console, err = os.OpenFile("/dev/tty0", os.O_RDWR, 0)
	if err == nil {
		return console, nil
	}

	return nil, fmt.Errorf("cannot open console or tty0")
}

func clearConsole() {
	writeConsole("\033[2J\033[H")
}

func setHostname(name string) {
	if err := syscall.Sethostname([]byte(name)); err != nil {
		logInit(fmt.Sprintf("sethostname failed: %v", err))
	}
}

func mkdir(path string, mode os.FileMode) {
	if err := os.MkdirAll(path, mode); err != nil {
		logInit(fmt.Sprintf("mkdir %s failed: %v", path, err))
	}
}

func mount(source, target, fstype string, flags uintptr, data string) {
	if err := syscall.Mount(source, target, fstype, flags, data); err != nil {
		if err == syscall.EBUSY {
			return
		}

		logInit(fmt.Sprintf("mount %s on %s failed: %v", fstype, target, err))
	}
}

func symlink(oldname, newname string) {
	if _, err := os.Lstat(newname); err == nil {
		return
	}

	if err := os.Symlink(oldname, newname); err != nil {
		logInit(fmt.Sprintf("symlink %s -> %s failed: %v", newname, oldname, err))
	}
}

func logInit(msg string) {
	writeConsole("[init] " + msg + "\n")
}

func drawAscii() {
	const ASCII string = `

  ▄▄▄          ▄▄      ▄▄▄▄▄     ▄▄▄      ▄▄▄▄▄▄
 █▀██  ██    ▄█▀▀█▄   ██▀▀▀▀█▄  ▀██▀     █▀██▀▀██
   ██  ██    ██  ██   ▀██▄  ▄▀   ██        ██   ██
   ██  ██    ██▀▀██     ▀██▄▄    ██        ██   ██
   ██  ██  ▄ ██  ██   ▄   ▀██▄   ██      ▄ ██   ██
   ▀█████▄ ▀██▀  ▀█▄█ ▀██████▀  ████████ ▀██▀███▀
   ▄   ██
   ▀████▀
	`
	fmt.Println(ASCII)
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

type consoleWriter struct{}

func (consoleWriter) Write(p []byte) (int, error) {
	writeConsole(string(p))
	return len(p), nil
}

func configureDHCP() {
	logInit("configuring network with DHCP")

	if _, err := os.Stat("/bin/gobox"); err != nil {
		logInit(fmt.Sprintf("dhcp skipped: /bin/gobox missing: %v", err))
		return
	}

	waitForNetworkInterface(3 * time.Second)

	cmd := exec.Command("/bin/gobox", "dhcp", "auto")
	cmd.Stdout = consoleWriter{}
	cmd.Stderr = consoleWriter{}
	cmd.Env = []string{
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin",
		"SHELL=/bin/sh",
		"HOME=/root",
		"TERM=linux",
		"USER=root",
		"LOGNAME=root",
	}

	if err := cmd.Run(); err != nil {
		logInit(fmt.Sprintf("dhcp failed: %v", err))
		return
	}

	logInit("dhcp configured")
}

func waitForNetworkInterface(timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		entries, err := os.ReadDir("/sys/class/net")
		if err == nil {
			for _, entry := range entries {
				if entry.Name() != "lo" {
					logInit("network interface found: " + entry.Name())
					return
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	logInit("no network interface found before timeout")
}
