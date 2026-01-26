package compile

import (
	"io"
	"strings"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2target"
)

// CompileD2 reads from the provided io.Reader and returns the compiled D2 map.
//
// path is the file path used for error reporting
// reader is the input source containing D2 content
func CompileD2(path string, reader io.Reader) (*d2graph.Graph, *d2target.Config, error) {
	compileOpts := &d2compiler.CompileOptions{}
	return d2compiler.Compile(path, reader, compileOpts)
}

// GetViewsNodes extracts and returns all view nodes from the given D2 map.
func GetViewsNodes(graph *d2graph.Graph) []*d2graph.Graph {
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

// ExtractRootObjectIds extracts the entity IDs from the base layer of the D2 map.
// The retuned ID of an entity includes all parents separated by dots.
func ExtractRootObjectIds(d2graph *d2graph.Graph) []string {
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

// OverrideEdgeKey uniquely identifies an edge for override purposes.
// Uses lowercase IDs for case-insensitive matching.
type OverrideEdgeKey struct {
	SrcID    string
	DstID    string
	SrcArrow bool
	DstArrow bool
	Label    string // The new label to apply
}

// GetOverrideEdges returns a set of edges that have the #override comment in the given view layer.
// The returned map uses edge keys for matching against base layer edges.
func GetOverrideEdges(graph *d2graph.Graph, viewName string) map[OverrideEdgeKey]struct{} {
	result := make(map[OverrideEdgeKey]struct{})

	viewNodes := getViewASTNodes(graph, viewName)
	if viewNodes == nil {
		return result
	}

	for i, node := range viewNodes {
		if key, ok := extractOverrideEdge(node, viewNodes, i); ok {
			result[key] = struct{}{}
		}
	}

	return result
}

// getViewASTNodes returns the AST nodes for a specific view layer.
func getViewASTNodes(graph *d2graph.Graph, viewName string) []d2ast.MapNodeBox {
	layersNode := getLayersNode(graph)
	if layersNode == nil || layersNode.MapKey == nil || layersNode.MapKey.Value.Map == nil {
		return nil
	}

	for _, layerNode := range layersNode.MapKey.Value.Map.Nodes {
		if !mapKeyHasId(layerNode, viewName) {
			continue
		}
		if layerNode.MapKey.Value.Map == nil {
			return nil
		}
		return layerNode.MapKey.Value.Map.Nodes
	}

	return nil
}

// extractOverrideEdge checks if the node at index i is an edge with an inline #override comment.
// Returns the OverrideEdgeKey and true if found, otherwise returns empty key and false.
func extractOverrideEdge(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) (OverrideEdgeKey, bool) {
	if node.MapKey == nil || len(node.MapKey.Edges) == 0 {
		return OverrideEdgeKey{}, false
	}

	if !hasInlineOverrideComment(node, nodes, i) {
		return OverrideEdgeKey{}, false
	}

	edge := node.MapKey.Edges[0]
	return OverrideEdgeKey{
		SrcID:    strings.ToLower(getKeyPathString(edge.Src)),
		DstID:    strings.ToLower(getKeyPathString(edge.Dst)),
		SrcArrow: edge.SrcArrow != "",
		DstArrow: edge.DstArrow != "",
		Label:    getEdgeLabelFromValue(node.MapKey.Value),
	}, true
}

// hasInlineOverrideComment checks if the next node is a #override comment on the same line as the edge.
// For single-line edges, the comment should be on the same line as the edge.
// For multi-line blocks, the comment should be on the line where the block ends.
func hasInlineOverrideComment(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) bool {
	if i+1 >= len(nodes) {
		return false
	}

	nextNode := nodes[i+1]
	if nextNode.Comment == nil {
		return false
	}

	if strings.TrimSpace(nextNode.Comment.Value) != "override" {
		return false
	}

	commentLine := nextNode.Comment.Range.Start.Line
	edgeStartLine := node.MapKey.Range.Start.Line
	edgeEndLine := node.MapKey.Range.End.Line

	// Comment can be on the start line (single-line edge) or end line (multi-line block)
	return commentLine == edgeStartLine || commentLine == edgeEndLine
}

// getKeyPathString returns the dot-separated string representation of a KeyPath.
func getKeyPathString(kp *d2ast.KeyPath) string {
	if kp == nil {
		return ""
	}
	parts := make([]string, len(kp.Path))
	for i, part := range kp.Path {
		parts[i] = part.Unbox().ScalarString()
	}
	return strings.Join(parts, ".")
}

// getEdgeLabelFromValue extracts the edge label from a ValueBox struct.
// It handles both quoted and unquoted string values.
func getEdgeLabelFromValue(value d2ast.ValueBox) string {
	if value.UnquotedString != nil {
		return value.UnquotedString.ScalarString()
	}
	if value.DoubleQuotedString != nil {
		return value.DoubleQuotedString.ScalarString()
	}
	if value.SingleQuotedString != nil {
		return value.SingleQuotedString.ScalarString()
	}
	return ""
}
