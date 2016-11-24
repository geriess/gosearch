package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	inputDir   string
	searchText string
)

func init() {
	flag.StringVar(&inputDir, "p", "", "Path to directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
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

func main() {
	fmt.Println("SearchInText Utility written in Go.")
	flag.Parse()

	ok := true

	if inputDir == "" {
		ok = errorOut("Error: Missing path to directory")
	}

	if searchText == "" {
		ok = errorOut("Error: Missing keyword to search")
	}

	if !ok {
		fmt.Println("Usage:")
		fmt.Println("    searchintext -p path -k keyword")
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(inputDir)
	errorCheck(err)

	for _, file := range files {
		fmt.Println(file.Name(), file.Size(), file.IsDir())
	}
}
