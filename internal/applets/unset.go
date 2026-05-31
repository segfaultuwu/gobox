package applets

import (
	"fmt"
	"os"
)

func init() {
	Register("unset", "unset environment variable", Unset)
}

func Unset(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing variable name")
	}

	for _, name := range args {
		if err := os.Unsetenv(name); err != nil {
			return err
		}
	}

	return nil
}
