package watcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Op represents filesystem operation type (write, rename, etc).
type Op = fsnotify.Op

// These are possible filesystem operation types.
const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

// Callback is a function that is invoked when a change is detected.
// The argument to the callback is a list of files, which have been changed.
// A callback should return an error in case it could not process the files.
type Callback func(string, Op) error

// Watcher watches a path for changes and invokes callbacks when a change is detected.
type Watcher struct {
	path      string
	fsWatcher *fsnotify.Watcher
	callbacks []Callback
	stopChan  chan struct{}
}

// New returns an instance of Watcher, that watches the given path.
func New(path string) (*Watcher, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		path:      path,
		fsWatcher: w,
		stopChan:  make(chan struct{}, 1),
	}, nil
}

// AddCallback adds new callback which will be invoked when a change is detected.
func (w *Watcher) AddCallback(cb Callback) {
	w.callbacks = append(w.callbacks, cb)
}

// Start starts watching for changes.
func (w *Watcher) Start() error {
	info, err := os.Stat(w.path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		err := filepath.Walk(w.path, func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			return w.fsWatcher.Add(path)
		})

		if err != nil {
			return err
		}
	} else {
		if err := w.fsWatcher.Add(w.path); err != nil {
			return err
		}
	}

	go w.watch()

	return nil
}

// Stop stops watching for changes.
func (w *Watcher) Stop() (err error) {
	err = w.fsWatcher.Close()
	close(w.stopChan)
	return
}

func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if ok {
				w.process(&event)
			} else {
				fmt.Println("failed to process fsnotify event")
			}
		case err := <-w.fsWatcher.Errors:
			if err != nil {
				fmt.Printf("error watching file changes %s\n", err.Error())
			}
		case <-w.stopChan:
			return
		}
	}
}

func (w *Watcher) process(event *fsnotify.Event) {
	for _, cb := range w.callbacks {
		if err := cb(event.Name, Op(event.Op)); err != nil {
			fmt.Printf("callback failed to process an operation %v, %s\n", cb, err)
		}
	}
}
