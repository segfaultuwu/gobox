package applets

import "fmt"

func init() {
	Register("true", "always returns true", True)
}

func True(args []string) error {
	for true {
		fmt.Println(true)
	}
	return nil
}
