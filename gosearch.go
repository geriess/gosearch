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

// timer
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

// check path exists
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
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

type walkresult struct {
	path  string
	found bool
}

func walkFiles(directory string, keyword string, filesFound chan walkresult, done chan bool) {

	// launch goroutine to walk path; add wait count
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
			errorCheck(err)
			// launch search process for files only
			if !f.IsDir() {
				fileVisit++

				// launch goroutine to read file; add wait count
				wg.Add(1)
				go func(path string) {
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

					// search file contents for keyword
					x := string(content)
					search := strings.Contains(x, keyword)

					switch search {
					case true:
						numFound++
						found := true
						filesFound <- walkresult{path, found}
						return
					case false:
						found := false
						filesFound <- walkresult{path, found}
						return
					}
				}(path)
			}
			folderVisit++
			return nil
		})
		// launch cleanup, but wait for goroutines to complete
		go func() {
			wg.Wait()
			close(filesFound)
			done <- true
			<-done
			fmt.Println("Close cleanup goroutine")
			return
		}()
		// check errors for walk path
		errorCheck(err)
		fmt.Println("Close walk goroutine")
		return
	}()
	return
}

func main() {
	// timer
	defer duration(time.Now(), "main")

	// user messaging
	fmt.Println("==================================")
	fmt.Println("gosearch: A search in text utility written in Go.")
	fmt.Println("searching...")
	fmt.Println("==================================")

	// check args provided
	flag.Parse()
	ok := true
	if inputDir == "" {
		ok = errorOut("ERROR: Missing path to directory")
	}
	if searchText == "" {
		ok = errorOut("ERROR: Missing keyword to search")
	}

	// check path exists
	verify := exists(inputDir)
	if !verify {
		ok = errorOut("ERROR: Path provided does not exist.")
	}
	if !ok {
		usage()
		os.Exit(1)
	}

	// create channels
	filesFound := make(chan walkresult)
	done := make(chan bool)

	// start work
	go walkFiles(inputDir, searchText, filesFound, done)

	// loop through results channel and print
work:
	for {
		select {
		case print := <-filesFound:
			if verbose && (print.found == false) {
				fmt.Printf("%s does NOT contain %s\n", print.path, searchText)
			}
			if print.found == true {
				fmt.Printf("%s contains %s\n", print.path, searchText)
			}
		case <-done:
			fmt.Println("goroutines done. Closing work loop.")
			done <- true
			break work
		}
	}

	// print search summary, file counts
	fmt.Println("==================================")
	log.Println("")
	fmt.Printf("Done searching for %s\n", searchText)
	fmt.Printf("Path: %s\n", inputDir)
	fmt.Printf("Checked %d files in %d directories\n", fileVisit, folderVisit)
	fmt.Printf("Found %d files containing %s\n", numFound, searchText)
	fmt.Println("==================================")
}
