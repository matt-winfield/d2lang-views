package d2view

import (
	"strings"
	"testing"

	"github.com/matt-winfield/d2lang-views/compile"
)

func TestProcessViews_SingleView(t *testing.T) {
	content := `
a -> b
layers: {
    view1: { #view
        a
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compile.CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	viewLayers := compile.GetViewsNodes(graph)
	views := ProcessViews(viewLayers, graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	if views[0].Name != "view1" {
		t.Fatalf("expected view name 'view1', got '%s'", views[0].Name)
	}

	if len(views[0].Edges) != 0 {
		t.Fatalf("expected 0 edges in view, got %d", len(views[0].Edges))
	}

	if len(views[0].Objects) != 1 {
		t.Fatalf("expected 1 object in view, got %d", len(views[0].Objects))
	}

	if views[0].Objects[0].ID != "a" {
		t.Fatalf("expected object ID 'a', got '%s'", views[0].Objects[0].ID)
	}
}

func TestProcessViews_SingleViewWithBaseLabel(t *testing.T) {
	content := `
a: "Node A"
a -> b
layers: {
    view1: { #view
        a
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compile.CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	viewLayers := compile.GetViewsNodes(graph)
	views := ProcessViews(viewLayers, graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	if views[0].Name != "view1" {
		t.Fatalf("expected view name 'view1', got '%s'", views[0].Name)
	}

	if len(views[0].Edges) != 0 {
		t.Fatalf("expected 0 edges in view, got %d", len(views[0].Edges))
	}

	if len(views[0].Objects) != 1 {
		t.Fatalf("expected 1 object in view, got %d", len(views[0].Objects))
	}

	if views[0].Objects[0].ID != "a" {
		t.Fatalf("expected object ID 'a', got '%s'", views[0].Objects[0].ID)
	}

	if views[0].Objects[0].Label != "Node A" {
		t.Fatalf("expected object Label 'Node A', got '%s'", views[0].Objects[0].Label)
	}
}

func TestProcessViews_SingleViewWithViewLabel(t *testing.T) {
	content := `
a
a -> b
layers: {
    view1: { #view
        a: "View Node A"
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compile.CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	viewLayers := compile.GetViewsNodes(graph)
	views := ProcessViews(viewLayers, graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	if views[0].Name != "view1" {
		t.Fatalf("expected view name 'view1', got '%s'", views[0].Name)
	}

	if len(views[0].Edges) != 0 {
		t.Fatalf("expected 0 edges in view, got %d", len(views[0].Edges))
	}

	if len(views[0].Objects) != 1 {
		t.Fatalf("expected 1 object in view, got %d", len(views[0].Objects))
	}

	if views[0].Objects[0].ID != "a" {
		t.Fatalf("expected object ID 'a', got '%s'", views[0].Objects[0].ID)
	}

	if views[0].Objects[0].Label != "View Node A" {
		t.Fatalf("expected object Label 'View Node A', got '%s'", views[0].Objects[0].Label)
	}
}

func TestProcessViews_SingleViewWithBaseAndViewLabel(t *testing.T) {
	content := `
a: "Base Node A"
a -> b
layers: {
    view1: { #view
        a: "View Node A"
    }
}
`
	reader := strings.NewReader(content)
	graph, _, err := compile.CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	viewLayers := compile.GetViewsNodes(graph)
	views := ProcessViews(viewLayers, graph)

	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}

	if views[0].Name != "view1" {
		t.Fatalf("expected view name 'view1', got '%s'", views[0].Name)
	}

	if len(views[0].Edges) != 0 {
		t.Fatalf("expected 0 edges in view, got %d", len(views[0].Edges))
	}

	if len(views[0].Objects) != 1 {
		t.Fatalf("expected 1 object in view, got %d", len(views[0].Objects))
	}

	if views[0].Objects[0].ID != "a" {
		t.Fatalf("expected object ID 'a', got '%s'", views[0].Objects[0].ID)
	}

	if views[0].Objects[0].Label != "View Node A" {
		t.Fatalf("expected object Label 'View Node A', got '%s'", views[0].Objects[0].Label)
	}
}
