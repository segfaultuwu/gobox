package applets

import (
	"fmt"
	"sort"
)

func init() {
	Register("help", "show available applets", Help)
}

func Help(args []string) error {
	fmt.Println("gobox applets:")
	fmt.Println()

	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		applet := Registry[name]
		fmt.Printf("  %-10s %s\n", applet.Name, applet.Description)
	}

	return nil
}
