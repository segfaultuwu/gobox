package applets

import (
	"fmt"
	"os"

	"gobox/internal/utils"
)

func init() {
	Register("mkdir", "create directories", Mkdir)
}

func Mkdir(args []string) error {
	parsed := utils.ParseArgs(args)

	parents := parsed.Has("p", "parents")

	if len(parsed.Rest) == 0 {
		return fmt.Errorf("missing operand")
	}

	for _, path := range parsed.Rest {
		var err error

		if parents {
			err = os.MkdirAll(path, 0755)
		} else {
			err = os.Mkdir(path, 0755)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
