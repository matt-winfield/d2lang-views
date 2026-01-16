package main

import (
	"strings"
	"testing"
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

	expected := content + "\n\n# View: view1\na: \"Entity A\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na: \"Entity A\"\n\n\n# View: view2\nb: \"Entity B\"\nc: \"Entity C\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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
	if result != content {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", content, result)
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

	expected := content + "\n\n# View: view1\na: \"Custom Label\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na.b: \"Nested B\"\nc: \"Entity C\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\nx: \"X\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na: \"Entity A\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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
	expected := content + "\n\n# View: view1\na: \"Entity A\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: backend\nserver: \"API Server\"\ndatabase: \"PostgreSQL\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na.b.c.d: \"Deep\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\nclient: \"Web Client\"\nserver: \"API Server\"\ndatabase: \"PostgreSQL\"\ncache: \"Redis\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na: \"Entity A\"\n\n\n# View: view2\nc: \"Entity C\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na\nb: \"Entity B\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
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

	expected := content + "\n\n# View: view1\na.x: \"X in A\"\nb.x: \"X in B\"\n"
	if result != expected {
		t.Errorf("unexpected result.\nexpected:\n%s\n\ngot:\n%s", expected, result)
	}
}
