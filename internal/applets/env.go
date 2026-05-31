package applets

import (
	"fmt"
	"os"
)

func init() {
	Register("env", "print environment variables", Env)
}

func Env(args []string) error {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}

	return nil
}
