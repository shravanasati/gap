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

			if config.onlyFiles {
				// dont need to worry about sending contextual lines here
				for i, line := range data {
					if searchMethod(line) {
						*out <- &searchResult{filename: entry, lineNumber: i + 1, text: string(line)}
						count++
						break
					}
				}
			} else {
				N := len(data)
				for i, line := range data {
					if searchMethod(line) {
						count++

						// before context
						for j := max(0, i-config.beforeContext); j < i; j++ {
							*out <- &searchResult{
								filename:   entry,
								lineNumber: j + 1,
								text:       string(data[j]),
								contextual: true,
							}
						}

						// actual match
						*out <- &searchResult{
							filename:   entry,
							lineNumber: i + 1,
							text:       string(line),
							contextual: false,
						}

						// after context
						for j := min(N-1, i+1); j < i+config.afterContext+1 && j < N; j++ {
							*out <- &searchResult{
								filename:   entry,
								lineNumber: j + 1,
								text:       string(data[j]),
								contextual: true,
							}
						}
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
