package applets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gobox/internal/utils"
)

func init() {
	Register("ls", "list directory contents", Ls)
}

func Ls(args []string) error {
	parsed := utils.ParseArgs(args)

	showAll := parsed.Has("a", "all")
	long := parsed.Has("l", "long")

	paths := parsed.Rest
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for i, path := range paths {
		if len(paths) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", path)
		}

		if err := listPath(path, showAll, long); err != nil {
			return err
		}
	}

	return nil
}

func listPath(path string, showAll bool, long bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		printEntry(path, info, long)
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()

		if !showAll && len(name) > 0 && name[0] == '.' {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		fullPath := filepath.Join(path, name)
		printEntry(fullPath, info, long)
	}

	return nil
}

func printEntry(path string, info os.FileInfo, long bool) {
	name := filepath.Base(path)

	if long {
		fmt.Printf("%s %8d %s\n", info.Mode().String(), info.Size(), name)
		return
	}

	if info.IsDir() {
		fmt.Printf("%s/\n", name)
		return
	}

	fmt.Println(name)
}
