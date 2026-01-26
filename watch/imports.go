package watch

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// importPattern matches D2 import syntax:
// - Regular imports: @filename or @path/to/file
// - Spread imports: ...@filename
// - Quoted imports: @"filename-with-dots"
// - Partial imports: @file.subpath (only captures 'file' part)
var importPattern = regexp.MustCompile(`@("([^"]+)"|([a-zA-Z0-9_./\-]+))`)

// extractImports extracts all import file paths from D2 content.
// Returns a deduplicated list of import paths with .d2 extension added.
func extractImports(content string) []string {
	seen := make(map[string]struct{})
	imports := []string{}

	// Process line by line to avoid matching imports inside strings
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip if the @ is inside a string (simple heuristic: count quotes before @)
		atIdx := strings.Index(line, "@")
		if atIdx == -1 {
			continue
		}

		// Check if @ is inside a string by counting unescaped quotes before it
		beforeAt := line[:atIdx]
		quoteCount := strings.Count(beforeAt, `"`) - strings.Count(beforeAt, `\"`)
		if quoteCount%2 == 1 {
			// @ is inside a string, skip this line
			continue
		}

		matches := importPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			var importPath string
			isQuoted := false
			if match[2] != "" {
				// Quoted import: @"filename"
				importPath = match[2]
				isQuoted = true
			} else if match[3] != "" {
				// Unquoted import: @filename or @path/file
				importPath = match[3]
			}

			if importPath == "" {
				continue
			}

			// Handle partial imports: @file.subpath -> file
			// But preserve path separators: @path/file.subpath -> path/file
			// Skip this for quoted imports - they are literal filenames
			if !isQuoted {
				importPath = extractBaseImport(importPath)
			}

			// Add .d2 extension if not present
			if !strings.HasSuffix(importPath, ".d2") {
				importPath = importPath + ".d2"
			}

			if _, exists := seen[importPath]; !exists {
				seen[importPath] = struct{}{}
				imports = append(imports, importPath)
			}
		}
	}

	return imports
}

// extractBaseImport extracts the base file path from a partial import.
// For example: "people.managers" -> "people"
// But preserves paths: "path/people.managers" -> "path/people"
func extractBaseImport(importPath string) string {
	// Find the last path separator
	lastSlash := strings.LastIndex(importPath, "/")

	// Get the filename part (after the last slash, or the whole string)
	var prefix, filename string
	if lastSlash >= 0 {
		prefix = importPath[:lastSlash+1]
		filename = importPath[lastSlash+1:]
	} else {
		prefix = ""
		filename = importPath
	}

	// If filename contains a dot that's not at the start (not a relative path indicator)
	// and doesn't look like an extension, it's a partial import
	dotIdx := strings.Index(filename, ".")
	if dotIdx > 0 && !strings.HasSuffix(filename, ".d2") {
		// Check if this looks like a partial import (e.g., people.managers)
		// vs a relative path (e.g., ./styles)
		if !strings.HasPrefix(filename, ".") {
			filename = filename[:dotIdx]
		}
	}

	return prefix + filename
}

// resolveImportPaths resolves import paths relative to the source file's directory.
// sourcePath is the path to the D2 source file.
// imports is a list of import paths extracted from the source.
// Returns absolute paths to the imported files.
func resolveImportPaths(sourcePath string, imports []string) []string {
	sourceDir := filepath.Dir(sourcePath)
	resolved := make([]string, 0, len(imports))

	for _, imp := range imports {
		// Join with source directory and clean the path
		absPath := filepath.Join(sourceDir, imp)
		absPath = filepath.Clean(absPath)
		resolved = append(resolved, absPath)
	}

	return resolved
}

// ExtractAllImports recursively extracts all import file paths from a D2 file
// and all files it imports. This handles nested imports where an imported file
// may itself import other files.
// Returns a deduplicated list of absolute paths to all imported files.
// Handles circular imports by tracking visited files.
func ExtractAllImports(sourcePath string) []string {
	visited := make(map[string]struct{})
	var result []string

	// Start recursive extraction
	extractImportsRecursive(sourcePath, visited, &result)

	return result
}

// extractImportsRecursive is a helper that recursively extracts imports.
// visited tracks files we've already processed to handle circular imports.
// result accumulates all discovered import paths.
func extractImportsRecursive(filePath string, visited map[string]struct{}, result *[]string) {
	// Normalize the path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return
	}
	absPath = filepath.Clean(absPath)

	// Skip if already visited (handles circular imports)
	if _, ok := visited[absPath]; ok {
		return
	}
	visited[absPath] = struct{}{}

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		// File doesn't exist or can't be read - skip it
		return
	}

	// Extract direct imports from this file
	imports := extractImports(string(content))
	importPaths := resolveImportPaths(absPath, imports)

	// Process each import
	for _, importPath := range importPaths {
		// Check if file exists
		if _, err := os.Stat(importPath); err != nil {
			continue
		}

		// Add to result if not the source file and not already added
		if _, ok := visited[importPath]; !ok {
			*result = append(*result, importPath)
		}

		// Recursively process this import's imports
		extractImportsRecursive(importPath, visited, result)
	}
}
