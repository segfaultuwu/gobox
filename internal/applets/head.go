package applets

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func init() {
	Register("head", "show last N lines (default 10)", Head)
}

func Head(args []string) error {
	n := 10
	var files []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-n" || arg == "--lines":
			if i+1 >= len(args) {
				return fmt.Errorf("missing line count after %s", arg)
			}

			value, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid line count: %s", args[i+1])
			}

			if value < 0 {
				return fmt.Errorf("line count cannot be negative")
			}

			n = value
			i++

		case strings.HasPrefix(arg, "-n") && len(arg) > 2:
			value, err := strconv.Atoi(strings.TrimPrefix(arg, "-n"))
			if err != nil {
				return fmt.Errorf("invalid line count: %s", strings.TrimPrefix(arg, "-n"))
			}

			if value < 0 {
				return fmt.Errorf("line count cannot be negative")
			}

			n = value

		case strings.HasPrefix(arg, "--lines="):
			valueRaw := strings.TrimPrefix(arg, "--lines=")

			value, err := strconv.Atoi(valueRaw)
			if err != nil {
				return fmt.Errorf("invalid line count: %s", valueRaw)
			}

			if value < 0 {
				return fmt.Errorf("line count cannot be negative")
			}

			n = value

		default:
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return headReader(os.Stdin, n)
	}

	for index, path := range files {
		if len(files) > 1 {
			if index > 0 {
				fmt.Println()
			}

			fmt.Printf("==> %s <==\n", path)
		}

		if err := headFile(path, n); err != nil {
			return err
		}
	}

	return nil
}

func headFile(path string, n int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer file.Close()

	return headReader(file, n)
}

func headReader(r io.Reader, n int) error {
	if n == 0 {
		return nil
	}

	scanner := bufio.NewScanner(r)

	count := 0
	for scanner.Scan() {
		fmt.Println(scanner.Text())

		count++
		if count >= n {
			break
		}
	}

	return scanner.Err()
}
