package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/radiohead/gopass/internal/watcher"
)

// Return codes of the program.
const (
	CodeOK  = 0
	CodeErr = 1
)

func main() {
	var path string
	flag.StringVar(&path, "path", "", "path to watch the changes for")
	flag.Parse()

	if path == "" {
		flag.Usage()
		os.Exit(CodeErr)
	}

	w, err := watcher.New(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(CodeErr)
	}

	fmt.Printf("Starting to watch %s\n", path)

	w.AddCallback(func(file string, op watcher.Op) error {
		fmt.Printf("File %s has received operation %s!\n", file, op)

		return nil
	})

	osCh := make(chan os.Signal, 1)
	signal.Notify(osCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	if err := w.Start(); err != nil {
		fmt.Println(err)
		os.Exit(CodeErr)
	}

	fmt.Println("Watcher has started")

	<-osCh

	fmt.Println("Stopping the watcher")

	w.Stop()

	fmt.Println("Watcher stopped, exiting")
	os.Exit(CodeOK)
}
