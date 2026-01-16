package main

import (
	"fmt"
	"io"
	"strings"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2target"
)

// compileD2 reads from the provided io.Reader and returns the compiled D2 map.
//
// path is the file path used for error reporting
// reader is the input source containing D2 content
func compileD2(path string, reader io.Reader) (*d2graph.Graph, *d2target.Config, error) {
	compileOpts := &d2compiler.CompileOptions{}
	return d2compiler.Compile(path, reader, compileOpts)
}

// getViewsNodes extracts and returns all view nodes from the given D2 map.
func getViewsNodes(graph *d2graph.Graph) []*d2graph.Graph {
	if graph == nil || graph.AST == nil {
		return []*d2graph.Graph{}
	}

	var views []*d2graph.Graph
	var astViews []d2ast.MapNodeBox
	layersNode := getLayersNode(graph)

	if layersNode == nil || layersNode.MapKey == nil || layersNode.MapKey.Value.Map == nil {
		return views
	}

	for _, node := range layersNode.MapKey.Value.Map.Nodes {
		mapKey := node.MapKey
		if mapKey == nil || mapKey.Key == nil {
			continue
		}

		viewKey := mapKey.Key.StringIDA()
		viewName := getNodeDisplayName(node)

		fmt.Printf("Checking node %s - %s\n", viewKey, viewName)
		if isViewNode(node) {
			astViews = append(astViews, node)
		}
	}

	for _, astView := range astViews {
		for _, gNode := range graph.Layers {
			if gNode.Name == astView.MapKey.Key.StringIDA()[0] {
				views = append(views, gNode)
				break
			}
		}
	}

	return views
}

// isViewNode determines if the given D2 map node represents a view.
//
// node is a view node if it contains a top-level comment with the text "view".
func isViewNode(node d2ast.MapNodeBox) bool {
	if node.MapKey == nil || node.MapKey.Value.Map == nil {
		return false
	}

	for _, child := range node.MapKey.Value.Map.Nodes {
		if child.Comment != nil && strings.TrimSpace(child.Comment.Value) == "view" {
			return true
		}
	}

	return false
}

// getNodeDisplayName safely retrieves the display name of a D2 map node.
func getNodeDisplayName(node d2ast.MapNodeBox) string {
	if node.MapKey == nil {
		return ""
	}

	unboxed := node.MapKey.Primary.Unbox()
	if unboxed == nil {
		return strings.Join(node.MapKey.Key.StringIDA(), " ")
	}

	return node.MapKey.Primary.ScalarString()
}

// getLayersNode extracts and returns the layers node from the given D2 map.
func getLayersNode(d2graph *d2graph.Graph) *d2ast.MapNodeBox {
	for _, node := range d2graph.AST.Nodes {
		if mapKeyHasId(node, "layers") {
			return &node
		}
	}
	return nil
}

// mapKeyHasId determines if the given D2 map node represents a specific key in a map
func mapKeyHasId(node d2ast.MapNodeBox, key string) bool {
	if node.MapKey == nil || node.MapKey.Key == nil {
		return false
	}

	return node.MapKey.Key.StringIDA()[0] == key
}

// extractRootObjects extracts the entity IDs from the base layer of the D2 map.
// The retuned ID of an entity includes all parents separated by dots.
func extractRootObjects(d2graph *d2graph.Graph) []string {
	var entities = make(map[string]struct{})

	for _, child := range d2graph.Root.ChildrenArray {
		extractObjectIds(child, "", entities)
	}

	var entityList []string
	for id := range entities {
		entityList = append(entityList, id)
	}

	return entityList
}

// extractObjectIds is a helper function that recursively traverses the D2 graph nodes
// to extract entity IDs, prefixing them with their parent IDs.
// entities is a map used to collect unique entity IDs.
func extractObjectIds(node *d2graph.Object, prefix string, entities map[string]struct{}) {
	if node == nil {
		return
	}

	if prefix != "" {
		prefix = prefix + "."
	}

	currentId := prefix + node.ID
	entities[currentId] = struct{}{}

	for _, child := range node.ChildrenArray {
		extractObjectIds(child, currentId, entities)
	}
}
