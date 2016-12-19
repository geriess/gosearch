package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
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
	json        bool           // output in json if true
	help        bool           // display help if true
)

// walkresult struct for result document
type walkresult struct {
	path    string
	name    string
	found   bool
	isDir   bool
	size    int64
	modTime time.Time
}

func usage() {
	// user messaging
	fmt.Println("==================================")
	fmt.Println("gosearch: A search-in-text utility written in Go.")
	fmt.Println("==================================")
	fmt.Println("Usage:")
	fmt.Println("    gosearch [OPTIONS] -p path -k keyword")
	flag.PrintDefaults()
}

func init() {
	// flag init
	flag.StringVar(&inputDir, "p", "", "Path of directory to search")
	flag.StringVar(&searchText, "k", "", "Keyword to search")
	flag.Int64Var(&maxSize, "s", 100, "Max file size to search in MB - optional")
	flag.BoolVar(&json, "j", false, "Output in JSON - optional")
	flag.BoolVar(&verbose, "v", false, "Verbose = optional (prints all files searched)")
	flag.BoolVar(&help, "h", false, "Print help menu")
}

// duration keeps track of function elapsed time
func duration(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("func %s elapsed %s\n", name, elapsed)
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
				if f.Size() < maxSize*1024*1024 {
					wg.Add(1)
					go readFile(path, f, filesFound)
				} else {
					log.WithFields(log.Fields{
						"type": "file",
						"name": f.Name(),
						"path": path,
					}).Warn("Skip file too large: ", f.Size())
				}
			}

			// folder path, increment count
			folderCount()
			wg.Add(1)
			go searchPath(path, f, filesFound)
			return nil
		})

		// launch cleanup, but sync wait until goroutines complete
		go cleanup(filesFound, done)

		// check errors for walk func
		errorCheck(err)
		return
	}()
	return
}

// readFile puts contents of file in memory, starts search
func readFile(path string, f os.FileInfo, filesFound chan walkresult) {
	defer wg.Done()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if !verbose {
			return
		}
		log.WithFields(log.Fields{
			"type": "file",
			"name": f.Name(),
			"path": path,
		}).Warn("File cannot be read", f.Size())
		return
	}
	wg.Add(1)
	go searchFile(path, content, f, filesFound)
}

// searchFile parses the contents of file looking for keyword
func searchFile(path string, content []byte, f os.FileInfo, filesFound chan walkresult) {
	defer wg.Done()
	x := string(content)
	search := strings.Contains(x, searchText)
	switch search {
	case true:
		lock.Lock()
		numFound++
		lock.Unlock()
		found := true
		filesFound <- walkresult{path, f.Name(), found, f.IsDir(), f.Size(), f.ModTime()}
		return
	case false:
		found := false
		filesFound <- walkresult{path, f.Name(), found, f.IsDir(), f.Size(), f.ModTime()}
		return
	}
}

// searchPath searches match in file or folder name
func searchPath(path string, f os.FileInfo, filesFound chan walkresult) {
	defer wg.Done()
	search := strings.Contains(f.Name(), searchText)
	switch search {
	case true:
		if f.IsDir() {
			lock.Lock()
			dirFound++
			lock.Unlock()
		} else {
			lock.Lock()
			numFound++
			lock.Unlock()
		}
		found := true
		filesFound <- walkresult{path, f.Name(), found, f.IsDir(), f.Size(), f.ModTime()}
		return
	case false:
		found := false
		filesFound <- walkresult{path, f.Name(), found, f.IsDir(), f.Size(), f.ModTime()}
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

// waits for goroutines to complete, sets done signal and closes channels
func cleanup(filesFound chan walkresult, done chan bool) {
	wg.Wait()
	close(filesFound)
	done <- true
	<-done
	close(done)
	return
}

// summary prints results, counts, lets user know search is done
func summary(searchText string, path string) {
	log.WithFields(log.Fields{
		"searchString":   searchText,  // text to search
		"path":           path,        // file path requeted to search
		"filesChecked":   fileVisit,   // num of files visited during search
		"foldersChecked": folderVisit, // num of folders visited during search
		"filesFound":     numFound,    // num of files that contain match for search string
		"foldersFound":   dirFound,    // num of folders that contain match for search string
	}).Info("Search completed")
}

func main() {
	// main timer
	defer duration(time.Now(), "main")

	// check args provided
	flag.Parse()
	ok := true

	if help == true {
		usage()
		os.Exit(1)
	}
	if inputDir == "" {
		ok = errorOut("ERROR: Missing path to directory")
	} else {
		// check path exists
		verify := exists(inputDir)
		if !verify {
			ok = errorOut("ERROR: Path provided does not exist.")
		}
	}
	if searchText == "" {
		ok = errorOut("ERROR: Missing keyword to search")
	}

	if !ok {
		usage()
		os.Exit(1)
	}

	// log set to JSON format
	if json == true {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}

	// create channels
	filesFound := make(chan walkresult)
	done := make(chan bool)

	// notify user search started
	log.WithFields(log.Fields{
		"searchString": searchText,
		"path":         inputDir,
	}).Info("Search started")

	// start search work
	go walkFiles(inputDir, searchText, filesFound, done)

	// receive channel results and print
loop:
	for {
		select {
		case print := <-filesFound:
			if (len(print.path) > 0) && verbose && (print.found == false) {
				switch print.isDir {
				case true:
					log.WithFields(log.Fields{
						"type": "folder",
						"name": print.name,
						"path": print.path,
					}).Info("Match not found")
				case false:
					log.WithFields(log.Fields{
						"type": "file",
						"name": print.name,
						"path": print.path,
					}).Info("Match not found")
				}
			}
			if print.found == true {
				switch print.isDir {
				case true:
					log.WithFields(log.Fields{
						"type": "folder",
						"name": print.name,
						"path": print.path,
					}).Info("Match found")
				case false:
					log.WithFields(log.Fields{
						"type": "file",
						"name": print.name,
						"path": print.path,
					}).Info("Match found")
				}

			}
		case <-done:
			done <- true
			break loop
		}
	}

	// print search summary, file counts
	summary(searchText, inputDir)
}
