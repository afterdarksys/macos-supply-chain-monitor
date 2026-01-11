package watcher

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	*fsnotify.Watcher
	path string
}

func New(path string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := fsWatcher.Add(path); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to watch path %s: %w", path, err)
	}

	return &Watcher{
		Watcher: fsWatcher,
		path:    path,
	}, nil
}

func (w *Watcher) Path() string {
	return w.path
}
