package applets

import (
	"fmt"
	"os"
)

func init() {
	Register("pwd", "print current working directory", Pwd)
}

func Pwd(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println(cwd)
	return nil
}
