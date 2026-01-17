package d2view

import (
	"strings"
	"testing"

	"github.com/matt-winfield/d2lang-views/compile"
)

func TestProcessViews(t *testing.T) {
	tests := []struct {
		name                   string
		content                string
		expectedViewNames      []string
		expectedObjectsPerView [][]struct {
			id    string
			label string
		}
	}{
		{
			name: "Single view with single object",
			content: `
a -> b
layers: {
    view1: { #view
        a
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: ""},
				},
			},
		},
		{
			name: "Single view with base label",
			content: `
a: "Node A"
a -> b
layers: {
    view1: { #view
        a
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: "Node A"},
				},
			},
		},
		{
			name: "Single view with view label",
			content: `
a
a -> b
layers: {
    view1: { #view
        a: "View Node A"
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: "View Node A"},
				},
			},
		},
		{
			name: "Single view with both base and view label",
			content: `
a: "Base Node A"
a -> b
layers: {
    view1: { #view
        a: "View Node A"
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: "View Node A"},
				},
			},
		},
		{
			name: "View with extra objects",
			content: `
a -> b
layers: {
	view1: { #view
		a
		extra
		extra-with-label: "Extra Node"
	}
}`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: ""},
					{id: "extra", label: ""},
					{id: "extra-with-label", label: "Extra Node"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			graph, _, err := compile.CompileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			viewLayers := compile.GetViewsNodes(graph)
			views := ProcessViews(viewLayers, graph)

			if len(views) != len(tt.expectedViewNames) {
				t.Fatalf("expected %d views, got %d", len(tt.expectedViewNames), len(views))
			}

			for i, expectedViewName := range tt.expectedViewNames {
				if views[i].Name != expectedViewName {
					t.Fatalf("expected view name '%s', got '%s'", expectedViewName, views[i].Name)
				}

				expectedObjects := tt.expectedObjectsPerView[i]
				if len(views[i].Objects) != len(expectedObjects) {
					t.Fatalf("expected %d objects in view, got %d", len(expectedObjects), len(views[i].Objects))
				}

				for j, expectedObj := range expectedObjects {
					if views[i].Objects[j].ID != expectedObj.id {
						t.Fatalf("expected object ID '%s', got '%s'", expectedObj.id, views[i].Objects[j].ID)
					}
					if views[i].Objects[j].Label != expectedObj.label {
						t.Fatalf("expected object Label '%s', got '%s'", expectedObj.label, views[i].Objects[j].Label)
					}
				}
			}
		})
	}
}
