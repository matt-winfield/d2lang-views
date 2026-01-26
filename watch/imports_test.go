package watch

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "no imports",
			content:  `client: "Web Client"`,
			expected: []string{},
		},
		{
			name:     "single regular import",
			content:  `icons: @icons`,
			expected: []string{"icons.d2"},
		},
		{
			name:     "single spread import",
			content:  `...@common`,
			expected: []string{"common.d2"},
		},
		{
			name: "multiple imports",
			content: `icons: @icons
...@common
theme: @styles/dark`,
			expected: []string{"icons.d2", "common.d2", "styles/dark.d2"},
		},
		{
			name:     "import with relative path",
			content:  `styles: @./styles/main`,
			expected: []string{"./styles/main.d2"},
		},
		{
			name:     "import with parent directory",
			content:  `shared: @../shared/components`,
			expected: []string{"../shared/components.d2"},
		},
		{
			name:     "partial import",
			content:  `manager: @people.managers`,
			expected: []string{"people.d2"},
		},
		{
			name:     "quoted filename with dots",
			content:  `schema: @"schema-v0.1.2"`,
			expected: []string{"schema-v0.1.2.d2"},
		},
		{
			name: "import in nested structure",
			content: `system: {
    icons: @icons
    ...@styles
}`,
			expected: []string{"icons.d2", "styles.d2"},
		},
		{
			name:     "import with extension already present should still work",
			content:  `data: @data.d2`,
			expected: []string{"data.d2"},
		},
		{
			name: "ignore import-like text in strings",
			content: `label: "use @import syntax"
real: @actual`,
			expected: []string{"actual.d2"},
		},
		{
			name: "multiple imports same file",
			content: `a: @shared
b: @shared`,
			expected: []string{"shared.d2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractImports(tt.content)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("ExtractImports() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveImportPaths(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		imports    []string
		expected   []string
	}{
		{
			name:       "simple import in same directory",
			sourcePath: "/project/diagram.d2",
			imports:    []string{"icons.d2"},
			expected:   []string{"/project/icons.d2"},
		},
		{
			name:       "import with subdirectory",
			sourcePath: "/project/diagram.d2",
			imports:    []string{"styles/dark.d2"},
			expected:   []string{"/project/styles/dark.d2"},
		},
		{
			name:       "import with relative path",
			sourcePath: "/project/src/diagram.d2",
			imports:    []string{"./styles/main.d2"},
			expected:   []string{"/project/src/styles/main.d2"},
		},
		{
			name:       "import with parent directory",
			sourcePath: "/project/src/diagram.d2",
			imports:    []string{"../shared/components.d2"},
			expected:   []string{"/project/shared/components.d2"},
		},
		{
			name:       "multiple imports",
			sourcePath: "/project/diagram.d2",
			imports:    []string{"icons.d2", "styles.d2"},
			expected:   []string{"/project/icons.d2", "/project/styles.d2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveImportPaths(tt.sourcePath, tt.imports)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("ResolveImportPaths() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
