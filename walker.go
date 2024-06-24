package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"net/http"

	"github.com/charlievieth/fastwalk"
)

func isBinaryFile(filename string) (bool, error) {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal("cant open file: " + filename)
        return false, err
    }
    defer file.Close()

    buffer := make([]byte, 512)
    if _, err = file.Read(buffer); err != nil && err != io.EOF {
        log.Fatal("cant read file: ", filename)
        return false, err
    }

    contentType := http.DetectContentType(buffer)

    return !strings.HasPrefix(contentType, "text/"), nil
}


func walk(dir string, processor *chan string) error {
	visit := func(path string, f fs.DirEntry, err error) error {
		if f.IsDir() && filepath.Base(path) == ".git" {
			return fastwalk.SkipDir
		}

		info, err := f.Info()
        if !info.Mode().IsRegular() {
            return nil
        }
		if err != nil {
            log.Fatal("cant get file info: " + f.Name())
			return err
		}
        isBinary, err := isBinaryFile(f.Name())
        if err != nil {
            return err
        }
		if !isBinary {
			*processor <- path
		}
		return nil
	}

	err := fastwalk.Walk(&fastwalk.DefaultConfig, dir, visit)
    close(*processor)
    return err
}
