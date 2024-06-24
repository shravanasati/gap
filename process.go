package main

import (
	"bytes"
	// "fmt"
	"io/ioutil"
	"log"
	"sync"
)

func process(in *chan string, out *chan *result, searchTerm string) {
	wg := new(sync.WaitGroup)
	toSearch := []byte(searchTerm)
	newLineSep := []byte("\n")

	for entry := range *in {
		// fmt.Println(entry)
		wg.Add(1)
		go func(entry string) {
			content, err := ioutil.ReadFile(entry)
			if err != nil {
				log.Fatal("unable to read file", err)
				return
			}

			data := bytes.Split(content, newLineSep)
			for i, line := range data {
				if bytes.Contains(line, toSearch) {
					*out <- &result{entry, i, string(line)}
				}
			}

			wg.Done()
		}(entry)
	}

	wg.Wait()
	gState.Lock()
	defer gState.Unlock()
	if !gState.resultChClosed {
		close(*out)
	}
}