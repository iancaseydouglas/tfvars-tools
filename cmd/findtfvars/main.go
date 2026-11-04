package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func main() {
	start := flag.String("d", ".", "Start searching from this directory")
	ext := flag.String("f", ".tfvars", "File extension to match")
	maxDepth := flag.Int("max-depth", envInt("FINDTFVARS_MAX_DEPTH", -1), "Parent directories to climb (-1 = unlimited, 0 = start dir only)")
	stopAt := flag.String("stop-at", os.Getenv("FINDTFVARS_STOP_AT"), "Stop at the first directory containing this entry (e.g. .git)")
	flag.Parse()

	dir, err := filepath.Abs(*start)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for depth := 0; *maxDepth < 0 || depth <= *maxDepth; depth++ {
		files, err := filepath.Glob(filepath.Join(dir, "*"+*ext))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, f := range files {
			fmt.Println(f)
		}
		if *stopAt != "" {
			if _, err := os.Stat(filepath.Join(dir, *stopAt)); err == nil {
				break
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}
