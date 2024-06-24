package main

import (
	"log"
	"os"
	"runtime"
	"sync"
)

func main() {
	// todo add ignore matcher
	// todo add regex feature
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

	nProcs := runtime.GOMAXPROCS(-1)
	// the process waitgroup is only used to orchestrate the processor goroutines
	var processWg sync.WaitGroup
	for range nProcs {
		processWg.Add(1)
		go func() {
			// these goroutines will deadlock when the channel is not closed
			process(&processor, &resultCh, searchTerm)
			processWg.Done()
		}()
	}

	// this goroutine will wait for all processors to finish
	// once that's done, it will close the result channel so that
	// result printer can stop too
	// wg waitgroup only handles resultPrinter and the below goroutine
	wg.Add(1)
	go func() {
		processWg.Wait()
		close(resultCh)
		wg.Done()
	}()

	if err := walk(`.`, &processor); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
