package main

import (
	// "fmt"
	"os"
	"path/filepath"
	"sync"
)

var wg sync.WaitGroup

func walk(dir string, processor *chan string) {
    visit := func(path string, f os.FileInfo, err error) error {
		wg.Add(1)
        if f.IsDir() && path != dir && filepath.Base(path) != ".git" {
            go walk(path, processor)
            return filepath.SkipDir
        }

        if f.Mode().IsRegular() {
            *processor <- path
        }
		wg.Done()
        return nil
    }

    filepath.Walk(dir, visit)
}

