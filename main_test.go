package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/matt-winfield/d2lang-views/compile"
)

func TestEnsureDirExists_CreatesNewDir(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "newdir")

	err := ensureDirExists(newDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected path to be a directory")
	}
}

func TestEnsureDirExists_ExistingDir(t *testing.T) {
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")
	err := os.Mkdir(existingDir, 0755)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = ensureDirExists(existingDir)

	if err != nil {
		t.Fatalf("expected no error for existing directory, got: %v", err)
	}
}

func TestEnsureDirExists_NestedDirs(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "a", "b", "c", "d")

	err := ensureDirExists(nestedDir)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("expected nested directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected path to be a directory")
	}
}

func TestEnsureDirExists_PartiallyExistingPath(t *testing.T) {
	tempDir := t.TempDir()
	existingPart := filepath.Join(tempDir, "existing")
	err := os.Mkdir(existingPart, 0755)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newPart := filepath.Join(existingPart, "new", "subdir")

	err = ensureDirExists(newPart)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	info, err := os.Stat(newPart)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected path to be a directory")
	}
}

func TestEnsureDirExists_FileExistsAtPath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "afile")

	// Create a file at the path
	err := os.WriteFile(filePath, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Try to ensure dir exists where file already exists
	err = ensureDirExists(filePath)

	// Should return an error when a file exists at the expected directory path
	if err == nil {
		t.Fatal("expected error when file exists at directory path, got nil")
	}
}

func TestCheckOutputConflict(t *testing.T) {
	tests := []struct {
		name        string
		sourcePath  string
		destination string
		expectError bool
	}{
		{
			name:        "no conflict - different directories",
			sourcePath:  "/docs/d2/views.d2",
			destination: "/output/diagram.svg",
			expectError: false,
		},
		{
			name:        "no conflict - output is file alongside source dir",
			sourcePath:  "/docs/d2/views.d2",
			destination: "/docs/diagram.svg",
			expectError: false,
		},
		{
			name:        "CONFLICT - output dir same as source dir",
			sourcePath:  "/docs/architecture/d2/views.d2",
			destination: "/docs/architecture/d2.svg",
			expectError: true,
		},
		{
			name:        "CONFLICT - output dir is parent of source",
			sourcePath:  "/docs/architecture/d2/subdir/views.d2",
			destination: "/docs/architecture/d2.svg",
			expectError: true,
		},
		{
			name:        "no conflict - similar names but different paths",
			sourcePath:  "/docs/d2-source/views.d2",
			destination: "/docs/d2.svg",
			expectError: false,
		},
		{
			name:        "no conflict - output in completely different tree",
			sourcePath:  "/project/src/diagrams/views.d2",
			destination: "/project/output/diagrams.svg",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOutputConflict(tt.sourcePath, tt.destination)
			if tt.expectError && err == nil {
				t.Errorf("expected error for source=%q dest=%q, got nil", tt.sourcePath, tt.destination)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for source=%q dest=%q, got: %v", tt.sourcePath, tt.destination, err)
			}
		})
	}
}

func TestGetViewLayerNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "single view layer",
			input: `client: "Web Client"
server: "API Server"

client -> server

layers: {
    frontend: { #view
        client
        server
    }
}`,
			expected: []string{"frontend"},
		},
		{
			name: "multiple view layers",
			input: `client: "Web Client"
server: "API Server"
db: "Database"

client -> server
server -> db

layers: {
    frontend: { #view
        client
        server
    }
    backend: { #view
        server
        db
    }
}`,
			expected: []string{"frontend", "backend"},
		},
		{
			name: "mixed view and non-view layers",
			input: `client: "Web Client"
server: "API Server"

client -> server

layers: {
    frontend: { #view
        client
        server
    }
    regular_layer: {
        other: "Something"
    }
    backend: { #view
        server
    }
}`,
			expected: []string{"frontend", "backend"},
		},
		{
			name:     "no layers",
			input:    `client: "Web Client"`,
			expected: []string{},
		},
		{
			name: "layers but no views",
			input: `client: "Web Client"

layers: {
    regular: {
        other: "Something"
    }
}`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			graph, _, err := compile.CompileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("failed to compile D2: %v", err)
			}

			result := getViewLayerNames(graph)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d view layers, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("view layer[%d] = %q, want %q", i, name, tt.expected[i])
				}
			}
		})
	}
}

func TestGetD2OutputPath(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		expected   string
	}{
		{
			name:       "source with directory",
			sourcePath: "/path/to/source.d2",
			expected:   "/path/to/source-compiled.d2",
		},
		{
			name:       "source in current directory",
			sourcePath: "source.d2",
			expected:   "source-compiled.d2",
		},
		{
			name:       "deep nested source path",
			sourcePath: "/very/deep/nested/path/diagram.d2",
			expected:   "/very/deep/nested/path/diagram-compiled.d2",
		},
		{
			name:       "relative source path",
			sourcePath: "./relative/path/source.d2",
			expected:   "./relative/path/source-compiled.d2",
		},
		{
			name:       "source without extension",
			sourcePath: "/path/to/source",
			expected:   "/path/to/source-compiled",
		},
		{
			name:       "source with multiple dots",
			sourcePath: "/path/to/my.diagram.d2",
			expected:   "/path/to/my.diagram-compiled.d2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getD2OutputPath(tt.sourcePath)
			if result != tt.expected {
				t.Errorf("getD2OutputPath(%q) = %q, want %q",
					tt.sourcePath, result, tt.expected)
			}
		})
	}
}
