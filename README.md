# gosearch

An OS "search in text" utility written in Go. 

This utility will walk through the directory you specify, including any sub-folders, and return any files whose contents match the keyword you provide.

The search supports system files as well as user created files (e.g. pdf, doc, txt, ini, cfg, etc...). It also searches hidden directories and files.

### Installation:

Note that you **first** need to install  <a href="https://golang.org/" target="_blank">Go</a>

After you've installed Go, open your terminal (or command console) and type:
```
go get github.com/geriess/gosearch
```

The above command automatically fetches the source code and any dependencies, compile the binary and puts an executable binary in the $GOPATH/bin directory. The $GOPATH is the Go working folder that was configured when you installed Go.


### Usage:

Open your terminal (or command console) and type:
```
gosearch -p path -k keyword
```

**[OPTIONS]**

- `-k` : Keyword to search (required)

- `-p` : Path to directory to search (required)

- `-s` : Max file size to search in MB

- `-v` : Verbose prints all files searched

- `-j` : Output in JSON format

- `-h` : Print help menu


### Results:

The output of the utility includes:

- `path` - path provided by user

- `keyword` - search term provided by user 

- `files` - utility output of path to files whose contents match keyword

- `found files count` - utility output with count of files whose contents or name match keyword

- `found folder count` - utility output with count of folders whose name match keyword

- `timestamp` - utility output of date and time of search

- `visited file count` - utility output with total count of files visited

- `visited folder count` - utility output with total count of folders visited

- `elapsed time` - utility output of time it took to search

If Verbose is selected by user:

- `files not matching` - utility output of path to files whose contents did not match keyword

- `read errors` - utility output of path to files that could not be read (e.g. file too big)  



If you have any comments or feature requests please let me know.

## To-Do

Optimize with GOMACPROCS, see commentsl from thread below:

No, those are the "procs". In Go scheduler's lingo, there are three different concepts:
* "G": these are goroutines
* * "M": these are OS-level threads (the ones I was mentioning). These are bounded by debug.SetMaxThreads (default: 10000).
* * "P": these are basically locks that Ms must acquire to run Go code. These are bounded by GOMAXPROCS.
* So when you way "GOMAXPROCS=2", you're saying "I want at most two Gs executing code simultaneously", but you can still have a very high number of OS-level threads that are I/O blocked. Notice that when a G does a blocking I/O there are two possibilities: if it's a pollable operation, the handle is passed to the epoll/kqueue thread, the M releases its P and is recycled to do something else. If it's a non-pollable operation, the M releases the P but stays allocated to wait for the operation to finish (e.g.: the syscall to return). At that point, given that at least one P is free, any ready G can be scheduled, but a new M might be needed.
