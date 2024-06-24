package main

import (
	"fmt"

	"github.com/fatih/color"
)

// color functions
var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()

// searchResult represents a gap search searchResult.
type searchResult struct {
	filename   string
	lineNumber int
	text       string
	finished bool
}

// result.String() returns a string representation of the result.
func (r *searchResult) String() string {
	return fmt.Sprintf("%s \n%s: %s\n\n", cyan(r.filename), yellow(r.lineNumber), green(r.text))
}

func (r *searchResult) Filename() string {
	return fmt.Sprint(cyan(r.filename))
}

func (r *searchResult) Line() string {
	return fmt.Sprintf("%s: %s", yellow(r.lineNumber), green(r.text))
}

// resultPrinter takes a channel of results and prints them.
func resultPrinter(results *chan *searchResult) {
	// this fileMap is used to temporarily store the file and individual result entries.
	// all the occurences will be printed once filedone notification is recieved
	fileMap := map[string][]*searchResult{}
	for r := range *results {
		if r.finished {
			fmt.Println(r.Filename())
			for _, v := range fileMap[r.filename] {
				fmt.Println(v.Line())
			}
			fmt.Println()
			delete(fileMap, r.filename)
		} else {
			fileMap[r.filename] = append(fileMap[r.filename], r)
		}
	}
}
