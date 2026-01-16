package main

import (
	"fmt"
	"io"
	"strings"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2parser"
)

// parseD2 reads from the provided io.Reader and returns the parsed D2 map.
//
// path is the file path used for error reporting
// reader is the input source containing D2 content
func parseD2(path string, reader io.Reader) (*d2ast.Map, error) {
	opts := &d2parser.ParseOptions{}
	return d2parser.Parse(path, reader, opts)
}

// getViewsNodes extracts and returns all view nodes from the given D2 map.
func getViewsNodes(d2map *d2ast.Map) []d2ast.MapNodeBox {
	var views []d2ast.MapNodeBox
	layersNode := getLayersNode(d2map)

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
			views = append(views, node)
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
func getLayersNode(d2map *d2ast.Map) *d2ast.MapNodeBox {
	for _, node := range d2map.Nodes {
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
