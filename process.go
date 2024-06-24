package main

import (
	"bytes"
	// "fmt"
	"io/ioutil"
	"log"
	"sync"
)

func process(in *chan string, out *chan *searchResult, searchTerm string) {
	wg := new(sync.WaitGroup)
	toSearch := []byte(searchTerm)
	newLineSep := []byte("\n")

	for entry := range *in {
		wg.Add(1)
		go func(entry string) {
			content, err := ioutil.ReadFile(entry)
			if err != nil {
				log.Fatal("unable to read file", err)
				return
			}

			data := bytes.Split(content, newLineSep)
			count := 0
			for i, line := range data {
				if bytes.Contains(line, toSearch) {
					*out <- &searchResult{filename: entry, lineNumber: i + 1, text: string(line), finished: false}
					count++
				}
			}

			if count > 0 {
				// ṇotify that the file is finished reading
				*out <- &searchResult{filename: entry, finished: true}
			}

			wg.Done()
		}(entry)
	}

	wg.Wait()
	// todo move this logic to a processManager
	gState.Lock()
	defer gState.Unlock()
	if !gState.resultChClosed {
		close(*out)
		gState.resultChClosed = true
	}
}
