package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sync"

	"github.com/urfave/cli/v2"
)

const (
	NAME    = "gap"
	VERSION = "v0.1.0"
)

type processorConfig struct {
	regexEnabled bool
	pattern      string
	regex        *regexp.Regexp
}

type walkerConfig struct {
	dir            string
	followSymlinks bool
}

func main() {
	// todo add ignore matcher
	// todo read from stdin

	app := &cli.App{
		Name:    NAME,
		Version: VERSION,
		Authors: []*cli.Author{
			{
				Name:  "Shravan Asati",
				Email: "dev.shravan@proton.me",
			},
		},
		Usage:           "a *fast* grep like tool",
		UsageText:       "gap is a fast grep like tool. It searches the given regex or literal text and searches it recursively in the given directory, while ignoring hidden files and folders, binary files and obeying the gitignore patterns.",
		HideHelpCommand: true,
		ArgsUsage:       "PATTERN [PATH]",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:        "workers",
				Aliases:     []string{"w"},
				Value:       0,
				Usage:       "Number of parallel workers for processing text.",
				DefaultText: "as many logical cores",
			},
			&cli.BoolFlag{
				Name:    "regex",
				Aliases: []string{"x"},
				Value:   false,
				Usage:   "Whether the pattern is a regular expression.",
			},
			&cli.BoolFlag{
				Name:    "smart-casing", // todo
				Aliases: []string{"S"},
				Value:   true,
				Usage:   "Whether to use smart-casing. gap will search case-insensitively if all the terms are in lower case. Overrides other case-sensitivity flags.",
			},
			&cli.BoolFlag{
				Name:    "insensitive", // todo
				Aliases: []string{"i"},
				Value:   false,
				Usage:   "Whether to search case-insensitively",
			},
			&cli.BoolFlag{
				Name:    "sensitive", // todo
				Aliases: []string{"c"},
				Value:   false,
				Usage:   "Whether to search case-sensitively",
			},
			&cli.BoolFlag{
				Name:  "stats", // todo
				Value: false,
				Usage: "Show statistics about the search.",
			},
			&cli.BoolFlag{
				Name:    "files-with-matches", // todo
				Aliases: []string{"f"},
				Value:   false,
				Usage:   "Print paths with atleast one match.",
			},
			&cli.BoolFlag{
				Name:    "files-without-matches", // todo
				Aliases: []string{"F"},
				Value:   false,
				Usage:   "Print paths with zero matches.",
			},
			&cli.BoolFlag{
				Name:    "line-number", // todo
				Aliases: []string{"n"},
				Value:   true,
				Usage:   "Print line numbers where matches occur.",
			},
			&cli.BoolFlag{
				Name:    "no-line-number", // todo
				Aliases: []string{"N"},
				Value:   false,
				Usage:   "Don't print line numbers where matches occur.",
			},
			&cli.BoolFlag{
				Name:    "follow",
				Aliases: []string{"L"},
				Value:   false,
				Usage:   "Follow symbolic links.",
			},
			&cli.StringSliceFlag{
				Name:    "glob", // todo
				Aliases: []string{"g"},
				Usage:   "Glob patterns for files to search.",
			},
			&cli.StringSliceFlag{
				Name:    "ignore-glob", // todo
				Aliases: []string{"G"},
				Usage:   "Glob patterns for files to ignore.",
			},
		},
		Action: func(cCtx *cli.Context) error {
			searchPattern := cCtx.Args().Get(0)
			if searchPattern == "" {
				return errors.New("require a search pattern. do `gap -h` for help")
			}
			dir := cCtx.Args().Get(1)
			if dir == "" {
				dir = "."
			}
			walkConfig := &walkerConfig{
				dir: dir,
				followSymlinks: cCtx.Bool("follow"),
			}

			regexEnabled := cCtx.Bool("regex")
			processConfig := &processorConfig{
				regexEnabled: regexEnabled,
				pattern:      searchPattern,
			}
			if regexEnabled {
				re, err := regexp.Compile(searchPattern)
				if err != nil {
					return errors.New("unable to compile regex: " + err.Error())
				}
				processConfig.regex = re
			}

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
				nProcs = max(uint(runtime.NumCPU())/4, 1)
			}
			// the process waitgroup is only used to orchestrate the processor goroutines
			var processWg sync.WaitGroup
			for range nProcs {
				processWg.Add(1)
				go func() {
					process(&processor, &resultCh, processConfig)
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

			if err := walk(walkConfig, &processor); err != nil {
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
