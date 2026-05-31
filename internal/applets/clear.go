package applets

import "fmt"

func init() {
	Register("clear", "clear terminal screen", Clear)
}

func Clear(args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}
