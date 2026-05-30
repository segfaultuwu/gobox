package applets

import (
	"os"
	"strings"
)

func init() {
	Register("cd", "change directory", Cd)
}

func Cd(args []string) error {
	err := os.Chdir(strings.Join(args, ""))
	if err != nil {
		return err
	}
	return nil
}
