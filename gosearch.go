package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	inputDir    string         // user input; top-level path to search
	searchText  string         // user input; keyword to search
	verbose     bool           // user input; if true displays all paths
	numFound    int            // # of files matching keyword
	fileVisit   int            // # of files visited by search
	folderVisit int            // # of folders visited during search
	wg          sync.WaitGroup // synchronize channels and goroutines
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

func readDir(path string) os.FileInfo {
	var x os.FileInfo
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		x = file
		fmt.Println(file.Name(), file.IsDir())
	}
	return x
}

func walkDir(path string, f os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%s with %d bytes\n", path, f.Size())
	}
	return nil
}

func getFiles(path string) {
	err := filepath.Walk(path, walkDir)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

type walkresult struct {
	path  string
	found bool
}

func walkFiles(directory string, keyword string) (<-chan walkresult, <-chan error) {
	// create channels
	filesFound := make(chan walkresult)
	walkResult := make(chan error, 1)

	// launch goroutine to walk path; add wait count
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			// if path is file, then launch search process
			if !f.IsDir() {
				fileVisit++

				// launch goroutine to read file; add wait count
				wg.Add(1)
				go func() {
					defer wg.Done()
					content, err := ioutil.ReadFile(path)
					if err != nil {
						if !verbose {
							fmt.Print("")
							return
						}
						fmt.Printf("%s FILE cannot be read\n", path)
						return
					}

					// convert file contents to string
					x := string(content)

					// search file contents for keyword
					search := strings.Contains(x, keyword)

					switch search {
					case true:
						numFound++
						found := true
						filesFound <- walkresult{path, found}
					case false:
						found := false
						filesFound <- walkresult{path, found}
					}
				}()
			}
			folderVisit++
			return nil
		})
		go func() {
			wg.Wait()
			close(filesFound)
			close(walkResult)
		}()
		walkResult <- err
	}()

	return filesFound, walkResult
}

func main() {
	defer duration(time.Now(), "main")

	fmt.Println("==================================")
	fmt.Println("gosearch: A search in text utility written in Go.")
	fmt.Println("searching...")
	fmt.Println("==================================")

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

	files, err := walkFiles(inputDir, searchText)
	if err != nil {
		fmt.Println(err)
	}
	for file := range files {
		if file.found == true {
			fmt.Println(file)
		}
		if file.found == false && verbose {
			fmt.Println(file)
		}
	}
	// print summary
	fmt.Println("==================================")
	log.Println("")
	fmt.Printf("Done searching for %s\n", searchText)
	fmt.Printf("Path: %s\n", inputDir)
	fmt.Printf("Checked %d files in %d directories\n", fileVisit, folderVisit)
	fmt.Printf("Found %d files containing %s\n", numFound, searchText)
	fmt.Println("==================================")
}
