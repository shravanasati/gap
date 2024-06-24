package main

import (
	"log"
	"os"

	// "runtime"
	"sync"
)

type globalState struct {
	resultChClosed bool
	sync.Mutex
}

var gState = globalState{resultChClosed: false}

func main() {
	// todo add ignore matcher
	if len(os.Args) < 2 {
		log.Fatal("require a search term")
		return
	}
	searchTerm := os.Args[1]

	processor := make(chan string)
	resultCh := make(chan *searchResult)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		resultPrinter(&resultCh)
		wg.Done()
	}()

	// nProcs := runtime.GOMAXPROCS(-1)
	for range 1 {
		wg.Add(1)
		go func() {
			// these goroutines will 
			process(&processor, &resultCh, searchTerm)
			wg.Done()
		}()
	}

	if err := walk(`.`, &processor); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
