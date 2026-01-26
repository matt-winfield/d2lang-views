package watch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractAllImports(t *testing.T) {
	// Create a temp directory with nested imports
	tempDir := t.TempDir()

	// main.d2 imports icons.d2
	mainPath := filepath.Join(tempDir, "main.d2")
	mainContent := `client: "Web Client"
icons: @icons`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.d2: %v", err)
	}

	// icons.d2 imports shared/base.d2
	iconsPath := filepath.Join(tempDir, "icons.d2")
	iconsContent := `...@shared/base
icon1: "Icon 1"`
	if err := os.WriteFile(iconsPath, []byte(iconsContent), 0644); err != nil {
		t.Fatalf("failed to write icons.d2: %v", err)
	}

	// Create shared directory
	sharedDir := filepath.Join(tempDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared dir: %v", err)
	}

	// shared/base.d2 imports ../colors.d2
	basePath := filepath.Join(sharedDir, "base.d2")
	baseContent := `colors: @../colors
base_icon: "Base"`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write shared/base.d2: %v", err)
	}

	// colors.d2 has no imports
	colorsPath := filepath.Join(tempDir, "colors.d2")
	colorsContent := `primary: "#ff0000"`
	if err := os.WriteFile(colorsPath, []byte(colorsContent), 0644); err != nil {
		t.Fatalf("failed to write colors.d2: %v", err)
	}

	// Extract all imports recursively
	allImports := ExtractAllImports(mainPath)

	// Should find all 3 imported files
	expected := map[string]struct{}{
		iconsPath:  {},
		basePath:   {},
		colorsPath: {},
	}

	if len(allImports) != len(expected) {
		t.Errorf("expected %d imports, got %d: %v", len(expected), len(allImports), allImports)
	}

	for _, imp := range allImports {
		if _, ok := expected[imp]; !ok {
			t.Errorf("unexpected import: %s", imp)
		}
	}
}

func TestExtractAllImports_CircularImport(t *testing.T) {
	// Create a temp directory with circular imports
	tempDir := t.TempDir()

	// a.d2 imports b.d2
	aPath := filepath.Join(tempDir, "a.d2")
	aContent := `a: @b`
	if err := os.WriteFile(aPath, []byte(aContent), 0644); err != nil {
		t.Fatalf("failed to write a.d2: %v", err)
	}

	// b.d2 imports a.d2 (circular)
	bPath := filepath.Join(tempDir, "b.d2")
	bContent := `b: @a`
	if err := os.WriteFile(bPath, []byte(bContent), 0644); err != nil {
		t.Fatalf("failed to write b.d2: %v", err)
	}

	// Should handle circular imports without infinite loop
	allImports := ExtractAllImports(aPath)

	// Should find b.d2 (and not loop forever)
	if len(allImports) != 1 {
		t.Errorf("expected 1 import, got %d: %v", len(allImports), allImports)
	}

	if len(allImports) > 0 && allImports[0] != bPath {
		t.Errorf("expected %s, got %s", bPath, allImports[0])
	}
}

func TestExtractAllImports_MissingFile(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()

	// main.d2 imports a file that doesn't exist
	mainPath := filepath.Join(tempDir, "main.d2")
	mainContent := `icons: @missing`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.d2: %v", err)
	}

	// Should handle missing files gracefully
	allImports := ExtractAllImports(mainPath)

	// Should return empty (missing file can't be read)
	if len(allImports) != 0 {
		t.Errorf("expected 0 imports for missing file, got %d: %v", len(allImports), allImports)
	}
}
