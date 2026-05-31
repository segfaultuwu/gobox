package applets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func init() {
	Register("ps", "show processes", Ps)
}

type ProcessInfo struct {
	PID     int
	PPID    int
	State   string
	Command string
}

func Ps(args []string) error {
	processes, err := readProcesses()
	if err != nil {
		return err
	}

	fmt.Printf("%5s %5s %-5s %s\n", "PID", "PPID", "STAT", "CMD")

	for _, proc := range processes {
		fmt.Printf("%5d %5d %-5s %s\n", proc.PID, proc.PPID, proc.State, proc.Command)
	}

	return nil
}

func readProcesses() ([]ProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("cannot read /proc: %w", err)
	}

	var processes []ProcessInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		proc, err := readProcess(pid)
		if err != nil {
			continue
		}

		processes = append(processes, proc)
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].PID < processes[j].PID
	})

	return processes, nil
}

func readProcess(pid int) (ProcessInfo, error) {
	statPath := filepath.Join("/proc", strconv.Itoa(pid), "stat")

	data, err := os.ReadFile(statPath)
	if err != nil {
		return ProcessInfo{}, err
	}

	ppid, state, comm, err := parseProcStat(string(data))
	if err != nil {
		return ProcessInfo{}, err
	}

	cmd := readCmdline(pid)
	if cmd == "" {
		cmd = comm
	}

	return ProcessInfo{
		PID:     pid,
		PPID:    ppid,
		State:   state,
		Command: cmd,
	}, nil
}

func parseProcStat(stat string) (ppid int, state string, comm string, err error) {
	open := strings.Index(stat, "(")
	close := strings.LastIndex(stat, ")")

	if open == -1 || close == -1 || close <= open {
		return 0, "", "", fmt.Errorf("invalid stat format")
	}

	comm = stat[open+1 : close]

	rest := strings.Fields(stat[close+1:])
	if len(rest) < 2 {
		return 0, "", "", fmt.Errorf("invalid stat fields")
	}

	state = rest[0]

	ppid, err = strconv.Atoi(rest[1])
	if err != nil {
		return 0, "", "", err
	}

	return ppid, state, comm, nil
}

func readCmdline(pid int) string {
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")

	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return ""
	}

	if len(data) == 0 {
		return ""
	}

	parts := strings.Split(string(data), "\x00")

	var args []string
	for _, part := range parts {
		if part != "" {
			args = append(args, part)
		}
	}

	return strings.Join(args, " ")
}
