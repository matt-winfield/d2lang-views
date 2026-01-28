package compile

import (
	"os"
	"path/filepath"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
)

// compileTestFile is a helper that opens, compiles, and properly closes a D2 file for testing.
func compileTestFile(t *testing.T, path string) *d2graph.Graph {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("failed to close file: %v", err)
		}
	})

	graph, _, err := CompileD2(path, file)
	if err != nil {
		t.Fatalf("failed to compile D2: %v", err)
	}
	return graph
}

func TestGetIncludeParentsReferences_ImportedFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file that has a view importing an external file
	mainContent := `parent: {
	child: "Child"
}

layers: {
	myview: { #view
		...@view_content
	}
}
`

	// Create imported file containing include-parents comment
	viewContentFile := `parent.child #include-parents
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child to be detected as an include-parents reference
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child"]; !ok {
		t.Errorf("expected 'parent.child' to be in include-parents references, got: %v", refs)
	}
}

func TestGetOverrideEdges_ImportedFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file that has a view importing an external file
	mainContent := `a: "A"
b: "B"
a -> b: "original"

layers: {
	myview: { #view
		...@view_content
	}
}
`

	// Create imported file containing override comment
	viewContentFile := `a
b
a -> b: "new label" #override
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get override edges for the view
	overrides := GetOverrideEdges(graph, "myview", nil)

	// We expect a -> b with "new label" to be detected as an override
	if len(overrides) != 1 {
		t.Fatalf("expected 1 override edge, got %d: %v", len(overrides), overrides)
	}

	// Check that the override has the expected label
	found := false
	for key := range overrides {
		if key.Label == "new label" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected override edge with label 'new label', got: %v", overrides)
	}
}

