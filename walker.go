package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"unicode"

	"github.com/charlievieth/fastwalk"
	"github.com/shravanasati/gap/gitignore"
)

func isBinaryFile(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("cant open file: ", filename, err)
		return false, err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatal("cannot read file:", filename, err)
		return false, err
	}

	for i := 0; i < n; i++ {
		if !unicode.IsPrint(rune(buf[i])) && !unicode.IsSpace(rune(buf[i])) {
			return true, nil
		}
	}
	return false, nil
}

func walk(dir string, processor *chan string) error {
	matcher, err := gitignore.NewGitignoreMatcher().FromFile("./.gitignore")
	if err != nil {
		log.Println("gitignore not found", err)
		// if there is some error, we can let it pass
		// but the matcher would be nil
		// so setting it to a empty matcher
		matcher = gitignore.NewGitignoreMatcher()
	}

	visit := func(path string, f fs.DirEntry, err error) error {
		if f.IsDir() && (filepath.Base(path) == ".git" || matcher.IsIgnored(path)) {
			return fastwalk.SkipDir
		}

		if matcher.IsIgnored(path) {
			// no need to have various checks if the file is ignored
			return nil
		}

		info, err := f.Info()
		if err != nil {
			log.Fatal("cant get file info:", f.Name(), err.Error())
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		isBinary, err := isBinaryFile(filepath.Join(".", path))
		if err != nil {
			return nil
		}
		if !isBinary {
			*processor <- path
		}
		return nil
	}

	err = fastwalk.Walk(&fastwalk.DefaultConfig, dir, visit)
	close(*processor)
	return err
}
