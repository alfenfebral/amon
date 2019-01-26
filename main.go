package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	flag "github.com/ogier/pflag"
)

var (
	watcher      *fsnotify.Watcher
	command      string
	otherCommand string
	rootDir      string
)

func main() {
	// parse flags
	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 {
		printUsage()
	}

	// Print version
	cyanPrint := color.New(color.FgCyan)
	cyanPrint.Printf("[amon] 0.0.1\n")
	cyanPrint.Printf("[amon] to restart at any time, enter `rs`\n")
	cyanPrint.Printf("[amon] watching: *.*\n")

	// Exec command in first run
	greenPrint := color.New(color.FgGreen)
	greenPrint.Printf("[amon] starting `go run main.go`\n")
	execCommand(command)

	// Exec other command if exist
	if len(strings.TrimSpace(otherCommand)) > 0 {
		execCommand(otherCommand)
	}

	// get current directory
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each files
	// even in subdirectory
	err = watchDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				fmt.Printf("EVENT! %#v\n", event)

				// create events
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("New Create")
					// watch files refresh
					err = watchDir(rootDir)
					if err != nil {
						log.Fatal(err)
					}
				}

				// write events
				if event.Op&fsnotify.Write == fsnotify.Write {
					if !strings.Contains(event.Name, "/.") {
						greenPrint.Printf("[amon] restarting due to changes...`\n")
						greenPrint.Printf("[amon] starting `go run main.go`\n")

						// Execute command
						execCommand(command)

						// Execute other command
						if len(strings.TrimSpace(otherCommand)) > 0 {
							execCommand(otherCommand)
						}
					}
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

func init() {
	flag.StringVarP(&command, "command", "c", "", "Command to run")
	flag.StringVarP(&otherCommand, "otherCommand", "o", "", "Other Command to run")
}

func printUsage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(1)
}

func execCommand(command string) {
	var cmd *exec.Cmd

	// Split command by whitespace
	commands := strings.Split(command, " ")

	if len(commands) > 1 {
		var command = commands[0]
		commands := commands[1:len(commands)]

		cmd = exec.Command(command, commands...)
	} else {
		cmd = exec.Command(command)
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	// print stdout
	if len(strings.TrimSpace(outb.String())) > 0 {
		fmt.Printf("%s\n", outb.String())
	}

	// print stderr
	if len(strings.TrimSpace(errb.String())) > 0 {
		fmt.Printf("%s\n", errb.String())
	}
}

func watchDir(dirPath string) error {
	var err error

	// Folder walk then ignore hidden folder
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, "/.") {
			fmt.Println(path)
			watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return err
}
