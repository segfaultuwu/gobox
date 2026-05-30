package applets

import (
	"fmt"
	"strings"
)

func init() {
	Register("echo", "print arguments", Echo)
}

func Echo(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}
