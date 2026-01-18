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
		{
			name: "Multiple view layers",
			content: `
				a -> b
				layers: {
				    view1: { #view
				        a
				    }
				    view2: { #view
				        b
				    }
				}
			`,
			expectedViewNames: []string{"view1", "view2"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
			}{
				{
					{id: "a", label: ""},
				},
				{
					{id: "b", label: ""},
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

				if len(views[i].Edges) != 0 {
					t.Fatalf("expected 0 edges in view, got %d", len(views[i].Edges))
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

func TestProcessViews_WithRelationships(t *testing.T) {
	tests := []struct {
		name                   string
		content                string
		expectedViewNames      []string
		expectedObjectsPerView [][]struct {
			id    string
			label string
		}
		expectedEdgesPerView [][]struct {
			src      string
			dst      string
			srcArrow bool
			dstArrow bool
		}
	}{
		{
			name: "2 nodes with relationship in the base layer",
			content: `
a -> b
c
layers: {
    view1: { #view
        a
		b
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
					{id: "b", label: ""},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "a", dst: "b", srcArrow: false, dstArrow: true},
				},
			},
		},
		{
			name: "2 nodes with relationship in the view layer",
			content: `
a
b
c
layers: {
    view1: { #view
        a -> b
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
					{id: "b", label: ""},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "a", dst: "b", srcArrow: false, dstArrow: true},
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

				expectedEdges := tt.expectedEdgesPerView[i]
				if len(views[i].Edges) != len(expectedEdges) {
					t.Fatalf("expected %d edges in view, got %d", len(expectedEdges), len(views[i].Edges))
				}

				for k, expectedEdge := range expectedEdges {
					edge := views[i].Edges[k]
					if compile.GetAbsoluteId(edge.Src) != expectedEdge.src {
						t.Fatalf("expected edge source '%s', got '%s'", expectedEdge.src, compile.GetAbsoluteId(edge.Src))
					}
					if compile.GetAbsoluteId(edge.Dst) != expectedEdge.dst {
						t.Fatalf("expected edge destination '%s', got '%s'", expectedEdge.dst, compile.GetAbsoluteId(edge.Dst))
					}
					if edge.SrcArrow != expectedEdge.srcArrow {
						t.Fatalf("expected edge source arrow '%v', got '%v'", expectedEdge.srcArrow, edge.SrcArrow)
					}
					if edge.DstArrow != expectedEdge.dstArrow {
						t.Fatalf("expected edge destination arrow '%v', got '%v'", expectedEdge.dstArrow, edge.DstArrow)
					}
				}
			}
		})
	}
}
