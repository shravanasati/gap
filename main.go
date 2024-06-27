package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

const (
	NAME    = "gap"
	VERSION = "v0.1.0"
)

type processorConfig struct {
	regexEnabled  bool
	pattern       string
	regex         *regexp.Regexp
	caseSensitive bool
	invertMatch   bool
	onlyFiles     bool
	beforeContext int
	afterContext int
}

type walkerConfig struct {
	dir            string
	followSymlinks bool
}

type printerConfig struct {
	showLineNumbers bool
	onlyFiles       bool
	countStats      bool
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
		UsageText:       "gap is a fast grep like tool. It searches the given regex or literal text and searches it recursively in the given directory, while ignoring hidden files and folders, binary files and obeying the gitignore patterns. \n\nProject Home Page: https://github.com/shravanasati/gap \n\n$ gap pattern [path] {flags}",
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
				Name:    "insensitive",
				Aliases: []string{"i"},
				Value:   false,
				Usage:   "Whether to search case-insensitively.",
			},
			&cli.BoolFlag{
				Name:    "sensitive",
				Aliases: []string{"c"},
				Value:   false,
				Usage:   "Whether to search case-sensitively.",
			},
			&cli.BoolFlag{
				Name:    "invert-match",
				Aliases: []string{"V"},
				Value:   false,
				Usage:   "Print lines and files where the given pattern doesn't match.",
			},
			&cli.BoolFlag{
				Name:  "stats",
				Value: false,
				Usage: "Show statistics about the search.",
			},
			&cli.BoolFlag{
				Name:    "files-with-matches",
				Aliases: []string{"f"},
				Value:   false,
				Usage:   "Print paths with atleast one match.",
			},
			&cli.BoolFlag{
				Name:    "files-without-matches",
				Aliases: []string{"F"},
				Value:   false,
				Usage:   "Print paths with zero matches.",
			},
			&cli.BoolFlag{
				Name:    "no-line-number",
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
			&cli.UintFlag{
				Name:    "after-context",
				Aliases: []string{"A"},
				Value:   0,
				Usage:   "Show `NUM` lines after each match.",
			},
			&cli.UintFlag{
				Name:    "before-context",
				Aliases: []string{"B"},
				Value:   0,
				Usage:   "Show `NUM` lines before each match.",
			},
			&cli.UintFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Value:   0,
				Usage:   "Show `NUM` lines before & after each match. Overrides after-context and before-context.",
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
				dir:            dir,
				followSymlinks: cCtx.Bool("follow"),
			}

			invertMatch := cCtx.Bool("invert-match")
			filesWithMatches := cCtx.Bool("files-with-matches")
			filesWithoutMatches := cCtx.Bool("files-without-matches")
			beforeContext := cCtx.Uint("before-context")
			afterContext := cCtx.Uint("after-context")
			context := cCtx.Uint("context")
			if context != 0 {
				beforeContext = context
				afterContext = context
			}
			regexEnabled := cCtx.Bool("regex")
			processConfig := &processorConfig{
				regexEnabled: regexEnabled,
				pattern:      searchPattern,
				invertMatch:  invertMatch,
				onlyFiles:    filesWithMatches || filesWithoutMatches,
				beforeContext: int(beforeContext),
				afterContext: int(afterContext),
			}
			insensitive := cCtx.Bool("insensitive")
			sensitive := cCtx.Bool("sensitive")
			if !sensitive && !insensitive {
				// if no flags are provided, use smart casing
				if strings.ToLower(searchPattern) == searchPattern {
					// if the pattern is all lower case, apply case-insensitive search
					processConfig.caseSensitive = false
				} else {
					// otherwise do case-sensitive search
					processConfig.caseSensitive = true
				}
			} else if sensitive {
				processConfig.caseSensitive = true
			} else {
				processConfig.caseSensitive = false
			}

			if regexEnabled {
				// if regex is enabled, try compiling the pattern
				// in regex we can take advantage of the case-insensitive mode
				if !processConfig.caseSensitive {
					searchPattern = "(?i)" + searchPattern
				}
				re, err := regexp.Compile(searchPattern)
				if err != nil {
					return errors.New("unable to compile regex: " + err.Error())
				}
				processConfig.regex = re
			}

			noLineNumber := cCtx.Bool("no-line-number")
			countStats := cCtx.Bool("stats")

			printConfig := &printerConfig{
				showLineNumbers: !noLineNumber,
				countStats:      countStats,
				onlyFiles:       filesWithMatches || filesWithoutMatches,
			}

			if filesWithoutMatches {
				// files without matches is essentially printing all the files
				// where the match is inverted
				processConfig.invertMatch = true
			}

			processor := make(chan string)
			resultCh := make(chan *searchResult)
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				resultPrinter(&resultCh, printConfig)
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
