package main

import (
	"fmt"
	"io"
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
	stopChannel  chan bool
)

func main() {
	// parse flags
	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 {
		printUsage()
	}

	// get current directory
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// Print Info
	PrintInfo(rootDir)

	// Exec command in first run
	greenPrint := color.New(color.FgGreen)
	greenPrint.Printf("[amon] starting `%s`\n", command)
	ExecCommand(command)

	// Exec other command if exist
	if len(strings.TrimSpace(otherCommand)) > 0 {
		ExecCommand(otherCommand)
	}

	// creates a new file watcher
	watcher, _ = fsnotify.NewWatcher()
	defer watcher.Close()

	// starting at the root of the project, walk each files
	// even in subdirectory
	err = WatchDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				// fmt.Printf("EVENT! %#v\n", event)

				// create events
				if event.Op&fsnotify.Create == fsnotify.Create {
					// watch files refresh
					err = WatchDir(rootDir)
					if err != nil {
						log.Fatal(err)
					}
				}

				// write events
				if event.Op&fsnotify.Write == fsnotify.Write {
					if !strings.Contains(event.Name, "/.") {
						greenPrint.Printf("\n[amon] restarting due to changes...`\n")
						greenPrint.Printf("[amon] starting `%s`\n", command)

						stopChannel <- true

						// Execute command
						ExecCommand(command)

						// Execute other command
						if len(strings.TrimSpace(otherCommand)) > 0 {
							ExecCommand(otherCommand)
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
	stopChannel = make(chan bool)

	flag.StringVarP(&command, "command", "c", "", "Command to run")
	flag.StringVarP(&otherCommand, "otherCommand", "o", "", "Other Command to run")
}

func printUsage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(1)
}

// ExecCommand : taken from fresh
func ExecCommand(command string) bool {
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

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	go func() {
		<-stopChannel
		pid := cmd.Process.Pid
		fmt.Printf("Killing PID %d\n", pid)
		cmd.Process.Kill()
	}()

	return true
}

// WatchDir : add directory and files to watcher
func WatchDir(dirPath string) error {
	var err error

	// Folder walk then ignore hidden folder
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, "/.") {
			watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return err
}

// PrintInfo : print info when start
func PrintInfo(rootDir string) {
	// Print version
	fmt.Printf("> %s\n", rootDir)
	cyanPrint := color.New(color.FgCyan)
	cyanPrint.Printf("[amon] 0.0.1\n")
	cyanPrint.Printf("[amon] to restart at any time, enter `rs`\n")
	cyanPrint.Printf("[amon] watching: *.*\n")
}
