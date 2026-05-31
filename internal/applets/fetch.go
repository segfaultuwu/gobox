package applets

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func init() {
	Register("fetch", "show system information", Fastfetch)
}

func Fastfetch(args []string) error {
	osName := readOSName()
	kernel := readKernel()
	cpu := readCPU()
	memUsed, memTotal := readMemory()
	uptime := readUptime()
	hostname := readHostname()
	shell := getenvDefault("SHELL", "/bin/sh")
	term := getenvDefault("TERM", "unknown")
	user := readUsername()
	tty := readTTY()
	initName := readInitName()
	cmdline := readKernelCmdline()
	procs := countProcesses()
	loadavg := readLoadavg()
	rootfs := readRootFS()
	diskUsed, diskTotal := readDiskUsage("/")
	goRuntime := runtime.Version()

	logo := []string{
		"\033[36m ▄· ▄▌ ▄▄▄· .▄▄ ·    \033[0m",
		"\033[36m▐█▪██▌▐█ ▀█ ▐█ ▀.    \033[0m",
		"\033[36m▐█▌▐█▪▄█▀▀█ ▄▀▀▀█▄   \033[0m",
		"\033[36m ▐█▀·.▐█ ▪▐▌▐█▄▪▐█   \033[0m",
		"\033[36m  ▀ •  ▀  ▀  ▀▀▀▀    \033[0m",
		"\033[36m    ▄▄▌  ·▄▄▄▄       \033[0m",
		"\033[36m    ██•  ██▪ ██      \033[0m",
		"\033[36m    ██▪  ▐█· ▐█▌     \033[0m",
		"\033[36m    ▐█▌▐▌██. ██      \033[0m",
		"\033[36m    .▀▀▀ ▀▀▀▀▀•      \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
		"\033[36m                     \033[0m",
	}

	info := []string{
		fmt.Sprintf("%s@%s", user, hostname),
		"-------------",
		"OS: " + osName,
		"Host: " + hostname,
		"Kernel: " + kernel,
		"Arch: " + runtime.GOARCH,
		"CPU: " + cpu,
		fmt.Sprintf("Memory: %d MiB / %d MiB", memUsed, memTotal),
		fmt.Sprintf("Disk (/): %d MiB / %d MiB", diskUsed, diskTotal),
		"RootFS: " + rootfs,
		"Uptime: " + uptime,
		"Shell: " + shell,
		"Terminal: " + term,
		"TTY: " + tty,
		"Init: " + initName,
		"Processes: " + procs,
		"Loadavg: " + loadavg,
		"Go: " + goRuntime,
		"Cmdline: " + cmdline,
		"Colors: \033[30m●\033[31m ●\033[32m ●\033[33m ●\033[34m ●\033[35m ●\033[36m ●\033[37m ●\033[0m",
	}

	max := len(info)
	if len(logo) > max {
		max = len(logo)
	}

	for i := 0; i < max; i++ {
		left := "\033[36m                          \033[0m"
		right := ""

		if i < len(logo) {
			left = logo[i]
		}

		if i < len(info) {
			right = info[i]
		}

		fmt.Println(left + right)
	}

	return nil
}

func readOSName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "yasld Linux"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return cleanValue(strings.TrimPrefix(line, "PRETTY_NAME="))
		}
		if strings.HasPrefix(line, "NAME=") {
			return cleanValue(strings.TrimPrefix(line, "NAME="))
		}
	}

	return "yasld Linux"
}

func readKernel() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return "unknown"
	}

	return charsToString(uts.Release[:])
}

func readCPU() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return runtime.GOARCH
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return runtime.GOARCH
}

func readMemory() (usedMiB int, totalMiB int) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	values := map[string]int{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		values[strings.TrimSuffix(fields[0], ":")] = value
	}

	total := values["MemTotal"]
	available := values["MemAvailable"]

	if total == 0 {
		return 0, 0
	}

	if available == 0 {
		available = values["MemFree"]
	}

	used := total - available

	return used / 1024, total / 1024
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "unknown"
	}

	d := time.Duration(seconds) * time.Second

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}

	return fmt.Sprintf("%ds", secs)
}

func readHostname() string {
	data, err := os.ReadFile("/proc/sys/kernel/hostname")
	if err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" {
			return name
		}
	}

	name, err := os.Hostname()
	if err == nil && name != "" {
		return name
	}

	return "yasld"
}

func readUsername() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}

	if logname := os.Getenv("LOGNAME"); logname != "" {
		return logname
	}

	if os.Getuid() == 0 {
		return "root"
	}

	return strconv.Itoa(os.Getuid())
}

func readTTY() string {
	link, err := os.Readlink("/proc/self/fd/0")
	if err != nil {
		return "unknown"
	}

	if strings.HasPrefix(link, "/dev/") {
		return strings.TrimPrefix(link, "/dev/")
	}

	return link
}

func readInitName() string {
	data, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return "unknown"
	}

	name := strings.TrimSpace(string(data))
	if name == "" {
		return "unknown"
	}

	return name
}

func readKernelCmdline() string {
	data, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return "unknown"
	}

	cmdline := strings.TrimSpace(string(data))
	if cmdline == "" {
		return "unknown"
	}

	if len(cmdline) > 80 {
		return cmdline[:77] + "..."
	}

	return cmdline
}

func countProcesses() string {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return "unknown"
	}

	count := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if isNumeric(name) {
			count++
		}
	}

	return strconv.Itoa(count)
}

func readLoadavg() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return "unknown"
	}

	return fields[0] + " " + fields[1] + " " + fields[2]
}

func readRootFS() string {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "/" {
			return fields[2]
		}
	}

	return "unknown"
}

func readDiskUsage(path string) (usedMiB int, totalMiB int) {
	var stat syscall.Statfs_t

	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return int(used / 1024 / 1024), int(total / 1024 / 1024)
}

func getenvDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func cleanValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	return value
}

func charsToString(chars []int8) string {
	var b strings.Builder

	for _, c := range chars {
		if c == 0 {
			break
		}
		b.WriteByte(byte(c))
	}

	return b.String()
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
