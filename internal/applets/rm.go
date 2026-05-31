package applets

import (
	"fmt"
	"os"

	"gobox/internal/utils"
)

func init() {
	Register("rm", "remove files or directories", Rm)
}

func Rm(args []string) error {
	parsed := utils.ParseArgs(args)

	recursive := parsed.Has("r", "R", "recursive")
	force := parsed.Has("f", "force")

	if len(parsed.Rest) == 0 {
		if force {
			return nil
		}
		return fmt.Errorf("missing operand")
	}

	for _, path := range parsed.Rest {
		var err error

		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil && !force {
			return err
		}
	}

	return nil
}
