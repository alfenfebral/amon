package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

//
var watcher *fsnotify.Watcher

// main
func main() {
	// Print version
	cyanPrint := color.New(color.FgCyan)
	cyanPrint.Printf("[amon] 0.0.1\n")
	cyanPrint.Printf("[amon] to restart at any time, enter `rs`\n")
	cyanPrint.Printf("[amon] watching: *.*\n")

	// Exec command in first run
	greenPrint := color.New(color.FgGreen)
	greenPrint.Printf("[amon] starting `go run main.go`\n")
	execCommand()

	// get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(dir+"/test", watchDir); err != nil {
		fmt.Println("ERROR", err)
	}

	//
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				// fmt.Printf("EVENT! %#v\n", event)

				// create events
				if event.Op&fsnotify.Create == fsnotify.Create {
					// wathdir refresh when create new folder
					if err := filepath.Walk(dir, watchDir); err != nil {
						fmt.Println("ERROR", err)
					}
				}

				// write events
				if event.Op&fsnotify.Write == fsnotify.Write {
					greenPrint.Printf("[amon] restarting due to changes...`\n")
					greenPrint.Printf("[amon] starting `go run main.go`\n")

					// Execute command
					execCommand()
				}

				// remove events
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					// Do ...
				}

				// rename events
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					// Do ...
				}

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done
}

func execCommand() {
	cmd := exec.Command("go", "run", "test/main.go")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// print stdout
	if len(outb.String()) > 0 {
		fmt.Printf("%s\n", outb.String())
	}
	// print stderr
	if len(errb.String()) > 0 {
		fmt.Printf("%s\n", errb.String())
	}
}

func isFileOrDir(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		fmt.Println("directory")
	case mode.IsRegular():
		// do file stuff
		fmt.Println("file")
	}
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}
