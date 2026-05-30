package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gobox/internal/applets"
)

func main() {
	name := filepath.Base(os.Args[0])

	var cmd string
	var args []string

	if name == "gobox" {
		if len(os.Args) < 2 {
			printHelp()
			os.Exit(1)
		}

		cmd = os.Args[1]
		args = os.Args[2:]
	} else {
		cmd = name
		args = os.Args[1:]
	}

	applet, ok := applets.Registry[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "gobox: unknown command: %s\n", cmd)
		os.Exit(1)
	}

	if err := applet.Exec(args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("gobox - tiny busybox-like utilities")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gobox <command> [args...]")
	fmt.Println()
	fmt.Println("Commands:")

	for name := range applets.Registry {
		fmt.Println("  " + name)
	}
}
