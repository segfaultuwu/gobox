package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gobox/internal/applets"
)

func main() {
	name := normalizeCommand(filepath.Base(os.Args[0]))

	var cmd string
	var args []string

	if name == "gobox" {
		if len(os.Args) < 2 {
			cmd = "sh"
			args = nil
		} else {
			cmd = normalizeCommand(os.Args[1])
			args = os.Args[2:]
		}
	} else {
		cmd = name
		args = os.Args[1:]
	}

	applet, ok := applets.Registry[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "gobox: unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err := applet.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func normalizeCommand(name string) string {
	name = filepath.Base(name)

	// Dropbear/OpenSSH login shells often execute argv[0] as "-sh".
	name = strings.TrimPrefix(name, "-")

	switch name {
	case "shell":
		return "sh"
	default:
		return name
	}
}

func printHelp() {
	fmt.Println("gobox - tiny busybox-like utilities")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gobox <command> [args...]")
	fmt.Println()
	fmt.Println("Commands:")

	names := make([]string, 0, len(applets.Registry))
	for name := range applets.Registry {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		fmt.Println("  " + name)
	}
}
