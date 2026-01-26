package watcher

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors spec directories for file changes.
type Watcher struct {
	fsw      *fsnotify.Watcher
	specDirs []string
	Events   chan Event
}

// Event represents a file change event.
type Event struct {
	Type string `json:"type"` // "create", "modify", "delete"
	Path string `json:"path"`
}

// New creates a Watcher for the given spec directories.
func New(specDirs []string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fsw:      fsw,
		specDirs: specDirs,
		Events:   make(chan Event, 100),
	}

	for _, dir := range specDirs {
		if err := w.addRecursive(dir); err != nil {
			log.Printf("warning: could not watch %s: %v", dir, err)
		}
	}

	return w, nil
}

// Run starts processing filesystem events. Blocks until closed.
func (w *Watcher) Run() {
	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	close(w.Events)
	return w.fsw.Close()
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	var eventType string
	switch {
	case event.Op&fsnotify.Create != 0:
		eventType = "create"
	case event.Op&fsnotify.Write != 0:
		eventType = "modify"
	case event.Op&fsnotify.Remove != 0:
		eventType = "delete"
	default:
		return
	}

	w.Events <- Event{
		Type: eventType,
		Path: event.Name,
	}
}

func (w *Watcher) addRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.fsw.Add(path)
		}
		return nil
	})
}
