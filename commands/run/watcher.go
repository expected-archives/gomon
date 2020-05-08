package run

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/expectedsh/gomon/pkg/utils"
)

type watcher struct {
	ctx context.Context

	mutex         sync.Mutex
	lastEvent     time.Time
	appsToRestart map[string]bool
}

func newWatcher(ctx context.Context) *watcher {
	return &watcher{
		ctx:           ctx,
		mutex:         sync.Mutex{},
		lastEvent:     time.Time{},
		appsToRestart: make(map[string]bool),
	}
}

func (w *watcher) watchForRestarts() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := w.prepareWatcher(watcher); err != nil {
		return err
	}

	go func() {
		go w.handleRestarts()

		for {
			select {
			case <-w.ctx.Done():
				watcher.Close()
				return
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				// ignore intellij temporary files
				if strings.HasSuffix(event.Name, "~") {
					continue
				}

				w.processWatchedEvent(watcher, event)
			case <-watcher.Errors:
				continue
			}
		}
	}()

	return nil
}

func (w *watcher) prepareWatcher(watcher *fsnotify.Watcher) error {
	for _, directory := range fDirectories {
		subDirs := []string{directory}
		utils.GetSubDirectories(directory, &subDirs)

		for _, subDir := range subDirs {
			if err := watcher.Add(subDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *watcher) handleRestarts() {
	for {
		w.mutex.Lock()
		if !w.lastEvent.IsZero() && time.Since(w.lastEvent) >= fWatchTimeout {
			w.lastEvent = time.Time{}
			for app := range w.appsToRestart {
				fmt.Printf("\nRestarting applicaion %q due to file change...\n\n", app)
				applications[app].restart <- true
			}
			w.appsToRestart = map[string]bool{}
		}
		w.mutex.Unlock()

		time.Sleep(time.Millisecond * 500)
		if w.ctx.Err() != nil {
			return
		}
	}
}

func (w *watcher) processWatchedEvent(watcher *fsnotify.Watcher, ev fsnotify.Event) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if ev.Op&fsnotify.Rename == fsnotify.Rename {
		for _, app := range applications {
			delete(app.files, ev.Name)
		}

		watcher.Remove(ev.Name)
		return
	}

	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		for _, app := range applications {
			delete(app.files, ev.Name)
		}

		watcher.Remove(ev.Name)
		return
	}

	if ev.Op&fsnotify.Create == fsnotify.Create {
		watcher.Add(ev.Name)
	}

	if ev.Op&fsnotify.Write == fsnotify.Write {
		w.lastEvent = time.Now()
		for _, app := range applications {
			if _, ok := app.files[ev.Name]; ok {
				app.updateFiles(ev.Name)
				w.appsToRestart[app.config.Name] = true
			}
		}

		return
	}
}
