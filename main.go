package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/urfave/cli/v2"
)

const (
	NAME    = "gap"
	VERSION = "v0.1.0"
)

func main() {
	// todo add ignore matcher
	// todo add regex feature

	app := &cli.App{
		Name:    NAME,
		Version: VERSION,
		Authors: []*cli.Author{
			{
				Name:  "Shravan Asati",
				Email: "dev.shravan@proton.me",
			},
		},
		Usage:     "a *fast* grep like tool",
		UsageText: "gap is a fast grep like tool. It searches the given regex or literal text and searches it recursively in the given directory, while ignoring hidden files and folders, binary files and obeying the gitignore patterns.",
		HideHelpCommand: true,
		ArgsUsage: "PATTERN [PATH]",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name: "workers",
				Aliases: []string{"w"},
				Value: 0,
				Usage: "Number of parallel workers for processing text.",
				DefaultText: "as many logical cores",
			},
			&cli.BoolFlag{
				Name: "regex",
				Aliases: []string{"x"},
				Value: false,
				Usage: "Whether the pattern is a regular expression.",
			},
		},
		Action: func(cCtx *cli.Context) error {
			searchPattern := cCtx.Args().Get(0)
			if searchPattern == "" {
				return errors.New("require a search pattern. do `gap -h` for help.")
			}
			dir := cCtx.Args().Get(1)
			if dir == "" {
				dir = "."
			}

			// regexEnabled := cCtx.Bool("regex")

			processor := make(chan string)
			resultCh := make(chan *searchResult)
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				resultPrinter(&resultCh)
				wg.Done()
			}()

			nProcs := cCtx.Uint("workers")
			if nProcs == 0 {
				nProcs = uint(runtime.NumCPU())
			}
			// the process waitgroup is only used to orchestrate the processor goroutines
			var processWg sync.WaitGroup
			for range nProcs {
				processWg.Add(1)
				go func() {
					// these goroutines will deadlock when the channel is not closed
					process(&processor, &resultCh, searchPattern)
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

			if err := walk(dir, &processor); err != nil {
				log.Fatal(err)
			}
			wg.Wait()
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("error:", err)
	}

}
