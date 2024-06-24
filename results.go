package main

import "fmt"
import "github.com/fatih/color"

// result represents a gap search result.
type result struct {
	filename   string
	lineNumber int
	text       string
}

// color functions

var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()


// result.String() returns a string representation of the result.
func (r *result) String() string {
	return fmt.Sprintf("%s \n%s: %s\n\n", cyan(r.filename), yellow(r.lineNumber), green(r.text))
}

// resultPrinter takes a channel of results and prints them.
func resultPrinter(results *chan *result) {
	for r := range *results {
		fmt.Println(r.String())
	}
}