package applets

import "os"

func init() {
	Register("exit", "exit shell", Exit)
}

func Exit(args []string) error {
	os.Exit(0)
	return nil
}
