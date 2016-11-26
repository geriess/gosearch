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
	fmt.Println("    gosearch [OPTIONS] -p path -k keyword")
	flag.PrintDefaults()
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

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func init() {
	flag.StringVar(&inputDir, "p", "", "Path to directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
	flag.BoolVar(&verbose, "v", false, "Verbose (prints all files searched)")
	//TODO option to return in JSON format
	//TODO option to exclude hidden files in search
}

func main() {
	start := time.Now()
	fmt.Println("==================================")
	fmt.Println("gosearch: A search in text utility written in Go.")
	fmt.Println("searching...")
	fmt.Println("==================================")

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
	// walk directory
	err := filepath.Walk(inputDir, func(path string, f os.FileInfo, _ error) error {
		// if files
		if !f.IsDir() {
			// check if file exists
			exist := exists(path)
			if !exist {
				if !verbose {
					return nil
				}
				fmt.Printf("%s DOES NOT EXIST\n", path)
			} else {
				filesVisited++
				// read file
				content, err := ioutil.ReadFile(path)
				if err != nil {
					if !verbose {
						fmt.Print("")
						return nil
					}
					fmt.Printf("%s FILE cannot be read\n", path)
					return nil
				}

				// convert file contents to string
				x := string(content)

				// search file contents for keyword
				search := strings.Contains(x, searchText)
				switch search {
				case true:
					filesFound++
					fmt.Printf("%s FILE contains %s\n", path, searchText)
				case false:
					if !verbose {
						return nil
					}
					fmt.Printf("%s FILE does not contain %s\n", path, searchText)
				}
			}
		} else {
			// if directories
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

	// print summary
	fmt.Println("==================================")
	log.Println("")
	fmt.Printf("Done searching for %s\n", searchText)
	fmt.Printf("Path: %s\n", inputDir)
	fmt.Printf("Checked %d files in %d directories\n", filesVisited, foldersVisited)
	fmt.Printf("Found %d files containing %s\n", filesFound, searchText)
	fmt.Println("==================================")
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s\n", elapsed)
	//defer duration(time.Now(), "init")
}
