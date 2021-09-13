package main

import (
	"fmt"
	// "sync"
)

func main() {
	processor := make(chan string, 10)
	resultCh := make(chan *result, 10)

	go resultPrinter(&resultCh)
	go process(&processor, &resultCh)

	walk(`C:/Users/LENOVO/Documents/binod`, &processor)
	fmt.Scanln()

	// wg := &sync.WaitGroup{}

	// wg.Add(3)

	// go func() {
	// 	process(&processor, &resultCh)
	// 	fmt.Println("closing results")
	// 	close(resultCh)
	// 	wg.Done()
	// }()

	// go func() {
	// 	resultPrinter(&resultCh)
	// 	wg.Done()
	// }()

	// go func() {
	// 	walk(`C:/Users/LENOVO/Documents/binod`, &processor)
	// 	fmt.Println("closing processor")
	// 	close(processor)
	// 	wg.Done()
	// }()



	// wg.Wait()
}