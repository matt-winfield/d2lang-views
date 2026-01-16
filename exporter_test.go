package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"oss.terrastruct.com/d2/d2ast"
)

func TestReplaceViewLayers_SingleView(t *testing.T) {
	content := `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: { #view
        a
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: { #view
        a: "Entity A"

    }
}
`
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_MultipleViews(t *testing.T) {
	content := `a: "Entity A"
b: "Entity B"
c: "Entity C"

layers: {
    view1: { #view
        a
    }
    view2: { #view
        b
        c
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a: \"Entity A\"\nb: \"Entity B\"\nc: \"Entity C\"\n\nlayers: {\n    view1: { #view\n        a: \"Entity A\"\n\n    }\n    view2: { #view\n        b: \"Entity B\"\nc: \"Entity C\"\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_NoViews(t *testing.T) {
	content := `a: "Entity A"
b: "Entity B"

layers: {
    layer1: {
        a
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// No views means output should be identical to input
	if diff := cmp.Diff(content, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_EmptyContent(t *testing.T) {
	content := ``
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty result, got: %q", result)
	}
}

func TestReplaceViewLayers_ViewWithLabel(t *testing.T) {
	content := `a: "Entity A"
b: "Entity B"

layers: {
    view1: { #view
        a: "Custom Label"
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// Full statement "a: Custom Label" is replaced with new content
	expected := "a: \"Entity A\"\nb: \"Entity B\"\n\nlayers: {\n    view1: { #view\n        a: \"Custom Label\"\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_NestedEntities(t *testing.T) {
	content := `a: {
    b: "Nested B"
}
c: "Entity C"

layers: {
    view1: { #view
        a.b
        c
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a: {\n    b: \"Nested B\"\n}\nc: \"Entity C\"\n\nlayers: {\n    view1: { #view\n        a.b: \"Nested B\"\nc: \"Entity C\"\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_ViewWithCommentMarker(t *testing.T) {
	content := `x: "X"
y: "Y"

layers: {
    view1: {
        # view
        x
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "x: \"X\"\ny: \"Y\"\n\nlayers: {\n    view1: {\n        # view\n        x: \"X\"\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_PreservesOriginalSource(t *testing.T) {
	content := `# This is a comment
a: "Entity A"
b: "Entity B"

# Another comment
a -> b: connection

layers: {
    view1: { #view
        a
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "# This is a comment\na: \"Entity A\"\nb: \"Entity B\"\n\n# Another comment\na -> b: connection\n\nlayers: {\n    view1: { #view\n        a: \"Entity A\"\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_OnlyIncludesRootEntities(t *testing.T) {
	content := `a: "Entity A"

layers: {
    view1: { #view
        a
        missing
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// Only 'a' should be included since 'missing' is not in root
	expected := "a: \"Entity A\"\n\nlayers: {\n    view1: { #view\n        a: \"Entity A\"\n\n        missing\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_ViewWithCustomName(t *testing.T) {
	content := `server: "API Server"
database: "PostgreSQL"

layers: {
    backend: "Backend Architecture" { #view
        server
        database
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "server: \"API Server\"\ndatabase: \"PostgreSQL\"\n\nlayers: {\n    backend: \"Backend Architecture\" { #view\n        server: \"API Server\"\ndatabase: \"PostgreSQL\"\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_DeeplyNestedEntities(t *testing.T) {
	content := `a: {
    b: {
        c: {
            d: "Deep"
        }
    }
}

layers: {
    view1: { #view
        a.b.c.d
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a: {\n    b: {\n        c: {\n            d: \"Deep\"\n        }\n    }\n}\n\nlayers: {\n    view1: { #view\n        a.b.c.d: \"Deep\"\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_MultipleEntitiesInView(t *testing.T) {
	content := `client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

layers: {
    view1: { #view
        client
        server
        database
        cache
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// First entity is replaced with all content, other entities leave whitespace where removed
	expected := "client: \"Web Client\"\nserver: \"API Server\"\ndatabase: \"PostgreSQL\"\ncache: \"Redis\"\n\nlayers: {\n    view1: { #view\n        client: \"Web Client\"\nserver: \"API Server\"\ndatabase: \"PostgreSQL\"\ncache: \"Redis\"\n\n\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_WithEdgesInView(t *testing.T) {
	content := `client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

layers: {
    view1: { #view
        client
        server
		database -> cache
		database -> something-else
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// First entity is replaced with all content, other entities leave whitespace where removed
	expected := `client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

layers: {
    view1: { #view
        client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"




    }
}
`
	if result != expected {
		diff := cmp.Diff(expected, result)
		t.Errorf("unexpected result.\nDiff (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_EmptyView(t *testing.T) {
	content := `a: "Entity A"

layers: {
    view1: { #view
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	// Empty view has no ranges to replace, so content is unchanged
	expected := content
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_MixedViewsAndLayers(t *testing.T) {
	content := `a: "Entity A"
b: "Entity B"
c: "Entity C"

layers: {
    view1: { #view
        a
    }
    layer1: {
        b
    }
    view2: { #view
        c
    }
    layer2: {
        a
        b
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a: \"Entity A\"\nb: \"Entity B\"\nc: \"Entity C\"\n\nlayers: {\n    view1: { #view\n        a: \"Entity A\"\n\n    }\n    layer1: {\n        b\n    }\n    view2: { #view\n        c: \"Entity C\"\n\n    }\n    layer2: {\n        a\n        b\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_EntityWithoutLabel(t *testing.T) {
	content := `a
b: "Entity B"

layers: {
    view1: { #view
        a
        b
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a\nb: \"Entity B\"\n\nlayers: {\n    view1: { #view\n        a\nb: \"Entity B\"\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

func TestReplaceViewLayers_DuplicateIdsDifferentParents(t *testing.T) {
	content := `a: {
    x: "X in A"
}
b: {
    x: "X in B"
}

layers: {
    view1: { #view
        a.x
        b.x
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	rootObjectIds := extractRootObjectIds(graph)

	reader2 := strings.NewReader(content)
	result, err := replaceViewLayers(reader2, graph, rootObjectIds)
	if err != nil {
		t.Fatalf("replaceViewLayers failed: %v", err)
	}

	expected := "a: {\n    x: \"X in A\"\n}\nb: {\n    x: \"X in B\"\n}\n\nlayers: {\n    view1: { #view\n        a.x: \"X in A\"\nb.x: \"X in B\"\n\n\n    }\n}\n"
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected result (-expected +got):\n%s", diff)
	}
}

// Helper to create a d2ast.Range with just byte positions
func makeRange(startByte, endByte int) d2ast.Range {
	return d2ast.Range{
		Start: d2ast.Position{Byte: startByte},
		End:   d2ast.Position{Byte: endByte},
	}
}

// Helper to create a rangeOperation
func makeOp(startByte, endByte int, replacement string) rangeOperation {
	return rangeOperation{
		r:           makeRange(startByte, endByte),
		replacement: replacement,
	}
}

func TestApplyRangeOperations_EmptyOps(t *testing.T) {
	source := "hello world"
	result := applyRangeOperations(source, []rangeOperation{})
	if result != source {
		t.Errorf("expected %q, got %q", source, result)
	}
}

func TestApplyRangeOperations_SingleReplacement(t *testing.T) {
	source := "hello world"
	ops := []rangeOperation{makeOp(0, 5, "hi")} // replace "hello" with "hi"
	result := applyRangeOperations(source, ops)
	expected := "hi world"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_SingleRemoval(t *testing.T) {
	source := "hello world"
	ops := []rangeOperation{makeOp(5, 11, "")} // remove " world"
	result := applyRangeOperations(source, ops)
	expected := "hello"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_MultipleNonOverlapping(t *testing.T) {
	source := "abcdefghij"
	ops := []rangeOperation{
		makeOp(2, 4, "XX"), // replace "cd" with "XX"
		makeOp(6, 8, ""),   // remove "gh"
	}
	result := applyRangeOperations(source, ops)
	expected := "abXXefij"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_OverlappingWithReplacement(t *testing.T) {
	source := "abcdefghij"
	ops := []rangeOperation{
		makeOp(2, 5, "NEW"), // replace "cde" with "NEW"
		makeOp(4, 8, ""),    // remove "efgh" - overlaps, should merge
	}
	result := applyRangeOperations(source, ops)
	expected := "abNEWij" // merged range 2-8, replacement from first op
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_OverlappingReplacementSecond(t *testing.T) {
	source := "abcdefghij"
	ops := []rangeOperation{
		makeOp(2, 5, ""),       // remove "cde"
		makeOp(4, 8, "SECOND"), // replace "efgh" - overlaps, but first op has no replacement
	}
	result := applyRangeOperations(source, ops)
	expected := "abSECONDij" // merged range 2-8, replacement from second op
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_ReplaceWithLongerContent(t *testing.T) {
	source := "abc"
	ops := []rangeOperation{makeOp(1, 2, "LONGER")} // replace "b" with "LONGER"
	result := applyRangeOperations(source, ops)
	expected := "aLONGERc"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyRangeOperations_MultipleReplacementsInOrder(t *testing.T) {
	source := "one two three"
	ops := []rangeOperation{
		makeOp(0, 3, "1"),  // "one" -> "1"
		makeOp(4, 7, "2"),  // "two" -> "2"
		makeOp(8, 13, "3"), // "three" -> "3"
	}
	result := applyRangeOperations(source, ops)
	expected := "1 2 3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
