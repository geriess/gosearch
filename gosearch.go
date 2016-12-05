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
	dirFound    int            // # of directories matching keyword
	folderVisit int            // # of folders visited by search
	wg          sync.WaitGroup // sync goroutines / channels
	lock        sync.Mutex     // control access to counters (race prevention)
	maxSize     int64          // max file size
)

type walkresult struct {
	path  string
	name  string
	found bool
	isDir bool
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("    gosearch [OPTIONS] -p path -k keyword")
	flag.PrintDefaults()
}

func init() {
	maxSize = 512000000
	flag.StringVar(&inputDir, "p", "", "Path to directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
	flag.BoolVar(&verbose, "v", false, "Verbose (prints all files searched)")
}

// duration keeps track of functions elapsed time
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

// walkFiles walks all files and sub-directory paths
func walkFiles(directory string, keyword string, filesFound chan walkresult, done chan bool) {

	// launch goroutine to walk path; add wait count
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
			errorCheck(err)

			// if file launch main search process
			if !f.IsDir() {
				fileCount()

				// only launch search if file is under size limit,
				if f.Size() < maxSize {
					wg.Add(1)
					go readFile(path, f.IsDir(), f.Name(), filesFound)
				} else {
					fmt.Printf("%s skipped. File too large.", path)
				}
			}

			// folder path, increment count
			folderCount()
			wg.Add(1)
			go searchPath(path, f.Name(), f.IsDir(), filesFound)
			return nil
		})

		// launch cleanup, but sync wait until goroutines complete
		go cleanup(filesFound, done)

		// check errors for walk func
		errorCheck(err)
		fmt.Println("Close walk goroutine")
		return
	}()
	return
}

// readFile puts contents of file in memory
func readFile(path string, isDir bool, name string, filesFound chan walkresult) {
	defer wg.Done()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if !verbose {
			return
		}
		fmt.Printf("%s FILE cannot be read\n", path)
		return
	}
	wg.Add(1)
	go searchFile(path, content, isDir, name, filesFound)
}

// searchFile parses the contents of file looking for keyword
func searchFile(path string, content []byte, isDir bool, name string, filesFound chan walkresult) {
	defer wg.Done()
	x := string(content)
	search := strings.Contains(x, searchText)
	switch search {
	case true:
		lock.Lock()
		numFound++
		lock.Unlock()
		found := true
		filesFound <- walkresult{path, name, found, isDir}
		return
	case false:
		found := false
		filesFound <- walkresult{path, name, found, isDir}
		return
	}
}

// searchPath searches match in file or folder name
func searchPath(path string, name string, isDir bool, filesFound chan walkresult) {
	defer wg.Done()
	search := strings.Contains(name, searchText)
	switch search {
	case true:
		if isDir {
			lock.Lock()
			dirFound++
			lock.Unlock()
		} else {
			lock.Lock()
			numFound++
			lock.Unlock()
		}
		found := true
		filesFound <- walkresult{path, name, found, isDir}
		return
	case false:
		found := false
		filesFound <- walkresult{path, name, found, isDir}
		return
	}
}

// folderCount keeps count of folders visited during search
func folderCount() {
	lock.Lock()
	folderVisit++
	lock.Unlock()
}

// fileCount keeps count of files visited during search
func fileCount() {
	lock.Lock()
	fileVisit++
	lock.Unlock()
}

func cleanup(filesFound chan walkresult, done chan bool) {
	wg.Wait()
	log.Println("Performing cleanup.")
	close(filesFound)
	done <- true
	<-done
	close(done)
	return
}

// summary prints results, counts, lets user know search is done
func summary() {
	fmt.Println("==================================")
	fmt.Printf("Done searching for %s\n", searchText)
	fmt.Printf("Path: %s\n", inputDir)
	fmt.Printf("Checked %d files in %d folders\n", fileVisit, folderVisit)
	fmt.Printf("Found %d files containing %s\n", numFound, searchText)
	fmt.Printf("Found %d folders containing %s\n", dirFound, searchText)
	fmt.Println("==================================")
}

func main() {
	// main timer
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

	// start search work
	go walkFiles(inputDir, searchText, filesFound, done)

	// receive channel results and print
loop:
	for {
		select {
		case print := <-filesFound:
			if (len(print.path) > 0) && verbose && (print.found == false) {
				fmt.Printf("%s does NOT contain %s\n", print.path, searchText)
			}
			if print.found == true {
				switch print.isDir {
				case true:
					fmt.Printf("%s folder contains %s\n", print.path, searchText)
				case false:
					fmt.Printf("%s file contains %s\n", print.path, searchText)
				}

			}
		case <-done:
			fmt.Println("==================================")
			log.Println("Search complete.")
			done <- true
			break loop
		}
	}

	// print search summary, file counts
	summary()
}
