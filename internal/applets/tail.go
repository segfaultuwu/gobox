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
	Register("tail", "print last lines of files", Tail)
}

func Tail(args []string) error {
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
			raw := strings.TrimPrefix(arg, "-n")

			value, err := strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("invalid line count: %s", raw)
			}

			if value < 0 {
				return fmt.Errorf("line count cannot be negative")
			}

			n = value

		case strings.HasPrefix(arg, "--lines="):
			raw := strings.TrimPrefix(arg, "--lines=")

			value, err := strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("invalid line count: %s", raw)
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
		return tailReader(os.Stdin, n)
	}

	for index, path := range files {
		if len(files) > 1 {
			if index > 0 {
				fmt.Println()
			}

			fmt.Printf("==> %s <==\n", path)
		}

		if err := tailFile(path, n); err != nil {
			return err
		}
	}

	return nil
}

func tailFile(path string, n int) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	defer file.Close()

	return tailReader(file, n)
}

func tailReader(r io.Reader, n int) error {
	if n == 0 {
		return nil
	}

	scanner := bufio.NewScanner(r)

	buffer := make([]string, 0, n)

	for scanner.Scan() {
		line := scanner.Text()

		if len(buffer) < n {
			buffer = append(buffer, line)
			continue
		}

		copy(buffer, buffer[1:])
		buffer[len(buffer)-1] = line
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for _, line := range buffer {
		fmt.Println(line)
	}

	return nil
}
