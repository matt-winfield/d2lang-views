package compile

import (
	"slices"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 0 {
		t.Fatalf("expected 0 views for empty layers, got %d", len(views))
	}
}

func TestGetViewsNodes_NilMap(t *testing.T) {
	var d2graph *d2graph.Graph = &d2graph.Graph{}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("simple.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view in simple.d2, got %d", len(views))
	}

	name := views[0].Name
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
	d2graph, _, err := CompileD2("basic.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 3 {
		t.Fatalf("expected 3 views in basic.d2, got %d", len(views))
	}

	// Verify custom name is extracted
	var foundCustomName bool
	for _, view := range views {
		name := view.Root.Label.Value
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
	d2graph, _, err := CompileD2("no_views.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	reader := strings.NewReader(content)
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)
	if len(views) != 0 {
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
	reader := strings.NewReader(content)
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 0 {
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
	reader := strings.NewReader(content)
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 0 {
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
	reader := strings.NewReader(content)
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	viewName := views[0].Name
	if viewName != "myview" {
		t.Fatalf("expected view to be detected when # view comment exists with other comments, got '%s'", viewName)
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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	name := views[0].Root.Label.Value
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
	d2graph, _, err := CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	views := GetViewsNodes(d2graph)

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
			d2graph, _, err := CompileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			entities := ExtractRootObjectIds(d2graph)

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
