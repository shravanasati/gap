package main

import (
	"bytes"
	// "fmt"
	"io/ioutil"
	"log"
	"sync"
)

type searchCallable func([]byte) bool

func process(in *chan string, out *chan *searchResult, config *processorConfig) {
	wg := new(sync.WaitGroup)
	toSearch := []byte(config.pattern)
	newLineSep := []byte("\n")

	for entry := range *in {
		wg.Add(1)
		go func(entry string) {
			defer wg.Done()

			content, err := ioutil.ReadFile(entry)
			if err != nil {
				log.Fatal("unable to read file", err)
				return
			}

			count := 0

			data := bytes.Split(content, newLineSep)
			var searchMethod searchCallable
			if config.regexEnabled {
				searchMethod = func(b []byte) bool {
					return config.regex.Match(b)
				}
			} else {
				if config.caseSensitive {
					searchMethod = func(b []byte) bool {
						return bytes.Contains(b, toSearch)
					}
				} else {
					// for case-insensitive search, convert both search pattern
					// and the given bytes to lower case
					toSearch = bytes.ToLower(toSearch)
					searchMethod = func(b []byte) bool {
						return bytes.Contains(bytes.ToLower(b), toSearch)
					}
				}
			}

			if config.invertMatch {
				oldSearchMethod := searchMethod
				searchMethod = func(b []byte) bool {
					return !oldSearchMethod(b)
				}
			}

			for i, line := range data {
				if searchMethod(line) {
					*out <- &searchResult{filename: entry, lineNumber: i + 1, text: string(line), finished: false}
					count++
					if config.onlyFiles {
						break
					}
				}
			}

			if count > 0 {
				// notify that the file is finished reading
				*out <- &searchResult{filename: entry, finished: true}
			}

		}(entry)
	}

	wg.Wait()
}
