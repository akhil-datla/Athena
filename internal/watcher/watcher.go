package watcher

import (
	"log/slog"

	"github.com/akhil-datla/athena/internal/engine"
	"github.com/akhil-datla/athena/internal/loader"
	"github.com/fsnotify/fsnotify"
)

// Watch monitors a file for changes and reloads the engine when modified.
// It blocks until the done channel is closed.
func Watch(path string, eng *engine.Engine, logger *slog.Logger, done <-chan struct{}) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("failed to create file watcher", "error", err)
		return
	}
	defer w.Close()

	if err := w.Add(path); err != nil {
		logger.Error("failed to watch file", "path", path, "error", err)
		return
	}

	logger.Info("watching file for changes", "path", path)

	for {
		select {
		case <-done:
			return
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				logger.Info("file changed, reloading", "path", path)
				entries, err := loader.Load(path)
				if err != nil {
					logger.Error("failed to reload file", "error", err)
					continue
				}
				eng.Reload(entries)
				logger.Info("reloaded entries", "count", eng.Count())
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			logger.Error("watcher error", "error", err)
		}
	}
}
