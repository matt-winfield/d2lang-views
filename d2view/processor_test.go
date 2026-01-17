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
	d2graph, _, err := compile.CompileD2("test.d2", reader)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	viewLayers := compile.GetViewsNodes(d2graph)
	views := ProcessViews(viewLayers)

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

	if views[0].Objects[0].Name != "a" {
		t.Fatalf("expected object ID 'a', got '%s'", views[0].Objects[0].Name)
	}
}
