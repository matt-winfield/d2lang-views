package watch

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_UpdateWatchedFiles(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.d2")
	importPath := filepath.Join(tempDir, "icons.d2")

	// Create source file with import
	sourceContent := `client: "Web Client"
icons: @icons`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create imported file
	if err := os.WriteFile(importPath, []byte(`icon: "test"`), 0644); err != nil {
		t.Fatalf("failed to write import file: %v", err)
	}

	// Create watcher
	watcher, err := NewWatcher(WatcherOptions{
		SourcePath: sourcePath,
		OnChange:   func() error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Update watched files
	if err := watcher.UpdateWatchedFiles(); err != nil {
		t.Fatalf("failed to update watched files: %v", err)
	}

	// Check that both files are being watched
	watchedFiles := watcher.GetWatchedFiles()
	if len(watchedFiles) != 2 {
		t.Errorf("expected 2 watched files, got %d: %v", len(watchedFiles), watchedFiles)
	}

	// Verify the specific files
	watchedSet := make(map[string]struct{})
	for _, f := range watchedFiles {
		watchedSet[f] = struct{}{}
	}
	if _, ok := watchedSet[sourcePath]; !ok {
		t.Errorf("source file not in watched files")
	}
	if _, ok := watchedSet[importPath]; !ok {
		t.Errorf("import file not in watched files")
	}
}

func TestWatcher_UpdateWatchedFiles_MissingImport(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.d2")

	// Create source file with import to non-existent file
	sourceContent := `client: "Web Client"
icons: @missing`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create watcher
	watcher, err := NewWatcher(WatcherOptions{
		SourcePath: sourcePath,
		OnChange:   func() error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Update watched files - should not fail even with missing import
	if err := watcher.UpdateWatchedFiles(); err != nil {
		t.Fatalf("failed to update watched files: %v", err)
	}

	// Check that only source file is being watched
	watchedFiles := watcher.GetWatchedFiles()
	if len(watchedFiles) != 1 {
		t.Errorf("expected 1 watched file, got %d: %v", len(watchedFiles), watchedFiles)
	}
}

func TestWatcher_DetectsFileChanges(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.d2")

	// Create source file
	sourceContent := `client: "Web Client"`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Track onChange calls
	var changeCount int32

	// Create watcher with short debounce
	watcher, err := NewWatcher(WatcherOptions{
		SourcePath: sourcePath,
		OnChange: func() error {
			atomic.AddInt32(&changeCount, 1)
			return nil
		},
		DebounceDelay: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start the watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Give the watcher time to set up
	time.Sleep(50 * time.Millisecond)

	// Modify the file
	newContent := `client: "Updated Client"`
	if err := os.WriteFile(sourcePath, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to modify source file: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(200 * time.Millisecond)

	// Check that onChange was called
	count := atomic.LoadInt32(&changeCount)
	if count < 1 {
		t.Errorf("expected onChange to be called at least once, got %d calls", count)
	}
}

func TestWatcher_UpdatesImportWatches(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.d2")
	newImportPath := filepath.Join(tempDir, "styles.d2")

	// Create source file without imports
	sourceContent := `client: "Web Client"`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create watcher
	watcher, err := NewWatcher(WatcherOptions{
		SourcePath: sourcePath,
		OnChange:   func() error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Initial update
	if err := watcher.UpdateWatchedFiles(); err != nil {
		t.Fatalf("failed to update watched files: %v", err)
	}

	// Should only watch source file
	if len(watcher.GetWatchedFiles()) != 1 {
		t.Errorf("expected 1 watched file initially")
	}

	// Create new import file
	if err := os.WriteFile(newImportPath, []byte(`style: "dark"`), 0644); err != nil {
		t.Fatalf("failed to write import file: %v", err)
	}

	// Update source to include import
	newSourceContent := `client: "Web Client"
styles: @styles`
	if err := os.WriteFile(sourcePath, []byte(newSourceContent), 0644); err != nil {
		t.Fatalf("failed to update source file: %v", err)
	}

	// Update watched files
	if err := watcher.UpdateWatchedFiles(); err != nil {
		t.Fatalf("failed to update watched files after adding import: %v", err)
	}

	// Should now watch both files
	watchedFiles := watcher.GetWatchedFiles()
	if len(watchedFiles) != 2 {
		t.Errorf("expected 2 watched files after adding import, got %d: %v", len(watchedFiles), watchedFiles)
	}
}