func TestGetIncludeParentsReferences_NestedImports(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file that has a view importing an external file
	mainContent := `parent: {
	child: {
		grandchild: "Grandchild"
	}
}

layers: {
	myview: { #view
		...@view_content
	}
}
`

	// Create first imported file that imports another file
	viewContentFile := `...@nested_content
parent.child
`

	// Create nested imported file containing include-parents comment
	nestedContentFile := `parent.child.grandchild #include-parents
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")
	nestedContentPath := filepath.Join(tempDir, "nested_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}
	if err := os.WriteFile(nestedContentPath, []byte(nestedContentFile), 0644); err != nil {
		t.Fatalf("failed to write nested content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child.grandchild to be detected from the nested import
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child.grandchild"]; !ok {
		t.Errorf("expected 'parent.child.grandchild' to be in include-parents references, got: %v", refs)
	}
}

func TestGetOverrideEdges_NestedImports(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file that has a view importing an external file
	mainContent := `a: "A"
b: "B"
c: "C"
a -> b: "original ab"
b -> c: "original bc"

layers: {
	myview: { #view
		...@view_content
	}
}
`

	// Create first imported file that imports another file
	viewContentFile := `...@nested_content
a
b
c
a -> b: "override ab" #override
`

	// Create nested imported file containing override comment
	nestedContentFile := `b -> c: "override bc" #override
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")
	nestedContentPath := filepath.Join(tempDir, "nested_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}
	if err := os.WriteFile(nestedContentPath, []byte(nestedContentFile), 0644); err != nil {
		t.Fatalf("failed to write nested content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get override edges for the view
	overrides := GetOverrideEdges(graph, "myview", nil)

	// We expect both overrides: a -> b from view_content, b -> c from nested_content
	if len(overrides) != 2 {
		t.Fatalf("expected 2 override edges, got %d: %v", len(overrides), overrides)
	}

	// Check both overrides are present
	foundAB := false
	foundBC := false
	for key := range overrides {
		if key.Label == "override ab" {
			foundAB = true
		}
		if key.Label == "override bc" {
			foundBC = true
		}
	}
	if !foundAB {
		t.Errorf("expected override edge with label 'override ab', got: %v", overrides)
	}
	if !foundBC {
		t.Errorf("expected override edge with label 'override bc', got: %v", overrides)
	}
}

func TestGetIncludeParentsReferences_MixedDirectAndImported(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file with a view that has both direct comments and imported comments
	mainContent := `parent: {
	child1: "Child1"
	child2: "Child2"
}

layers: {
	myview: { #view
		parent.child1 #include-parents
		...@view_content
	}
}
`

	// Create imported file containing another include-parents comment
	viewContentFile := `parent.child2 #include-parents
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect both references: child1 from direct, child2 from import
	if len(refs) != 2 {
		t.Fatalf("expected 2 include-parents references, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child1"]; !ok {
		t.Errorf("expected 'parent.child1' (direct) to be in include-parents references, got: %v", refs)
	}
	if _, ok := refs["parent.child2"]; !ok {
		t.Errorf("expected 'parent.child2' (imported) to be in include-parents references, got: %v", refs)
	}
}

func TestGetOverrideEdges_MixedDirectAndImported(t *testing.T) {
	tempDir := t.TempDir()

	// Create main file with a view that has both direct overrides and imported overrides
	mainContent := `a: "A"
b: "B"
c: "C"
a -> b: "original ab"
b -> c: "original bc"

layers: {
	myview: { #view
		a
		b
		c
		a -> b: "direct override" #override
		...@view_content
	}
}
`

	// Create imported file containing another override comment
	viewContentFile := `b -> c: "imported override" #override
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get override edges for the view
	overrides := GetOverrideEdges(graph, "myview", nil)

	// We expect both overrides: a -> b direct, b -> c imported
	if len(overrides) != 2 {
		t.Fatalf("expected 2 override edges, got %d: %v", len(overrides), overrides)
	}

	// Check both overrides are present
	foundDirect := false
	foundImported := false
	for key := range overrides {
		if key.Label == "direct override" {
			foundDirect = true
		}
		if key.Label == "imported override" {
			foundImported = true
		}
	}
	if !foundDirect {
		t.Errorf("expected override edge with label 'direct override', got: %v", overrides)
	}
	if !foundImported {
		t.Errorf("expected override edge with label 'imported override', got: %v", overrides)
	}
}

func TestGetIncludeParentsReferences_SubdirectoryImport(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory
	subdir := filepath.Join(tempDir, "views")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create main file that imports from a subdirectory
	mainContent := `parent: {
	child: "Child"
}

layers: {
	myview: { #view
		...@views/view_content
	}
}
`

	// Create imported file in subdirectory containing include-parents comment
	viewContentFile := `parent.child #include-parents
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(subdir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child to be detected
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child"]; !ok {
		t.Errorf("expected 'parent.child' to be in include-parents references, got: %v", refs)
	}
}

func TestGetOverrideEdges_SubdirectoryImport(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory
	subdir := filepath.Join(tempDir, "views")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create main file that imports from a subdirectory
	mainContent := `a: "A"
b: "B"
a -> b: "original"

layers: {
	myview: { #view
		...@views/view_content
	}
}
`

	// Create imported file in subdirectory containing override comment
	viewContentFile := `a
b
a -> b: "subdirectory override" #override
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(subdir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get override edges for the view
	overrides := GetOverrideEdges(graph, "myview", nil)

	// We expect the override to be detected
	if len(overrides) != 1 {
		t.Fatalf("expected 1 override edge, got %d: %v", len(overrides), overrides)
	}

	found := false
	for key := range overrides {
		if key.Label == "subdirectory override" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected override edge with label 'subdirectory override', got: %v", overrides)
	}
}

func TestGetIncludeParentsReferences_ParentDirectoryImport(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory for the main file
	subdir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create main file in subdirectory that imports from parent directory
	mainContent := `parent: {
	child: "Child"
}

layers: {
	myview: { #view
		...@../shared/view_content
	}
}
`

	// Create shared directory at same level as src
	sharedDir := filepath.Join(tempDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared directory: %v", err)
	}

	// Create imported file in shared directory containing include-parents comment
	viewContentFile := `parent.child #include-parents
`

	mainPath := filepath.Join(subdir, "main.d2")
	viewContentPath := filepath.Join(sharedDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child to be detected
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child"]; !ok {
		t.Errorf("expected 'parent.child' to be in include-parents references, got: %v", refs)
	}
}

func TestGetOverrideEdges_ParentDirectoryImport(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory for the main file
	subdir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create main file in subdirectory that imports from parent directory
	mainContent := `a: "A"
b: "B"
a -> b: "original"

layers: {
	myview: { #view
		...@../shared/view_content
	}
}
`

	// Create shared directory at same level as src
	sharedDir := filepath.Join(tempDir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared directory: %v", err)
	}

	// Create imported file in shared directory containing override comment
	viewContentFile := `a
b
a -> b: "parent directory override" #override
`

	mainPath := filepath.Join(subdir, "main.d2")
	viewContentPath := filepath.Join(sharedDir, "view_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get override edges for the view
	overrides := GetOverrideEdges(graph, "myview", nil)

	// We expect the override to be detected
	if len(overrides) != 1 {
		t.Fatalf("expected 1 override edge, got %d: %v", len(overrides), overrides)
	}

	found := false
	for key := range overrides {
		if key.Label == "parent directory override" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected override edge with label 'parent directory override', got: %v", overrides)
	}
}

func TestGetIncludeParentsReferences_NestedSubdirectoryImports(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested subdirectories: views/components/
	componentsDir := filepath.Join(tempDir, "views", "components")
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		t.Fatalf("failed to create nested subdirectories: %v", err)
	}

	// Create main file that imports from a subdirectory
	mainContent := `parent: {
	child: {
		grandchild: "Grandchild"
	}
}

layers: {
	myview: { #view
		...@views/view_content
	}
}
`

	// Create first imported file in views/ that imports from components/
	viewContentFile := `...@components/nested_content
parent.child
`

	// Create nested imported file in views/components/ containing include-parents comment
	nestedContentFile := `parent.child.grandchild #include-parents
`

	mainPath := filepath.Join(tempDir, "main.d2")
	viewContentPath := filepath.Join(tempDir, "views", "view_content.d2")
	nestedContentPath := filepath.Join(componentsDir, "nested_content.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}
	if err := os.WriteFile(nestedContentPath, []byte(nestedContentFile), 0644); err != nil {
		t.Fatalf("failed to write nested content file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child.grandchild to be detected from the nested import
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child.grandchild"]; !ok {
		t.Errorf("expected 'parent.child.grandchild' to be in include-parents references, got: %v", refs)
	}
}

func TestGetIncludeParentsReferences_RelativeImportFromSubdirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create directories: src/ and shared/
	srcDir := filepath.Join(tempDir, "src")
	sharedDir := filepath.Join(tempDir, "shared")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src directory: %v", err)
	}
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatalf("failed to create shared directory: %v", err)
	}

	// Create main file in src/ that imports from src/views/
	mainContent := `parent: {
	child: "Child"
}

layers: {
	myview: { #view
		...@views/view_content
	}
}
`

	// Create views subdirectory under src
	viewsDir := filepath.Join(srcDir, "views")
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	// Create view_content.d2 that imports from ../../shared/ (going up two levels)
	viewContentFile := `...@../../shared/common
`

	// Create common.d2 in shared/ containing include-parents comment
	commonFile := `parent.child #include-parents
`

	mainPath := filepath.Join(srcDir, "main.d2")
	viewContentPath := filepath.Join(viewsDir, "view_content.d2")
	commonPath := filepath.Join(sharedDir, "common.d2")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main file: %v", err)
	}
	if err := os.WriteFile(viewContentPath, []byte(viewContentFile), 0644); err != nil {
		t.Fatalf("failed to write view content file: %v", err)
	}
	if err := os.WriteFile(commonPath, []byte(commonFile), 0644); err != nil {
		t.Fatalf("failed to write common file: %v", err)
	}

	// Compile from the main file
	graph := compileTestFile(t, mainPath)

	// Get include-parents references for the view
	refs := GetIncludeParentsReferences(graph, "myview", nil)

	// We expect parent.child to be detected from the deeply nested relative import
	if len(refs) != 1 {
		t.Fatalf("expected 1 include-parents reference, got %d: %v", len(refs), refs)
	}

	if _, ok := refs["parent.child"]; !ok {
		t.Errorf("expected 'parent.child' to be in include-parents references, got: %v", refs)
	}
}
