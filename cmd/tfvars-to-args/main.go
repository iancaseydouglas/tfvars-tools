package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	reverse := flag.Bool("reverse", false, "Keep input order (default: reverse it, so outermost files come first)")
	format := flag.String("format", "-var-file=%s", "Printf format for each argument")
	verbose := flag.Bool("verbose", false, "Print files in precedence order to stderr")
	flag.Parse()

	if v := os.Getenv("TFVARS_TO_ARGS_FORMAT"); v != "" && !isFlagSet("format") {
		*format = v
	}

	var files []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			files = append(files, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if !*reverse {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	args := make([]string, len(files))
	for i, file := range files {
		if *verbose {
			fmt.Fprintln(os.Stderr, file)
		}
		args[i] = fmt.Sprintf(*format, file)
	}
	fmt.Println(strings.Join(args, " "))
}

func isFlagSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}
