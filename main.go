package main

import (
	"log"
	"os"
	"sync"
)

type globalState struct {
	resultChClosed bool
	sync.Mutex
}

var gState = globalState{resultChClosed: false}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("require a search term")
		return
	}
	searchTerm := os.Args[1]

	processor := make(chan string)
	resultCh := make(chan *result)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		resultPrinter(&resultCh)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		process(&processor, &resultCh, searchTerm)
		wg.Done()
	}()

	if err := walk(`.`, &processor); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
