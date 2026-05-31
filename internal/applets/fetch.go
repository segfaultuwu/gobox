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

	"gobox/internal/utils"
)

const (
	colorReset   = "\033[0m"
	colorBlack   = "\033[30m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
)

func init() {
	Register("fetch", "show system information", Fastfetch)
	Register("fastfetch", "show system information", Fastfetch)
}

type fetchOptions struct {
	showLogo bool
	plain    bool
	color    string
}

type systemInfo struct {
	User       string
	Hostname   string
	OS         string
	Kernel     string
	Machine    string
	Arch       string
	CPU        string
	Cores      string
	Memory     string
	Disk       string
	RootFS     string
	RootSource string
	Uptime     string
	Shell      string
	Terminal   string
	TTY        string
	Init       string
	Processes  string
	Loadavg    string
	GoRuntime  string
	Cmdline    string
}

func Fastfetch(args []string) error {
	opts, err := parseFetchOptions(args)
	if err != nil {
		return err
	}

	if opts.plain {
		opts.showLogo = false
	}

	info := collectSystemInfo()
	printFetch(info, opts)

	return nil
}

func parseFetchOptions(args []string) (fetchOptions, error) {
	opts := fetchOptions{
		showLogo: true,
		color:    colorCyan,
	}

	parsed := utils.ParseArgs(args)

	if parsed.Has("h", "help") {
		return opts, fmt.Errorf("usage: fetch [--no-logo] [--plain] [--color cyan|green|blue|magenta|yellow|red|white]")
	}

	if parsed.Has("no-logo") {
		opts.showLogo = false
	}

	if parsed.Has("plain") {
		opts.plain = true
		opts.showLogo = false
		opts.color = ""
	}

	if value, ok := parsed.Get("color"); ok {
		opts.color = colorByName(value)
		if opts.color == "" && !opts.plain {
			return opts, fmt.Errorf("unknown color: %s", value)
		}
	}

	return opts, nil
}

func collectSystemInfo() systemInfo {
	rootSource, rootfs := readRootMount()
	memUsed, memTotal := readMemory()
	diskUsed, diskTotal := readDiskUsage("/")

	return systemInfo{
		User:       readUsername(),
		Hostname:   readHostname(),
		OS:         readOSName(),
		Kernel:     readKernelRelease(),
		Machine:    readKernelMachine(),
		Arch:       runtime.GOARCH,
		CPU:        readCPU(),
		Cores:      readCPUCount(),
		Memory:     formatUsage(memUsed, memTotal),
		Disk:       formatUsage(diskUsed, diskTotal),
		RootFS:     rootfs,
		RootSource: rootSource,
		Uptime:     readUptime(),
		Shell:      getenvDefault("SHELL", "/bin/sh"),
		Terminal:   getenvDefault("TERM", "unknown"),
		TTY:        readTTY(),
		Init:       readInitName(),
		Processes:  countProcesses(),
		Loadavg:    readLoadavg(),
		GoRuntime:  runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH,
		Cmdline:    readKernelCmdline(),
	}
}

