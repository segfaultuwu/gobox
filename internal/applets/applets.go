package applets

import (
	"fmt"
	"sort"
)

func init() {
	Register("applets", "list available applets", Applets)
}

func Applets(args []string) error {
	names := make([]string, 0, len(Registry))

	for name := range Registry {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		fmt.Println(name)
	}

	return nil
}
