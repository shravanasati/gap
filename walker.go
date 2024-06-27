package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"unicode"
	// "unicode/utf8"

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
		// r, size := utf8.DecodeRune(buf[i:])
		// if r == utf8.RuneError {
		// 	// Invalid UTF-8 encoding found, assuming binary file
		// 	return true, nil
		// }
		if !unicode.IsPrint(rune(buf[i])) && !unicode.IsSpace(rune(buf[i])) {
			return true, nil
		}
		// i += size
	}
	return false, nil
}

func walk(config *walkerConfig, processor *chan string) error {
	matcher, err := gitignore.NewGitignoreMatcher().FromFile("./.gitignore")
	if err != nil {
		// if there is some error, we can let it pass
		// but the matcher would be nil
		// so setting it to a empty matcher
		matcher = gitignore.NewGitignoreMatcher()
	}
	err = matcher.Build()
	if err != nil {
		return err
	}

	baseDirLength := len(config.dir) + 1 // +1 for trailing slash

	visit := func(path string, f fs.DirEntry, err error) error {
		basePath := filepath.Base(path)
		if f.IsDir() && (basePath == ".git") {
			return fastwalk.SkipDir
		}

		ignored, err := matcher.Matches(path[min(baseDirLength, len(path) - 1):]) // strip the base dir prefix
		if err == nil && ignored {
			// no need to have various checks if the file is ignored
			// if f.IsDir() {
			// 	return fastwalk.ErrSkipFiles
			// }
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

	err = fastwalk.Walk(
		&fastwalk.Config{
			NumWorkers: fastwalk.DefaultNumWorkers(),
			Follow:     config.followSymlinks,
		},
		config.dir, visit,
	)
	close(*processor)
	return err
}
