package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    dir, _ := os.Getwd()
    for {
        files, _ := filepath.Glob(filepath.Join(dir, "*.tfvars"))
        for _, f := range files {
            fmt.Println(f)
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
}
