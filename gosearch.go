package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	inputDir   string
	searchText string
	verbose    bool
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("    gosearch -p path -k keyword")
}

func duration(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("func %s elapsed %s\n", name, elapsed)
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errorOut(message string) bool {
	fmt.Fprintln(os.Stderr, message)
	return false
}

func init() {
	flag.StringVar(&inputDir, "p", "", "Path to directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
	flag.BoolVar(&verbose, "v", false, "verbose")
}

func main() {
	start := time.Now()

	fmt.Println("gosearch: A search in text utility written in Go.")

	flag.Parse()

	ok := true
	filesFound := 0
	filesVisited := 0
	foldersVisited := 0

	if inputDir == "" {
		ok = errorOut("Error: Missing path to directory")
	}
	if searchText == "" {
		ok = errorOut("Error: Missing keyword to search")
	}
	if !ok {
		usage()
		os.Exit(1)
	}

	err := filepath.Walk(inputDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			filesVisited++
			// read file
			content, err := ioutil.ReadFile(path)
			errorCheck(err)
			x := string(content)

			// search for keyword
			search := strings.Contains(x, searchText)
			switch search {
			case true:
				filesFound++
				fmt.Printf("%s FILE contains %s\n", path, searchText)
			case false:
				if !verbose {
					// hide non-matches
					fmt.Printf("")
				} else {
					// show non-matches
					fmt.Printf("%s FILE does not contain %s\n", path, searchText)
				}
			}
		} else {
			foldersVisited++
			if !verbose {
				fmt.Printf("")
			} else {
				fmt.Printf("%s FOLDER\n", path)
			}
		}
		return nil
	})
	errorCheck(err)

	// summary
	fmt.Println("==================================")
	fmt.Printf("Done searching for %s\n", searchText)
	fmt.Printf("Path: %s\n", inputDir)
	fmt.Printf("Checked %d files in %d directories\n", filesVisited, foldersVisited)
	fmt.Printf("Found %d files containing %s\n", filesFound, searchText)
	fmt.Println("==================================")
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s\n", elapsed)
	//defer duration(time.Now(), "init")
}
