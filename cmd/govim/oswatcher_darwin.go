// +build darwin

package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsevents"
)

const fRemoved = fsevents.ItemRemoved | fsevents.ItemRenamed
const fChanged = fsevents.ItemCreated | fsevents.ItemModified | fsevents.ItemChangeOwner

func newOSWatcher(gomodpath string) (*osWatcher, error) {
	dirpath := filepath.Dir(gomodpath)
	dev, err := fsevents.DeviceForPath(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve device for path %v: %v", dirpath, err)
	}

	es := &fsevents.EventStream{
		Paths:   []string{dirpath},
		Latency: 200 * time.Millisecond,
		Device:  dev,
		Flags:   fsevents.FileEvents | fsevents.WatchRoot,
	}

	es.Start()

	eventCh := make(chan fsEvent)
	go func() {
		for {
			events, ok := <-es.Events
			if !ok {
				break
			}
			for i := range events {
				event := events[i]
				path := filepath.Join("/", event.Path)
				switch {
				case event.Flags&fRemoved > 0:
					eventCh <- fsEvent{path, fsOpRemoved}
				case event.Flags&fChanged > 0:
					eventCh <- fsEvent{path, fsOpChanged}
				}
			}
		}
		close(eventCh)
	}()

	return &osWatcher{
		events: eventCh,
		errors: make(chan error),
		add:    func(path string) error { return nil },
		remove: func(path string) error { return nil },
		close: func() error {
			es.Stop()
			return nil
		},
	}, nil
}
