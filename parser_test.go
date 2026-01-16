package main

import (
	"slices"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2graph"
)

// ============================================================================
// getLayersNode tests
// ============================================================================

func TestGetLayersNode_Found(t *testing.T) {
	content := `
a -> b
layers: {
    view1: {}
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result := getLayersNode(d2graph)

	if result == nil {
		t.Fatal("expected to find layers node")
	}
	if result.MapKey == nil || result.MapKey.Key == nil {
		t.Fatal("expected layers node to have a key")
	}
	keyStr := result.MapKey.Key.StringIDA()[0]
	if keyStr != "layers" {
		t.Fatalf("expected key 'layers', got '%s'", keyStr)
	}
}

func TestGetLayersNode_NotFound(t *testing.T) {
	content := `
a -> b
c: "Entity C"
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result := getLayersNode(d2graph)

	if result != nil {
		t.Fatal("expected nil when no layers node exists")
	}
}

func TestGetLayersNode_EmptyGraph(t *testing.T) {
	content := ``
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result := getLayersNode(d2graph)

	if result != nil {
		t.Fatal("expected nil for empty graph")
	}
}

func TestGetLayersNode_LayersNotFirst(t *testing.T) {
	content := `
a: "First"
b: "Second"
c -> d
layers: {
    view1: {}
}
e: "Last"
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	result := getLayersNode(d2graph)

	if result == nil {
		t.Fatal("expected to find layers node even when not first")
	}
}

func parseAndGetLayerNode(t *testing.T, content string, layerName string) d2ast.MapNodeBox {
	t.Helper()
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	layersNode := getLayersNode(d2graph)
	if layersNode == nil {
		t.Fatal("expected layers node")
	}

	for _, node := range layersNode.MapKey.Value.Map.Nodes {
		if node.MapKey != nil && node.MapKey.Key != nil {
			if node.MapKey.Key.StringIDA()[0] == layerName {
				return node
			}
		}
	}
	t.Fatalf("layer '%s' not found", layerName)
	return d2ast.MapNodeBox{}
}

func TestIsViewNode_WithViewComment(t *testing.T) {
	content := `
layers: {
    myview: { #view
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if !isViewNode(node) {
		t.Fatal("expected node with #view comment to be a view")
	}
}

func TestIsViewNode_WithStandaloneViewComment(t *testing.T) {
	content := `
layers: {
    myview: {
        # view
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if !isViewNode(node) {
		t.Fatal("expected node with standalone # view comment to be a view")
	}
}

func TestIsViewNode_WithViewCommentWhitespace(t *testing.T) {
	content := `
layers: {
    myview: {
        #   view
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if !isViewNode(node) {
		t.Fatal("expected node with whitespace in #view comment to be a view")
	}
}

func TestIsViewNode_NotAView(t *testing.T) {
	content := `
layers: {
    notaview: {
        a
        b
    }
}
`
	node := parseAndGetLayerNode(t, content, "notaview")

	if isViewNode(node) {
		t.Fatal("expected node without #view comment to not be a view")
	}
}

func TestIsViewNode_DifferentComment(t *testing.T) {
	content := `
layers: {
    someLayer: {
        # this is a regular comment
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "someLayer")

	if isViewNode(node) {
		t.Fatal("expected node with different comment to not be a view")
	}
}

func TestIsViewNode_EmptyLayer(t *testing.T) {
	content := `
layers: {
    emptyLayer: {}
}
`
	node := parseAndGetLayerNode(t, content, "emptyLayer")

	if isViewNode(node) {
		t.Fatal("expected empty layer to not be a view")
	}
}

func TestIsViewNode_NilMapKey(t *testing.T) {
	node := d2ast.MapNodeBox{MapKey: nil}
	if isViewNode(node) {
		t.Fatal("expected nil MapKey to not be a view")
	}
}

func TestGetNodeDisplayName_CustomName(t *testing.T) {
	content := `
layers: {
    myview: "Custom Display Name" { #view
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	name := getNodeDisplayName(node)

	if name != "Custom Display Name" {
		t.Fatalf("expected 'Custom Display Name', got '%s'", name)
	}
}

func TestGetNodeDisplayName_FallbackToKey(t *testing.T) {
	content := `
layers: {
    myview: { #view
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	name := getNodeDisplayName(node)

	if name != "myview" {
		t.Fatalf("expected 'myview', got '%s'", name)
	}
}

func TestGetNodeDisplayName_NilMapKey(t *testing.T) {
	node := d2ast.MapNodeBox{MapKey: nil}

	name := getNodeDisplayName(node)

	if name != "" {
		t.Fatalf("expected empty string for nil MapKey, got '%s'", name)
	}
}

func TestGetViewsNodes_SingleView(t *testing.T) {
	content := `
a -> b
layers: {
    view1: { #view
        a
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
}

func TestGetViewsNodes_MultipleViews(t *testing.T) {
	content := `
a -> b
layers: {
    view1: { #view
        a
    }
    view2: { #view
        b
    }
    notview: {
        a
    }
    view3: {
        # view
        a
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 3 {
		t.Fatalf("expected 3 views, got %d", len(views))
	}
}

func TestGetViewsNodes_NoViews(t *testing.T) {
	content := `
a -> b
layers: {
    layer1: {
        a
    }
    layer2: {
        b
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views, got %d", len(views))
	}
}

func TestGetViewsNodes_NoLayers(t *testing.T) {
	content := `
a -> b
c: "Entity C"
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views when no layers exist, got %d", len(views))
	}
}

func TestGetViewsNodes_EmptyLayers(t *testing.T) {
	content := `
a -> b
layers: {}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views for empty layers, got %d", len(views))
	}
}

func TestGetViewsNodes_NilMap(t *testing.T) {
	var d2graph *d2graph.Graph = &d2graph.Graph{}

	views := getViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views for nil map, got %d", len(views))
	}
}

func TestIntegration_SimpleFile(t *testing.T) {
	content := `
# Simple example with basic entities and one view
client: "Web Client"
server: "API Server"
database: "PostgreSQL"

client -> server: HTTP requests
server -> database: SQL queries

layers: {
    backend: { #view
        server
        database
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("simple.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view in simple.d2, got %d", len(views))
	}

	name := getNodeDisplayName(views[0])
	if name != "backend" {
		t.Fatalf("expected view name 'backend', got '%s'", name)
	}
}

func TestIntegration_BasicFile(t *testing.T) {
	content := `
first -> second

layers: {
    custom: "Custom Name" { #view
        first
        second
    }

    view2 {
        # view
        first
        second
    }

    not_a_view {
        first
        second
    }

    default: { #view
        first -> SomethingElse
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("basic.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 3 {
		t.Fatalf("expected 3 views in basic.d2, got %d", len(views))
	}

	// Verify custom name is extracted
	var foundCustomName bool
	for _, view := range views {
		name := getNodeDisplayName(view)
		if name == "Custom Name" {
			foundCustomName = true
		}
	}
	if !foundCustomName {
		t.Fatal("expected to find view with custom name 'Custom Name'")
	}
}

func TestIntegration_NoViewsFile(t *testing.T) {
	content := `
# Example with no view layers - edge case for testing
frontend: "React App"
backend: "Node.js API"
db: "MongoDB"

frontend -> backend: REST API
backend -> db: Mongoose

# This file has layers but none are marked as views
layers: {
    production: {
        frontend.style.fill: "#90EE90"
        backend.style.fill: "#90EE90"
    }
    staging: {
        frontend.style.fill: "#FFD700"
        backend.style.fill: "#FFD700"
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("no_views.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views in no_views.d2, got %d", len(views))
	}
}

func TestEdgeCase_ViewCommentCaseSensitive(t *testing.T) {
	// "VIEW" uppercase should not match
	content := `
layers: {
    myview: {
        # VIEW
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if isViewNode(node) {
		t.Fatal("expected uppercase VIEW to not match (case sensitive)")
	}
}

func TestEdgeCase_ViewCommentPartialMatch(t *testing.T) {
	// "viewer" should not match "view"
	content := `
layers: {
    myview: {
        # viewer
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if isViewNode(node) {
		t.Fatal("expected 'viewer' to not match 'view'")
	}
}

func TestEdgeCase_ViewCommentWithPrefix(t *testing.T) {
	// "myview" should not match "view"
	content := `
layers: {
    myview: {
        # myview
        a
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if isViewNode(node) {
		t.Fatal("expected 'myview' comment to not match 'view'")
	}
}

func TestEdgeCase_MultipleComments(t *testing.T) {
	// D2 parser may merge consecutive comments or handle them differently.
	// Test that a view comment with content after it is detected.
	content := `
layers: {
    myview: {
        # view
        a
        # another comment after content
    }
}
`
	node := parseAndGetLayerNode(t, content, "myview")

	if !isViewNode(node) {
		t.Fatal("expected view to be detected when # view comment exists with other comments")
	}
}

func TestEdgeCase_SpecialCharactersInEntityNames(t *testing.T) {
	content := `
"entity-with-dashes": "Display Name"
entity_with_underscores
entity.with.dots

layers: {
    myview: { #view
        "entity-with-dashes"
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
}

func TestEdgeCase_UnicodeInNames(t *testing.T) {
	content := `
日本語: "Japanese"
emoji: "🚀 Rocket"

layers: {
    unicode_view: "日本語ビュー" { #view
        日本語
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	name := getNodeDisplayName(views[0])
	if !strings.Contains(name, "日本語") {
		t.Fatalf("expected unicode display name, got '%s'", name)
	}
}

func TestEdgeCase_DeeplyNestedContent(t *testing.T) {
	content := `
a: {
    b: {
        c: {
            d: "Deep"
        }
    }
}

layers: {
    deep_view: { #view
        a.b.c.d
    }
}
`
	reader := strings.NewReader(content)
	d2graph, _, err := compileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := getViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
}

func TestExtractBaseLayerEntities(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedIDs []string
	}{
		{
			name: "Simple entities",
			content: `
a: "Entity A"
b: "Entity B"
a: "Entity A"
b: "Entity B"
c: "Entity C"

layers: {
	layer1: {
		a
	}
}

scenarios: {
    test: {
		b
	}
}
	
steps: {
    step1: {
		c
	}
}`,
			expectedIDs: []string{"a", "b", "c"},
		},
		{
			name: "Inline nested entities",
			content: `
a.b
a.b.c
d

layers: {
	layer1: {
		a
	}
}

scenarios: {
    test: {
		b
	}
}
	
steps: {
    step1: {
		c
	}
}`,
			expectedIDs: []string{"a", "a.b", "a.b.c", "d"},
		},
		{
			name: "Nested entities",
			content: `
a {
    b: "Entity B" {
		c
	}
}
d

layers: {
	layer1: {
		a
	}
}

scenarios: {
    test: {
		b
	}
}
	
steps: {
    step1: {
		c
	}
}`,
			expectedIDs: []string{"a", "a.b", "a.b.c", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			d2graph, _, err := compileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			entities := extractRootObjects(d2graph)

			for _, expected := range tt.expectedIDs {
				if !slices.Contains(entities, expected) {
					t.Fatalf("expected %s to be in %s", expected, entities)
				}
			}

			if len(entities) != len(tt.expectedIDs) {
				t.Fatalf("expected %d entities, got %d", len(tt.expectedIDs), len(entities))
			}
		})
	}
}
