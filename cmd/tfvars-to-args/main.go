package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	reverse := flag.Bool("reverse", false, "Reverse the order of var-file arguments")
	flag.Parse()

	var files []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	if !*reverse {
		// Reverse the slice
		for i := len(files)/2 - 1; i >= 0; i-- {
			opp := len(files) - 1 - i
			files[i], files[opp] = files[opp], files[i]
		}
	}

	for _, file := range files {
		fmt.Printf("-var-file=%s ", file)
	}
	fmt.Println()
}
