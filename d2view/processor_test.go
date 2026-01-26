package d2view

import (
	"reflect"
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
			ida   []string
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
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "View Node A", ida: []string{"a"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "View Node A", ida: []string{"a"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
					{id: "extra", label: "", ida: []string{"extra"}},
					{id: "extra-with-label", label: "Extra Node", ida: []string{"extra-with-label"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
				},
				{
					{id: "b", label: "", ida: []string{"b"}},
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
					if !reflect.DeepEqual(views[i].Objects[j].StringIDA(), expectedObj.ida) {
						t.Fatalf("expected object StringIDA() '%v', got '%v'", expectedObj.ida, views[i].Objects[j].StringIDA())
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
			ida   []string
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
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
					{id: "b", label: "", ida: []string{"b"}},
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
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
					{id: "b", label: "", ida: []string{"b"}},
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
			name: "nested nodes with relationships - implicit parents filtered",
			content: `
a: "Node A" {
    b: "Node B" {
		c: "Node C"
	}
}
a.b -> d

layers: {
    view1: { #view
        a.b.c
		d
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "c", label: "Node C", ida: []string{"c"}},
					{id: "d", label: "", ida: []string{"d"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "multiple relationship directions",
			content: `
a; b; c; d;
a -> b
b <-> c
c <- d
a -- d

layers: {
    view1: { #view
		a;b;c;d
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "", ida: []string{"a"}},
					{id: "b", label: "", ida: []string{"b"}},
					{id: "c", label: "", ida: []string{"c"}},
					{id: "d", label: "", ida: []string{"d"}},
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
					{src: "b", dst: "c", srcArrow: true, dstArrow: true},
					{src: "c", dst: "d", srcArrow: true, dstArrow: false},
					{src: "a", dst: "d", srcArrow: false, dstArrow: false},
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
					if !reflect.DeepEqual(views[i].Objects[j].StringIDA(), expectedObj.ida) {
						t.Fatalf("expected object StringIDA() '%v', got '%v'", expectedObj.ida, views[i].Objects[j].StringIDA())
					}
				}

				expectedEdges := tt.expectedEdgesPerView[i]
				if len(views[i].Edges) != len(expectedEdges) {
					t.Fatalf("expected %d edges in view, got %d", len(expectedEdges), len(views[i].Edges))
				}

				for k, expectedEdge := range expectedEdges {
					edge := views[i].Edges[k]
					if edge.Src != expectedEdge.src {
						t.Fatalf("expected edge source '%s', got '%s'", expectedEdge.src, edge.Src)
					}
					if edge.Dst != expectedEdge.dst {
						t.Fatalf("expected edge destination '%s', got '%s'", expectedEdge.dst, edge.Dst)
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

func TestProcessViews_ImplicitParentFiltering(t *testing.T) {
	tests := []struct {
		name                   string
		content                string
		expectedViewNames      []string
		expectedObjectsPerView [][]struct {
			id    string
			label string
			ida   []string
		}
		expectedEdgesPerView [][]struct {
			src      string
			dst      string
			srcArrow bool
			dstArrow bool
		}
	}{
		{
			name: "nested child without explicit parent - parent should be filtered",
			content: `
parent: "Parent Label" {
    child: "Child Label"
}
layers: {
    view1: { #view
        parent.child
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "child", label: "Child Label", ida: []string{"child"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "nested child with explicit parent - both should remain",
			content: `
parent: "Parent Label" {
    child: "Child Label"
}
layers: {
    view1: { #view
        parent
        parent.child
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "parent", label: "Parent Label", ida: []string{"parent"}},
					{id: "child", label: "Child Label", ida: []string{"parent", "child"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "deeply nested child without ancestors - only leaf should remain",
			content: `
a: "Node A" {
    b: "Node B" {
        c: "Node C"
    }
}
layers: {
    view1: { #view
        a.b.c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "c", label: "Node C", ida: []string{"c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "deeply nested with explicit root - root and leaf remain, intermediate filtered",
			content: `
a: "Node A" {
    b: "Node B" {
        c: "Node C"
    }
}
layers: {
    view1: { #view
        a
        a.b.c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
					{id: "c", label: "Node C", ida: []string{"a", "c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "all ancestors explicit - all remain",
			content: `
a: "Node A" {
    b: "Node B" {
        c: "Node C"
    }
}
layers: {
    view1: { #view
        a
        a.b
        a.b.c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
					{id: "b", label: "Node B", ida: []string{"a", "b"}},
					{id: "c", label: "Node C", ida: []string{"a", "b", "c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "edge with nested reference - edge filtered when source ancestor removed",
			content: `
a: "Node A" {
    b: "Node B"
}
a.b -> c
layers: {
    view1: { #view
        a.b
        c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "b", label: "Node B", ida: []string{"b"}},
					{id: "c", label: "", ida: []string{"c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "b", dst: "c", srcArrow: false, dstArrow: true},
				},
			},
		},
		{
			name: "multiple nested paths with shared implicit parent",
			content: `
parent: "Parent" {
    child1: "Child 1"
    child2: "Child 2"
}
layers: {
    view1: { #view
        parent.child1
        parent.child2
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "child1", label: "Child 1", ida: []string{"child1"}},
					{id: "child2", label: "Child 2", ida: []string{"child2"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "edge defined in view with implicit parent",
			content: `
a: "Node A" {
    b: "Node B"
}
c
layers: {
    view1: { #view
        a.b -> c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "b", label: "Node B", ida: []string{"b"}},
					{id: "c", label: "", ida: []string{"c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "b", dst: "c", srcArrow: false, dstArrow: true},
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
					t.Fatalf("expected %d objects in view, got %d.\nExpected: %+v\nGot objects with IDs: %v",
						len(expectedObjects), len(views[i].Objects), expectedObjects, getObjectIDs(views[i].Objects))
				}

				for j, expectedObj := range expectedObjects {
					if views[i].Objects[j].ID != expectedObj.id {
						t.Fatalf("expected object ID '%s', got '%s'", expectedObj.id, views[i].Objects[j].ID)
					}
					if views[i].Objects[j].Label != expectedObj.label {
						t.Fatalf("expected object Label '%s', got '%s'", expectedObj.label, views[i].Objects[j].Label)
					}
					if !reflect.DeepEqual(views[i].Objects[j].StringIDA(), expectedObj.ida) {
						t.Fatalf("expected object StringIDA() '%v', got '%v'", expectedObj.ida, views[i].Objects[j].StringIDA())
					}
				}

				expectedEdges := tt.expectedEdgesPerView[i]
				if len(views[i].Edges) != len(expectedEdges) {
					t.Fatalf("expected %d edges in view, got %d", len(expectedEdges), len(views[i].Edges))
				}

				for k, expectedEdge := range expectedEdges {
					edge := views[i].Edges[k]
					if edge.Src != expectedEdge.src {
						t.Fatalf("expected edge source '%s', got '%s'", expectedEdge.src, edge.Src)
					}
					if edge.Dst != expectedEdge.dst {
						t.Fatalf("expected edge destination '%s', got '%s'", expectedEdge.dst, edge.Dst)
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

func getObjectIDs(objects []*Object) []string {
	ids := make([]string, len(objects))
	for i, obj := range objects {
		ids[i] = obj.ID
	}
	return ids
}

func TestProcessViews_IncludeParentsComment(t *testing.T) {
	tests := []struct {
		name                   string
		content                string
		expectedViewNames      []string
		expectedObjectsPerView [][]struct {
			id    string
			label string
			ida   []string
		}
		expectedEdgesPerView [][]struct {
			src      string
			dst      string
			srcArrow bool
			dstArrow bool
		}
	}{
		{
			name: "nested child with include-parents comment - parent should be included",
			content: `
parent: "Parent Label" {
    child: "Child Label"
}
layers: {
    view1: { #view
        parent.child # include-parents
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "parent", label: "Parent Label", ida: []string{"parent"}},
					{id: "child", label: "Child Label", ida: []string{"parent", "child"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "deeply nested with include-parents - all ancestors included",
			content: `
a: "Node A" {
    b: "Node B" {
        c: "Node C"
    }
}
layers: {
    view1: { #view
        a.b.c # include-parents
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
					{id: "b", label: "Node B", ida: []string{"a", "b"}},
					{id: "c", label: "Node C", ida: []string{"a", "b", "c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "multiple references, only one with include-parents",
			content: `
parent1: "Parent 1" {
    child1: "Child 1"
}
parent2: "Parent 2" {
    child2: "Child 2"
}
layers: {
    view1: { #view
        parent1.child1 # include-parents
        parent2.child2
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "parent1", label: "Parent 1", ida: []string{"parent1"}},
					{id: "child1", label: "Child 1", ida: []string{"parent1", "child1"}},
					{id: "child2", label: "Child 2", ida: []string{"child2"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "include-parents with edge preservation",
			content: `
a: "Node A" {
    b: "Node B"
}
a.b -> c
layers: {
    view1: { #view
        a.b # include-parents
        c
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
					{id: "b", label: "Node B", ida: []string{"a", "b"}},
					{id: "c", label: "", ida: []string{"c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "a.b", dst: "c", srcArrow: false, dstArrow: true},
				},
			},
		},
		{
			name: "partial explicit with include-parents on leaf",
			content: `
a: "Node A" {
    b: "Node B" {
        c: "Node C"
    }
}
layers: {
    view1: { #view
        a
        a.b.c # include-parents
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "a", label: "Node A", ida: []string{"a"}},
					{id: "b", label: "Node B", ida: []string{"a", "b"}},
					{id: "c", label: "Node C", ida: []string{"a", "b", "c"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
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
					t.Fatalf("expected %d objects in view, got %d.\nExpected: %+v\nGot objects with IDs: %v",
						len(expectedObjects), len(views[i].Objects), expectedObjects, getObjectIDs(views[i].Objects))
				}

				for j, expectedObj := range expectedObjects {
					if views[i].Objects[j].ID != expectedObj.id {
						t.Fatalf("expected object ID '%s', got '%s'", expectedObj.id, views[i].Objects[j].ID)
					}
					if views[i].Objects[j].Label != expectedObj.label {
						t.Fatalf("expected object Label '%s', got '%s'", expectedObj.label, views[i].Objects[j].Label)
					}
					if !reflect.DeepEqual(views[i].Objects[j].StringIDA(), expectedObj.ida) {
						t.Fatalf("expected object StringIDA() '%v', got '%v'", expectedObj.ida, views[i].Objects[j].StringIDA())
					}
				}

				expectedEdges := tt.expectedEdgesPerView[i]
				if len(views[i].Edges) != len(expectedEdges) {
					t.Fatalf("expected %d edges in view, got %d", len(expectedEdges), len(views[i].Edges))
				}

				for k, expectedEdge := range expectedEdges {
					edge := views[i].Edges[k]
					if edge.Src != expectedEdge.src {
						t.Fatalf("expected edge source '%s', got '%s'", expectedEdge.src, edge.Src)
					}
					if edge.Dst != expectedEdge.dst {
						t.Fatalf("expected edge destination '%s', got '%s'", expectedEdge.dst, edge.Dst)
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

func TestProcessViews_CaseInsensitiveMatching(t *testing.T) {
	tests := []struct {
		name                   string
		content                string
		expectedViewNames      []string
		expectedObjectsPerView [][]struct {
			id    string
			label string
			ida   []string
		}
		expectedEdgesPerView [][]struct {
			src      string
			dst      string
			srcArrow bool
			dstArrow bool
		}
	}{
		{
			name: "lowercase reference to uppercase base object",
			content: `
Test: "Test Label"
layers: {
    view1: { #view
        test
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "test", label: "Test Label", ida: []string{"test"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "uppercase reference to lowercase base object",
			content: `
test: "Test Label"
layers: {
    view1: { #view
        TEST
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "TEST", label: "Test Label", ida: []string{"TEST"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "mixed case nested reference",
			content: `
Parent: "Parent Label" {
    Child: "Child Label"
}
layers: {
    view1: { #view
        parent.child
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "child", label: "Child Label", ida: []string{"child"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
			},
		},
		{
			name: "case insensitive edge matching from base layer",
			content: `
NodeA: "Node A"
NodeB: "Node B"
NodeA -> NodeB
layers: {
    view1: { #view
        nodea
        nodeb
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "nodea", label: "Node A", ida: []string{"nodea"}},
					{id: "nodeb", label: "Node B", ida: []string{"nodeb"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{
					{src: "nodea", dst: "nodeb", srcArrow: false, dstArrow: true},
				},
			},
		},
		{
			name: "deeply nested case insensitive with explicit parent",
			content: `
First: "First Label" {
    Second: "Second Label" {
        Third: "Third Label"
    }
}
layers: {
    view1: { #view
        first
        first.second.third
    }
}
`,
			expectedViewNames: []string{"view1"},
			expectedObjectsPerView: [][]struct {
				id    string
				label string
				ida   []string
			}{
				{
					{id: "first", label: "First Label", ida: []string{"first"}},
					{id: "third", label: "Third Label", ida: []string{"first", "third"}},
				},
			},
			expectedEdgesPerView: [][]struct {
				src      string
				dst      string
				srcArrow bool
				dstArrow bool
			}{
				{},
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
					t.Fatalf("expected %d objects in view, got %d.\nExpected: %+v\nGot objects with IDs: %v",
						len(expectedObjects), len(views[i].Objects), expectedObjects, getObjectIDs(views[i].Objects))
				}

				for j, expectedObj := range expectedObjects {
					if views[i].Objects[j].ID != expectedObj.id {
						t.Fatalf("expected object ID '%s', got '%s'", expectedObj.id, views[i].Objects[j].ID)
					}
					if views[i].Objects[j].Label != expectedObj.label {
						t.Fatalf("expected object Label '%s', got '%s'", expectedObj.label, views[i].Objects[j].Label)
					}
					if !reflect.DeepEqual(views[i].Objects[j].StringIDA(), expectedObj.ida) {
						t.Fatalf("expected object StringIDA() '%v', got '%v'", expectedObj.ida, views[i].Objects[j].StringIDA())
					}
				}

				expectedEdges := tt.expectedEdgesPerView[i]
				if len(views[i].Edges) != len(expectedEdges) {
					t.Fatalf("expected %d edges in view, got %d", len(expectedEdges), len(views[i].Edges))
				}

				for k, expectedEdge := range expectedEdges {
					edge := views[i].Edges[k]
					if edge.Src != expectedEdge.src {
						t.Fatalf("expected edge source '%s', got '%s'", expectedEdge.src, edge.Src)
					}
					if edge.Dst != expectedEdge.dst {
						t.Fatalf("expected edge destination '%s', got '%s'", expectedEdge.dst, edge.Dst)
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
