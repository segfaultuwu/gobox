package applets

import (
	"fmt"
	"os"
	"strings"
)

func init() {
	Register("export", "set environment variable", Export)
}

func Export(args []string) error {
	if len(args) == 0 {
		return Env(nil)
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid export: %s", arg)
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1]

		if key == "" {
			return fmt.Errorf("empty variable name")
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}
