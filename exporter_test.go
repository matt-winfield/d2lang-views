package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matt-winfield/d2lang-views/compile"
	"oss.terrastruct.com/d2/d2ast"
)

func TestReplaceViewLayers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "SingleView",
			content: `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: { #view
        a
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: {
        a: "Entity A"
    }
}
`,
		},
		{
			name: "MultipleViews",
			content: `a: "Entity A"
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
`,
			expected: `a: "Entity A"
b: "Entity B"
c: "Entity C"

layers: {
    view1: {
        a: "Entity A"
    }
    view2: {
        b: "Entity B"
        c: "Entity C"
    }
}
`,
		},
		{
			name: "NoViews",
			content: `a: "Entity A"
b: "Entity B"

layers: {
    layer1: {
        a
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

layers: {
    layer1: {
        a
    }
}
`,
		},
		{
			name:     "EmptyContent",
			content:  ``,
			expected: ``,
		},
		{
			name: "ViewWithLabel",
			content: `a: "Entity A"
b: "Entity B"

layers: {
    view1: { #view
        a: "Custom Label"
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

layers: {
    view1: {
        a: "Custom Label"
    }
}
`,
		},
		{
			name: "NestedEntities",
			content: `a: {
    b: "Nested B"
}
c: "Entity C"

layers: {
    view1: { #view
        a.b
        c
    }
}
`,
			expected: `a: {
    b: "Nested B"
}
c: "Entity C"

layers: {
    view1: {
        b: "Nested B"
        c: "Entity C"
    }
}
`,
		},
		{
			name: "ViewWithCommentMarker",
			content: `x: "X"
y: "Y"

layers: {
    view1: {
        # view
        x
    }
}
`,
			expected: `x: "X"
y: "Y"

layers: {
    view1: {
        x: "X"
    }
}
`,
		},
		{
			name: "PreservesOriginalSource",
			content: `# This is a comment
a: "Entity A"
b: "Entity B"

# Another comment
a -> b: connection

layers: {
    view1: { #view
        a
    }
}
`,
			expected: `# This is a comment
a: "Entity A"
b: "Entity B"

# Another comment
a -> b: connection

layers: {
    view1: {
        a: "Entity A"
    }
}
`,
		},
		{
			name: "OnlyIncludesRootEntities",
			content: `a: "Entity A"

layers: {
    view1: { #view
        a
        missing
    }
}
`,
			expected: `a: "Entity A"

layers: {
    view1: {
        a: "Entity A"
        missing
    }
}
`,
		},
		{
			name: "ViewWithCustomName",
			content: `server: "API Server"
database: "PostgreSQL"

layers: {
    backend: "Backend Architecture" { #view
        server
        database
    }
}
`,
			expected: `server: "API Server"
database: "PostgreSQL"

layers: {
    backend: "Backend Architecture" {
        server: "API Server"
        database: "PostgreSQL"
    }
}
`,
		},
		{
			name: "DeeplyNestedEntities",
			content: `a: {
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
`,
			expected: `a: {
    b: {
        c: {
            d: "Deep"
        }
    }
}

layers: {
    view1: {
        d: "Deep"
    }
}
`,
		},
		{
			name: "MultipleEntitiesInView",
			content: `client: "Web Client"
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
`,
			expected: `client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

layers: {
    view1: {
        client: "Web Client"
        server: "API Server"
        database: "PostgreSQL"
        cache: "Redis"
    }
}
`,
		},
		{
			name: "WithEdgesInView",
			content: `client: "Web Client"
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
`,
			expected: `client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

layers: {
    view1: {
        client: "Web Client"
        server: "API Server"
        database: "PostgreSQL"
        cache: "Redis"
        something-else
        database -> cache
        database -> something-else
    }
}
`,
		},
		{
			name: "MultipleViewsAndEdgesAndNonViewLayers",
			content: `first: "First Thing"
second: "Second Thing"
first -> second
third

layers: {
    custom: "Custom Name" { #view
        first
        second
    }

    view2 {
        # view
        first
        second.something
    }

    not_a_view {
        first
        second
    }

    default: { #view
        first -> SomethingElse
    }
}`,
			expected: `first: "First Thing"
second: "Second Thing"
first -> second
third

layers: {
    custom: "Custom Name" {
        first: "First Thing"
        second: "Second Thing"
        first -> second
    }

    view2: {
        first: "First Thing"
        something
    }

    not_a_view {
        first
        second
    }

    default: {
        first: "First Thing"
        SomethingElse
        first -> SomethingElse
    }
}`,
		},
		{
			name: "EmptyView",
			content: `a: "Entity A"

layers: {
    view1: {
    }
}
`,
			expected: `a: "Entity A"

layers: {
    view1: {
    }
}
`,
		},
		{
			name: "MixedViewsAndLayers",
			content: `a: "Entity A"
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
`,
			expected: `a: "Entity A"
b: "Entity B"
c: "Entity C"

layers: {
    view1: {
        a: "Entity A"
    }
    layer1: {
        b
    }
    view2: {
        c: "Entity C"
    }
    layer2: {
        a
        b
    }
}
`,
		},
		{
			name: "EntityWithoutLabel",
			content: `a
b: "Entity B"

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a
b: "Entity B"

layers: {
    view1: {
        a
        b: "Entity B"
    }
}
`,
		},
		{
			name: "DuplicateIdsDifferentParents",
			content: `a: {
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
`,
			expected: `a: {
    x: "X in A"
}
b: {
    x: "X in B"
}

layers: {
    view1: {
        x: "X in A"
        x: "X in B"
    }
}
`,
		},
		{
			name: "EdgeForwardArrow",
			content: `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a -> b

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a -> b
    }
}
`,
		},
		{
			name: "EdgeBackwardArrow",
			content: `a: "Entity A"
b: "Entity B"

a <- b

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a <- b

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a <- b
    }
}
`,
		},
		{
			name: "EdgeBidirectionalArrow",
			content: `a: "Entity A"
b: "Entity B"

a <-> b

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a <-> b

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a <-> b
    }
}
`,
		},
		{
			name: "EdgeNoArrows",
			content: `a: "Entity A"
b: "Entity B"

a -- b

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a -- b

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a -- b
    }
}
`,
		},
		{
			name: "EdgeWithLabel",
			content: `a: "Entity A"
b: "Entity B"

a -> b: "connection label"

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a -> b: "connection label"

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a -> b: "connection label"
    }
}
`,
		},
		{
			name: "EdgeNotIncludedWhenOneEndMissing",
			content: `a: "Entity A"
b: "Entity B"
c: "Entity C"

a -> b
b -> c

layers: {
    view1: { #view
        a
        c
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"
c: "Entity C"

a -> b
b -> c

layers: {
    view1: {
        a: "Entity A"
        c: "Entity C"
    }
}
`,
		},
		{
			name: "MultipleEdgesBetweenSameEntities",
			content: `a: "Entity A"
b: "Entity B"

a -> b: "first"
a -> b: "second"
a <- b: "third"

layers: {
    view1: { #view
        a
        b
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"

a -> b: "first"
a -> b: "second"
a <- b: "third"

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        a -> b: "first"
        a -> b: "second"
        a <- b: "third"
    }
}
`,
		},
		{
			name: "EdgeWithNestedEntities",
			content: `parent: {
    child1: "Child 1"
    child2: "Child 2"
}

parent.child1 -> parent.child2

layers: {
    view1: { #view
        parent.child1
        parent.child2
    }
}
`,
			expected: `parent: {
    child1: "Child 1"
    child2: "Child 2"
}

parent.child1 -> parent.child2

layers: {
    view1: {
        child1: "Child 1"
        child2: "Child 2"
        child1 -> child2
    }
}
`,
		},
		{
			name: "MixedEdgeDirections",
			content: `a: "Entity A"
b: "Entity B"
c: "Entity C"
d: "Entity D"

a -> b
b <- c
c <-> d
a -- d

layers: {
    view1: { #view
        a
        b
        c
        d
    }
}
`,
			expected: `a: "Entity A"
b: "Entity B"
c: "Entity C"
d: "Entity D"

a -> b
b <- c
c <-> d
a -- d

layers: {
    view1: {
        a: "Entity A"
        b: "Entity B"
        c: "Entity C"
        d: "Entity D"
        a -> b
        b <- c
        c <-> d
        a -- d
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_NestedChildOnly",
			content: `parent: "Parent" {
    child: "Child"
}

layers: {
    view1: { #view
        parent.child
    }
}
`,
			expected: `parent: "Parent" {
    child: "Child"
}

layers: {
    view1: {
        child: "Child"
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_ExplicitParentAndChild",
			content: `parent: "Parent" {
    child: "Child"
}

layers: {
    view1: { #view
        parent
        parent.child
    }
}
`,
			expected: `parent: "Parent" {
    child: "Child"
}

layers: {
    view1: {
        parent: "Parent"
        parent.child: "Child"
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_DeeplyNestedWithExplicitRoot",
			content: `a: "A" {
    b: "B" {
        c: "C" {
            d: "D"
        }
    }
}

layers: {
    view1: { #view
        a
        a.b.c.d
    }
}
`,
			expected: `a: "A" {
    b: "B" {
        c: "C" {
            d: "D"
        }
    }
}

layers: {
    view1: {
        a: "A"
        a.d: "D"
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_MultipleChildrenSameParent",
			content: `parent: "Parent" {
    child1: "Child 1"
    child2: "Child 2"
    child3: "Child 3"
}

parent.child1 -> parent.child2

layers: {
    view1: { #view
        parent.child1
        parent.child2
        parent.child3
    }
}
`,
			expected: `parent: "Parent" {
    child1: "Child 1"
    child2: "Child 2"
    child3: "Child 3"
}

parent.child1 -> parent.child2

layers: {
    view1: {
        child1: "Child 1"
        child2: "Child 2"
        child3: "Child 3"
        child1 -> child2
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_EdgeBetweenNestedAndTopLevel",
			content: `toplevel: "Top Level"
parent: "Parent" {
    nested: "Nested"
}

parent.nested -> toplevel

layers: {
    view1: { #view
        parent.nested
        toplevel
    }
}
`,
			expected: `toplevel: "Top Level"
parent: "Parent" {
    nested: "Nested"
}

parent.nested -> toplevel

layers: {
    view1: {
        nested: "Nested"
        toplevel: "Top Level"
        nested -> toplevel
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_PartialHierarchyExplicit",
			content: `a: "A" {
    b: "B" {
        c: "C"
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
			expected: `a: "A" {
    b: "B" {
        c: "C"
    }
}

layers: {
    view1: {
        a: "A"
        a.b: "B"
        a.b.c: "C"
    }
}
`,
		},
		{
			name: "ImplicitParentFiltering_MiddleAncestorExplicit",
			content: `a: "A" {
    b: "B" {
        c: "C"
    }
}

layers: {
    view1: { #view
        a.b
        a.b.c
    }
}
`,
			expected: `a: "A" {
    b: "B" {
        c: "C"
    }
}

layers: {
    view1: {
        b: "B"
        b.c: "C"
    }
}
`,
		},
		{
			name: "BaseLayerAttributes_BaseObjectAttributesPreserved",
			content: `a: "A" {
    shape: rectangle
    icon: https://example.com/icon.svg
    tooltip: "This is a tooltip"
    link: https://example.com
    width: 200
    height: 100
    near: top-center
    direction: right
    class: my-class
    style: {
        opacity: 0.8
        stroke: "#00ff00"
        fill: "#ff0000"
        fill-pattern: dots
        stroke-width: 2
        stroke-dash: 5
        border-radius: 10
        shadow: true
        3d: true
        multiple: true
        font: mono
        font-size: 16
        font-color: "#0000ff"
        animated: true
        bold: true
        italic: true
        underline: true
        filled: true
        double-border: true
        text-transform: uppercase
    }
}

layers: {
    view1: { #view
        a
    }
}
`,
			expected: `a: "A" {
    shape: rectangle
    icon: https://example.com/icon.svg
    tooltip: "This is a tooltip"
    link: https://example.com
    width: 200
    height: 100
    near: top-center
    direction: right
    class: my-class
    style: {
        opacity: 0.8
        stroke: "#00ff00"
        fill: "#ff0000"
        fill-pattern: dots
        stroke-width: 2
        stroke-dash: 5
        border-radius: 10
        shadow: true
        3d: true
        multiple: true
        font: mono
        font-size: 16
        font-color: "#0000ff"
        animated: true
        bold: true
        italic: true
        underline: true
        filled: true
        double-border: true
        text-transform: uppercase
    }
}

layers: {
    view1: {
        a: "A" {
            shape: rectangle
            icon: https://example.com/icon.svg
            tooltip: "This is a tooltip"
            link: https://example.com
            width: 200
            height: 100
            near: top-center
            direction: right
            class: my-class
            style: {
                opacity: 0.8
                stroke: "#00ff00"
                fill: "#ff0000"
                fill-pattern: dots
                stroke-width: 2
                stroke-dash: 5
                border-radius: 10
                shadow: true
                3d: true
                multiple: true
                font: mono
                font-size: 16
                font-color: "#0000ff"
                animated: true
                bold: true
                italic: true
                underline: true
                filled: true
                double-border: true
                text-transform: uppercase
            }
        }
    }
}
`,
		},
		{
			name: "ViewLayerAttributes_OverridesBaseAttributes",
			content: `a: "Base Label" {
    shape: circle
    style: {
        fill: "#ff0000"
        stroke: "#00ff00"
    }
}

layers: {
    view1: { #view
        a: "View Label" {
            shape: diamond
            style: {
                fill: "#0000ff"
            }
        }
    }
}
`,
			expected: `a: "Base Label" {
    shape: circle
    style: {
        fill: "#ff0000"
        stroke: "#00ff00"
    }
}

layers: {
    view1: {
        a: "View Label" {
            shape: diamond
            style: {
                stroke: "#00ff00"
                fill: "#0000ff"
            }
        }
    }
}
`,
		},
		{
			name: "ViewLayerAttributes_AddsNewAttributes",
			content: `a: "Entity A"

layers: {
    view1: { #view
        a {
            shape: hexagon
            style: {
                fill: "#ff0000"
            }
        }
    }
}
`,
			expected: `a: "Entity A"

layers: {
    view1: {
        a: "Entity A" {
            shape: hexagon
            style: {
                fill: "#ff0000"
            }
        }
    }
}
`,
		},
		{
			name: "ViewLayerAttributes_PartialStyleOverride",
			content: `a: "A" {
    style: {
        fill: "#ff0000"
        stroke: "#00ff00"
        opacity: 0.5
        bold: true
    }
}

layers: {
    view1: { #view
        a {
            style: {
                fill: "#0000ff"
                font-size: 20
            }
        }
    }
}
`,
			expected: `a: "A" {
    style: {
        fill: "#ff0000"
        stroke: "#00ff00"
        opacity: 0.5
        bold: true
    }
}

layers: {
    view1: {
        a: "A" {
            style: {
                opacity: 0.5
                stroke: "#00ff00"
                fill: "#0000ff"
                font-size: 20
                bold: true
            }
        }
    }
}
`,
		},
		{
			name: "ViewLayerAttributes_MultipleObjectsWithOverrides",
			content: `a: "Entity A" {
    shape: rectangle
}
b: "Entity B" {
    shape: circle
}

layers: {
    view1: { #view
        a {
            shape: diamond
        }
        b
    }
}
`,
			expected: `a: "Entity A" {
    shape: rectangle
}
b: "Entity B" {
    shape: circle
}

layers: {
    view1: {
        a: "Entity A" {
            shape: diamond
        }
        b: "Entity B" {
            shape: circle
        }
    }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			graph, _, err := compile.CompileD2("test.d2", reader)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			rootObjectIds := compile.ExtractRootObjectIds(graph)

			reader2 := strings.NewReader(tt.content)
			result, err := replaceViewLayers(reader2, graph, rootObjectIds)
			if err != nil {
				t.Fatalf("replaceViewLayers failed: %v", err)
			}

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-expected +got):\n%s", diff)
			}
		})
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

func TestApplyRangeOperations(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		ops      []rangeOperation
		expected string
	}{
		{
			name:     "EmptyOps",
			source:   "hello world",
			ops:      []rangeOperation{},
			expected: "hello world",
		},
		{
			name:     "SingleReplacement",
			source:   "hello world",
			ops:      []rangeOperation{makeOp(0, 5, "hi")}, // replace "hello" with "hi"
			expected: "hi world",
		},
		{
			name:     "SingleRemoval",
			source:   "hello world",
			ops:      []rangeOperation{makeOp(5, 11, "")}, // remove " world"
			expected: "hello",
		},
		{
			name:   "MultipleNonOverlapping",
			source: "abcdefghij",
			ops: []rangeOperation{
				makeOp(2, 4, "XX"), // replace "cd" with "XX"
				makeOp(6, 8, ""),   // remove "gh"
			},
			expected: "abXXefij",
		},
		{
			name:   "OverlappingWithReplacement",
			source: "abcdefghij",
			ops: []rangeOperation{
				makeOp(2, 5, "NEW"), // replace "cde" with "NEW"
				makeOp(4, 8, ""),    // remove "efgh" - overlaps, should merge
			},
			expected: "abNEWij", // merged range 2-8, replacement from first op
		},
		{
			name:   "OverlappingReplacementSecond",
			source: "abcdefghij",
			ops: []rangeOperation{
				makeOp(2, 5, ""),       // remove "cde"
				makeOp(4, 8, "SECOND"), // replace "efgh" - overlaps, but first op has no replacement
			},
			expected: "abSECONDij", // merged range 2-8, replacement from second op
		},
		{
			name:     "ReplaceWithLongerContent",
			source:   "abc",
			ops:      []rangeOperation{makeOp(1, 2, "LONGER")}, // replace "b" with "LONGER"
			expected: "aLONGERc",
		},
		{
			name:   "MultipleReplacementsInOrder",
			source: "one two three",
			ops: []rangeOperation{
				makeOp(0, 3, "1"),  // "one" -> "1"
				makeOp(4, 7, "2"),  // "two" -> "2"
				makeOp(8, 13, "3"), // "three" -> "3"
			},
			expected: "1 2 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyRangeOperations(tt.source, tt.ops)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
