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

- `-k` : Keyword to search

- `-p` : Path to directory to search

- `-v` : Verbose (prints all files searched)


### Results:

The output of the utility includes:

- `path` - path provided by user

- `keyword` - search term provided by user 

- `files` - utility output of path to files whose contents match keyword

- `find count` - utility output with count of files whose contents match keyword

- `timestamp` - utility output of date and time of search

- `visited file count` - utility output with total count of files visited

- `visited folder count` - utility output with total count of folders visited

- `elapsed time` - utility output of time it took to search

If Verbose is selected by user:

- `files not matching` - utility output of path to files whose contents did not match keyword

- `read errors` - utility output of path to files that could not be read (e.g. file does not exist)  


Current version is 0.1.0
=========================

If you have any comments or feature requests please let me know.