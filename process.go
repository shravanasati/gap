package main

import (
	"bytes"
	// "fmt"
	"io/ioutil"
	"log"
	"sync"
)

func process(in *chan string, out *chan *result) {
	wg := new(sync.WaitGroup)

	for entry := range *in {
		// fmt.Println(entry)
		wg.Add(1)
		go func(entry string) {
			content, err := ioutil.ReadFile(entry)
			if err != nil {
				log.Fatal("unable to read file", err)
			}

			data := bytes.Split(content, []byte("\n"))
			for i, line := range data {
				if bytes.Contains(line, []byte("package")) {
					*out <- &result{entry, i, string(line)}
				}
			}

			wg.Done()
		}(entry)
	}

	wg.Wait()
}