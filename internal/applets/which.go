package applets

import (
	"fmt"
	"os/exec"
)

func init() {
	Register("which", "find command path", Which)
}

func Which(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command")
	}

	for _, name := range args {
		path, err := exec.LookPath(name)
		if err != nil {
			return err
		}

		fmt.Println(path)
	}

	return nil
}