func printFetch(info systemInfo, opts fetchOptions) {
	logo := []string{
		" ▄· ▄▌ ▄▄▄· .▄▄ ·    ",
		"▐█▪██▌▐█ ▀█ ▐█ ▀.    ",
		"▐█▌▐█▪▄█▀▀█ ▄▀▀▀█▄   ",
		" ▐█▀·.▐█ ▪▐▌▐█▄▪▐█   ",
		"  ▀ •  ▀  ▀  ▀▀▀▀    ",
		"    ▄▄▌  ·▄▄▄▄       ",
		"    ██•  ██▪ ██      ",
		"    ██▪  ▐█· ▐█▌     ",
		"    ▐█▌▐▌██. ██      ",
		"    .▀▀▀ ▀▀▀▀▀•      ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
		"                     ",
	}

	infoLines := []string{
		fmt.Sprintf("%s@%s", info.User, info.Hostname),
		"-------------",
		"OS: " + info.OS,
		"Host: " + info.Hostname,
		"Kernel: " + info.Kernel,
		"Machine: " + info.Machine,
		"Arch: " + info.Arch,
		"CPU: " + info.CPU,
		"CPU Cores: " + info.Cores,
		"Memory: " + info.Memory,
		"Disk (/): " + info.Disk,
		"RootFS: " + info.RootFS,
		"Root Source: " + info.RootSource,
		"Uptime: " + info.Uptime,
		"Shell: " + info.Shell,
		"Terminal: " + info.Terminal,
		"TTY: " + info.TTY,
		"Init: " + info.Init,
		"Processes: " + info.Processes,
		"Loadavg: " + info.Loadavg,
		"Go: " + info.GoRuntime,
		"Cmdline: " + info.Cmdline,
		"Colors: " + colorBlocks(opts.plain),
	}

	if !opts.showLogo {
		for _, line := range infoLines {
			fmt.Println(utils.StripANSI(line))
		}
		return
	}

	max := len(infoLines)
	if len(logo) > max {
		max = len(logo)
	}

	for i := 0; i < max; i++ {
		left := "                     "
		right := ""

		if i < len(logo) {
			left = logo[i]
		}

		if i < len(infoLines) {
			right = infoLines[i]
		}

		if opts.plain {
			fmt.Println(left + right)
		} else {
			fmt.Println(opts.color + left + colorReset + right)
		}
	}
}

func readOSName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return cleanValue(strings.TrimPrefix(line, "PRETTY_NAME="))
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			return cleanValue(strings.TrimPrefix(line, "NAME="))
		}
	}

	return "yasld Linux"
}

func readKernelRelease() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return "unknown"
	}

	return utils.CharsToString(uts.Release[:])
}

func readKernelMachine() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return "unknown"
	}

	machine := utils.CharsToString(uts.Machine[:])
	if machine == "" {
		return runtime.GOARCH
	}

	return machine
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

		if strings.HasPrefix(line, "Hardware") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}

		if strings.HasPrefix(line, "Processor") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return runtime.GOARCH
}

func readCPUCount() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return strconv.Itoa(runtime.NumCPU())
	}
	defer file.Close()

	count := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}

	if count == 0 {
		count = runtime.NumCPU()
	}

	return strconv.Itoa(count)
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
	if err == nil && link != "" {
		if strings.HasPrefix(link, "/dev/") {
			return strings.TrimPrefix(link, "/dev/")
		}
		return link
	}

	stat, err := os.Stdin.Stat()
	if err == nil && stat.Mode()&os.ModeCharDevice != 0 {
		return "console"
	}

	return "unknown"
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

	return utils.Truncate(cmdline, 90)
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

		if utils.IsNumeric(entry.Name()) {
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

func readRootMount() (source string, fstype string) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return "unknown", "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "/" {
			return fields[0], fields[2]
		}
	}

	return "unknown", "unknown"
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

func formatUsage(usedMiB int, totalMiB int) string {
	if totalMiB <= 0 {
		return "unknown"
	}

	percent := float64(usedMiB) / float64(totalMiB) * 100

	return fmt.Sprintf("%d MiB / %d MiB (%.1f%%)", usedMiB, totalMiB, percent)
}

func colorBlocks(plain bool) string {
	if plain {
		return "black red green yellow blue magenta cyan white"
	}

	return colorBlack + "●" +
		colorRed + " ●" +
		colorGreen + " ●" +
		colorYellow + " ●" +
		colorBlue + " ●" +
		colorMagenta + " ●" +
		colorCyan + " ●" +
		colorWhite + " ●" +
		colorReset
}

func colorByName(name string) string {
	switch strings.ToLower(name) {
	case "black":
		return colorBlack
	case "red":
		return colorRed
	case "green":
		return colorGreen
	case "yellow":
		return colorYellow
	case "blue":
		return colorBlue
	case "magenta", "purple":
		return colorMagenta
	case "cyan":
		return colorCyan
	case "white":
		return colorWhite
	default:
		return ""
	}
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
