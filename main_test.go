package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

func TestAstOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		destDir    string
		expected   string
	}{
		{
			name:       "simple filename",
			sourcePath: "test.d2",
			destDir:    "/output",
			expected:   "/output/test-ast.json",
		},
		{
			name:       "with directory",
			sourcePath: "/path/to/file.d2",
			destDir:    "/output",
			expected:   "/output/file-ast.json",
		},
		{
			name:       "deep path",
			sourcePath: "/very/deep/nested/path/to/diagram.d2",
			destDir:    "/out",
			expected:   "/out/diagram-ast.json",
		},
		{
			name:       "relative path",
			sourcePath: "relative/path/file.d2",
			destDir:    "./output",
			expected:   "./output/file-ast.json",
		},
		{
			name:       "no extension",
			sourcePath: "noextension",
			destDir:    "/output",
			expected:   "/output/noextension-ast.json",
		},
		{
			name:       "multiple extensions",
			sourcePath: "file.backup.d2",
			destDir:    "/output",
			expected:   "/output/file.backup-ast.json",
		},
		{
			name:       "hidden file",
			sourcePath: ".hidden.d2",
			destDir:    "/output",
			expected:   "/output/.hidden-ast.json",
		},
		{
			name:       "windows style path",
			sourcePath: "C:\\Users\\test\\file.d2",
			destDir:    "/output",
			expected:   "/output/file-ast.json",
		},
		{
			name:       "mixed separators",
			sourcePath: "path/to\\mixed/file.d2",
			destDir:    "/output",
			expected:   "/output/file-ast.json",
		},
		{
			name:       "trailing slash in dest",
			sourcePath: "file.d2",
			destDir:    "/output/",
			expected:   "/output//file-ast.json",
		},
		{
			name:       "dot in directory",
			sourcePath: "/path.with.dots/file.d2",
			destDir:    "/output",
			expected:   "/output/file-ast.json",
		},
		{
			name:       "only extension",
			sourcePath: ".d2",
			destDir:    "/output",
			expected:   "/output/-ast.json",
		},
		{
			name:       "empty destination",
			sourcePath: "file.d2",
			destDir:    "",
			expected:   "/file-ast.json",
		},
		{
			name:       "different extension",
			sourcePath: "diagram.txt",
			destDir:    "/out",
			expected:   "/out/diagram-ast.json",
		},
		{
			name:       "long filename",
			sourcePath: "this-is-a-very-long-filename-for-testing.d2",
			destDir:    "/output",
			expected:   "/output/this-is-a-very-long-filename-for-testing-ast.json",
		},
		{
			name:       "current dir destination",
			sourcePath: "file.d2",
			destDir:    ".",
			expected:   "./file-ast.json",
		},
		{
			name:       "parent dir destination",
			sourcePath: "file.d2",
			destDir:    "..",
			expected:   "../file-ast.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := astOutputPath(tt.sourcePath, tt.destDir)
			if result != tt.expected {
				t.Errorf("astOutputPath(%q, %q) = %q, want %q",
					tt.sourcePath, tt.destDir, result, tt.expected)
			}
		})
	}
}

func TestIntegration_ReadAndParseTestData(t *testing.T) {
	testFiles := []struct {
		filename      string
		expectedViews int
	}{
		{"simple.d2", 1},
		{"basic.d2", 3},
		{"no_views.d2", 0},
	}

	for _, tf := range testFiles {
		t.Run(tf.filename, func(t *testing.T) {
			path := filepath.Join("test_data", tf.filename)

			content, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("test file not found: %s", path)
				return
			}

			reader := bytes.NewReader(content)
			d2map, err := parseD2(path, reader)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", tf.filename, err)
			}

			views := getViewsNodes(d2map)
			if len(views) != tf.expectedViews {
				t.Errorf("expected %d views in %s, got %d",
					tf.expectedViews, tf.filename, len(views))
			}
		})
	}
}
