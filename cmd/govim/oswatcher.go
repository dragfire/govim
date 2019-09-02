// +build !darwin

package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

func newOSWatcher(gomodpath string) (*osWatcher, error) {
	mw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create new watcher: %v", err)
	}

	eventCh := make(chan fsEvent)
	go func() {
		for {
			e, ok := <-mw.Events
			if !ok {
				break
			}
			switch e.Op {
			case fsnotify.Rename, fsnotify.Remove:
				eventCh <- fsEvent{e.Name, fsOpRemoved}
			case fsnotify.Create, fsnotify.Chmod, fsnotify.Write:
				eventCh <- fsEvent{e.Name, fsOpChanged}
			}
		}
		close(eventCh)
	}()

	return &osWatcher{
		events: eventCh,
		errors: mw.Errors,
		add: func(path string) error {
			return mw.Add(path)
		},
		remove: func(path string) error {
			return mw.Remove(path)
		},
		close: func() error {
			return mw.Close()
		},
	}, nil
}
