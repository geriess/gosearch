package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	inputDir   string
	searchText string
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("    gosearch -p path -k keyword")
}

// error check helper
func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// error print helper
func errorOut(message string) bool {
	fmt.Fprintln(os.Stderr, message)
	return false
}

func init() {
	flag.StringVar(&inputDir, "p", "", "Path to directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
}

func main() {
	fmt.Println("gosearch: A search in text utility written in Go.")
	flag.Parse()
	ok := true

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
			// read file
			content, err := ioutil.ReadFile(path)
			errorCheck(err)
			x := string(content)

			// search for keyword
			search := strings.Contains(x, searchText)
			if !search {
				fmt.Printf("Checked File: %s does not contain %s\n", path, searchText)
			} else {
				fmt.Printf("Checked File: %s contains %s\n", path, searchText)
			}
		} else {
			fmt.Printf("Checked Directory: %s\n", path)
		}
		return nil
	})
	errorCheck(err)
}
