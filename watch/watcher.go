package watch

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors D2 files and their imports for changes.
type Watcher struct {
	sourcePath    string
	fsWatcher     *fsnotify.Watcher
	watchedFiles  map[string]struct{}
	onChange      func() error
	debounceDelay time.Duration
	mu            sync.Mutex
	// done is a signal channel used to gracefully shut down the event loop goroutine.
	// When Stop() is called, this channel is closed, which unblocks the select in
	// eventLoop() and causes it to clean up and return.
	done chan struct{}
}

// WatcherOptions contains configuration for the Watcher.
type WatcherOptions struct {
	SourcePath    string        // Path to the main D2 source file
	OnChange      func() error  // Callback when files change
	DebounceDelay time.Duration // Delay before triggering onChange after last event
}

// NewWatcher creates a new Watcher for the given source file.
func NewWatcher(opts WatcherOptions) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	debounceDelay := opts.DebounceDelay
	if debounceDelay == 0 {
		debounceDelay = 100 * time.Millisecond
	}

	w := &Watcher{
		sourcePath:    opts.SourcePath,
		fsWatcher:     fsWatcher,
		watchedFiles:  make(map[string]struct{}),
		onChange:      opts.OnChange,
		debounceDelay: debounceDelay,
		done:          make(chan struct{}),
	}

	return w, nil
}

// Start begins watching the source file and its imports.
// It will call UpdateWatchedFiles to set up initial watches.
func (w *Watcher) Start() error {
	// Set up initial watches
	if err := w.UpdateWatchedFiles(); err != nil {
		return err
	}

	// Start the event loop
	go w.eventLoop()

	return nil
}

// Stop stops the watcher and closes all file watches.
func (w *Watcher) Stop() error {
	// Signal the event loop goroutine to stop by closing the done channel.
	// This causes the <-w.done case in eventLoop's select to unblock and return.
	close(w.done)
	return w.fsWatcher.Close()
}

// UpdateWatchedFiles reads the source file, extracts imports, and updates the watched files.
// This should be called after each successful compilation to pick up new imports.
func (w *Watcher) UpdateWatchedFiles() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Read the source file
	content, err := os.ReadFile(w.sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Extract imports
	imports := ExtractImports(string(content))
	importPaths := ResolveImportPaths(w.sourcePath, imports)

	// Build the new set of files to watch
	newFiles := make(map[string]struct{})
	newFiles[w.sourcePath] = struct{}{}
	for _, path := range importPaths {
		// Only add if file exists
		if _, err := os.Stat(path); err == nil {
			newFiles[path] = struct{}{}
		}
	}

	// Remove watches for files no longer needed
	for path := range w.watchedFiles {
		if _, exists := newFiles[path]; !exists {
			w.fsWatcher.Remove(path)
		}
	}

	// Add watches for new files
	for path := range newFiles {
		if _, exists := w.watchedFiles[path]; !exists {
			if err := w.fsWatcher.Add(path); err != nil {
				// Log but don't fail - the file might not exist yet
				fmt.Printf("Warning: could not watch %s: %v\n", path, err)
			}
		}
	}

	w.watchedFiles = newFiles
	return nil
}

// GetWatchedFiles returns a copy of the currently watched files.
func (w *Watcher) GetWatchedFiles() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	files := make([]string, 0, len(w.watchedFiles))
	for path := range w.watchedFiles {
		files = append(files, path)
	}
	return files
}

// eventLoop processes file system events and triggers onChange.
func (w *Watcher) eventLoop() {
	var debounceTimer *time.Timer
	var timerMu sync.Mutex

	for {
		select {
		// done channel is closed when Stop() is called, signaling graceful shutdown.
		// Clean up any pending debounce timer and exit the goroutine.
		case <-w.done:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			// Only react to write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce: reset timer on each event
			timerMu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(w.debounceDelay, func() {
				if w.onChange != nil {
					if err := w.onChange(); err != nil {
						fmt.Printf("Error during recompilation: %v\n", err)
					}
				}
				// Update watched files after compilation (imports may have changed)
				if err := w.UpdateWatchedFiles(); err != nil {
					fmt.Printf("Error updating watched files: %v\n", err)
				}
			})
			timerMu.Unlock()

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}
