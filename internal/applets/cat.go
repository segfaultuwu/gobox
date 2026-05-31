package applets

import (
	"fmt"
	"os"
	"strings"
)

func init() {
	Register("cat", "display the contents of a file", Cat)
}

func Cat(args []string) error {
	data, err := os.ReadFile(strings.Join(args, ""))
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
