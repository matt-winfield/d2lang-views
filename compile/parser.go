package compile

import (
	"io"
	"os"
	"path/filepath"
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
		if child.Comment == nil {
			continue
		}

		// Consecutive lines in a comment block get combined into a single comment node
		// so we need to check each line individually
		lines := strings.SplitSeq(child.Comment.Value, "\n")
		for line := range lines {
			if strings.TrimSpace(line) == "view" {
				return true
			}
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

// RemoveEdgeKey uniquely identifies an edge for removal purposes.
// Uses lowercase IDs for case-insensitive matching.
type RemoveEdgeKey struct {
	SrcID    string
	DstID    string
	SrcArrow bool
	DstArrow bool
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

// ImportCache caches parsed AST nodes from imported files to avoid re-parsing
// the same file multiple times. It should be created once per graph and reused
// across all view processing.
type ImportCache struct {
	// nodesByPath maps absolute file paths to their parsed AST nodes
	nodesByPath map[string][]d2ast.MapNodeBox
}

// NewImportCache creates a new ImportCache and pre-parses all imported files
// found in any view layer of the graph. This ensures each file is parsed only once.
func NewImportCache(graph *d2graph.Graph) *ImportCache {
	cache := &ImportCache{
		nodesByPath: make(map[string][]d2ast.MapNodeBox),
	}

	// Find all view layers and collect their imports
	layersNode := getLayersNode(graph)
	if layersNode == nil || layersNode.MapKey == nil || layersNode.MapKey.Value.Map == nil {
		return cache
	}

	for _, layerNode := range layersNode.MapKey.Value.Map.Nodes {
		if layerNode.MapKey == nil || layerNode.MapKey.Value.Map == nil {
			continue
		}
		if !isViewNode(layerNode) {
			continue
		}

		// Collect imports from this view
		cache.collectImports(layerNode.MapKey.Value.Map.Nodes)
	}

	return cache
}

// collectImports recursively parses and caches imported files.
func (c *ImportCache) collectImports(nodes []d2ast.MapNodeBox) {
	for _, node := range nodes {
		if node.Import == nil {
			continue
		}

		importPath := resolveImportPath(node.Import)
		if importPath == "" {
			continue
		}

		// Skip already cached files
		if _, exists := c.nodesByPath[importPath]; exists {
			continue
		}

		// Parse and cache the imported file
		importedNodes, err := parseImportedFile(importPath)
		if err != nil {
			// Cache empty slice to avoid retrying failed imports
			c.nodesByPath[importPath] = nil
			continue
		}

		c.nodesByPath[importPath] = importedNodes

		// Recursively collect imports from this file
		c.collectImports(importedNodes)
	}
}

// GetNodes returns the cached AST nodes for the given file path.
// Returns nil if the file was not found or failed to parse.
func (c *ImportCache) GetNodes(path string) []d2ast.MapNodeBox {
	if c == nil {
		return nil
	}
	return c.nodesByPath[path]
}

// GetOverrideEdges returns a set of edges that have the #override comment in the given view layer.
// The returned map uses edge keys for matching against base layer edges.
// This function also scans imported files within the view for override comments.
// If cache is nil, imports will be parsed on-demand (less efficient for multiple views).
func GetOverrideEdges(graph *d2graph.Graph, viewName string, cache *ImportCache) map[OverrideEdgeKey]struct{} {
	result := make(map[OverrideEdgeKey]struct{})

	// Collect all nodes from the view and its imports
	allNodeSets := collectAllViewNodes(graph, viewName, cache)

	// Extract override edges from all node sets
	for _, nodes := range allNodeSets {
		for i, node := range nodes {
			if key, ok := extractOverrideEdge(node, nodes, i); ok {
				result[key] = struct{}{}
			}
		}
	}

	return result
}

// GetRemoveEdges returns a set of edges that have the #remove comment in the given view layer.
// The returned map uses edge keys for matching against base layer edges.
// This function also scans imported files within the view for remove comments.
// If cache is nil, imports will be parsed on-demand (less efficient for multiple views).
func GetRemoveEdges(graph *d2graph.Graph, viewName string, cache *ImportCache) map[RemoveEdgeKey]struct{} {
	result := make(map[RemoveEdgeKey]struct{})

	// Collect all nodes from the view and its imports
	allNodeSets := collectAllViewNodes(graph, viewName, cache)

	// Extract remove edges from all node sets
	for _, nodes := range allNodeSets {
		for i, node := range nodes {
			if key, ok := extractRemoveEdge(node, nodes, i); ok {
				result[key] = struct{}{}
			}
		}
	}

	return result
}

// extractRemoveEdge checks if the node at index i is an edge with an inline #remove comment.
// Returns the RemoveEdgeKey and true if found, otherwise returns empty key and false.
func extractRemoveEdge(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) (RemoveEdgeKey, bool) {
	if node.MapKey == nil || len(node.MapKey.Edges) == 0 {
		return RemoveEdgeKey{}, false
	}

	if !hasInlineRemoveComment(node, nodes, i) {
		return RemoveEdgeKey{}, false
	}

	edge := node.MapKey.Edges[0]
	return RemoveEdgeKey{
		SrcID:    strings.ToLower(getKeyPathString(edge.Src)),
		DstID:    strings.ToLower(getKeyPathString(edge.Dst)),
		SrcArrow: edge.SrcArrow != "",
		DstArrow: edge.DstArrow != "",
	}, true
}

// hasInlineRemoveComment checks if the next node is a #remove comment on the same line as the edge.
func hasInlineRemoveComment(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) bool {
	if i+1 >= len(nodes) {
		return false
	}

	nextNode := nodes[i+1]
	if nextNode.Comment == nil {
		return false
	}

	if strings.TrimSpace(nextNode.Comment.Value) != "remove" {
		return false
	}

	commentLine := nextNode.Comment.Range.Start.Line
	edgeStartLine := node.MapKey.Range.Start.Line
	edgeEndLine := node.MapKey.Range.End.Line

	return commentLine == edgeStartLine || commentLine == edgeEndLine
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

// collectAllViewNodes collects all AST nodes from a view layer and its imported files.
// Returns a slice of node slices - each slice represents one file's nodes to preserve
// the node ordering needed for inline comment detection.
// If cache is provided, it will be used to retrieve pre-parsed imports.
// If cache is nil, imports will be parsed on-demand.
func collectAllViewNodes(graph *d2graph.Graph, viewName string, cache *ImportCache) [][]d2ast.MapNodeBox {
	var result [][]d2ast.MapNodeBox

	viewNodes := getViewASTNodes(graph, viewName)
	if viewNodes == nil {
		return result
	}

	// Add the direct view nodes
	result = append(result, viewNodes)

	// Recursively collect nodes from imported files
	visited := make(map[string]struct{})
	collectNodesFromImports(viewNodes, &result, visited, cache)

	return result
}

// collectNodesFromImports recursively collects AST nodes from imported files.
// Uses the cache if available, otherwise parses files on-demand.
func collectNodesFromImports(nodes []d2ast.MapNodeBox, result *[][]d2ast.MapNodeBox, visited map[string]struct{}, cache *ImportCache) {
	for _, node := range nodes {
		if node.Import == nil {
			continue
		}

		importPath := resolveImportPath(node.Import)
		if importPath == "" {
			continue
		}

		// Skip already visited files to prevent infinite recursion
		if _, seen := visited[importPath]; seen {
			continue
		}
		visited[importPath] = struct{}{}

		// Get the imported nodes from cache or parse on-demand
		var importedNodes []d2ast.MapNodeBox
		if cache != nil {
			importedNodes = cache.GetNodes(importPath)
		} else {
			var err error
			importedNodes, err = parseImportedFile(importPath)
			if err != nil {
				continue
			}
		}

		if importedNodes == nil {
			continue
		}

		// Add this file's nodes to the result
		*result = append(*result, importedNodes)

		// Recursively process imports within this file
		collectNodesFromImports(importedNodes, result, visited, cache)
	}
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

// GetIncludeParentsReferences returns a set of reference paths that have the #include-parents comment.
// Keys are stored in lowercase for case-insensitive matching.
// This function also scans imported files within the view for include-parents comments.
// If cache is nil, imports will be parsed on-demand (less efficient for multiple views).
func GetIncludeParentsReferences(graph *d2graph.Graph, viewName string, cache *ImportCache) map[string]struct{} {
	result := make(map[string]struct{})

	// Collect all nodes from the view and its imports
	allNodeSets := collectAllViewNodes(graph, viewName, cache)

	// Extract include-parents references from all node sets
	for _, nodes := range allNodeSets {
		for i, node := range nodes {
			if path := extractIncludeParentsReference(node, nodes, i); path != "" {
				result[strings.ToLower(path)] = struct{}{}
			}
		}
	}

	return result
}

// extractIncludeParentsReference checks if the node at index i is a reference with an inline #include-parents comment.
// Returns the reference path if found, otherwise returns empty string.
func extractIncludeParentsReference(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) string {
	if node.MapKey == nil || node.MapKey.Key == nil {
		return ""
	}

	// Skip nodes that define edges (we only want object references)
	if len(node.MapKey.Edges) > 0 {
		return ""
	}

	if !hasInlineIncludeParentsComment(node, nodes, i) {
		return ""
	}

	return getKeyPathString(node.MapKey.Key)
}

// hasInlineIncludeParentsComment checks if the next node is a #include-parents comment on the same line.
func hasInlineIncludeParentsComment(node d2ast.MapNodeBox, nodes []d2ast.MapNodeBox, i int) bool {
	if i+1 >= len(nodes) {
		return false
	}

	nextNode := nodes[i+1]
	if nextNode.Comment == nil {
		return false
	}

	if strings.TrimSpace(nextNode.Comment.Value) != "include-parents" {
		return false
	}

	commentLine := nextNode.Comment.Range.Start.Line
	refStartLine := node.MapKey.Range.Start.Line
	refEndLine := node.MapKey.Range.End.Line

	// Comment can be on the start line (single-line reference) or end line (multi-line block)
	return commentLine == refStartLine || commentLine == refEndLine
}

// GetIncludePatternReferences returns a slice of patterns from #include pattern=... comments.
// Patterns are stored in lowercase for case-insensitive matching.
// This function scans the view layer and imported files for include pattern comments.
// If cache is nil, imports will be parsed on-demand (less efficient for multiple views).
func GetIncludePatternReferences(graph *d2graph.Graph, viewName string, cache *ImportCache) []string {
	var result []string

	// Collect all nodes from the view and its imports
	allNodeSets := collectAllViewNodes(graph, viewName, cache)

	// Extract include patterns from all node sets
	for _, nodes := range allNodeSets {
		for _, node := range nodes {
			patterns := extractIncludeComments(node, "pattern")
			result = append(result, patterns...)
		}
	}

	return result
}

// GetIncludeClassReferences returns a slice of patterns from #include class=... comments.
// Patterns are stored in lowercase for case-insensitive matching.
// This function scans the view layer and imported files for include pattern comments.
// If cache is nil, imports will be parsed on-demand (less efficient for multiple views).
func GetIncludeClassReferences(graph *d2graph.Graph, viewName string, cache *ImportCache) []string {
	var result []string

	// Collect all nodes from the view and its imports
	allNodeSets := collectAllViewNodes(graph, viewName, cache)

	// Extract include patterns from all node sets
	for _, nodes := range allNodeSets {
		for _, node := range nodes {
			patterns := extractIncludeComments(node, "class")
			result = append(result, patterns...)
		}
	}

	return result
}

// extractIncludeComments checks if the node is a comment with #include <key>=<value>
// Returns all values found (since comments can span multiple lines with multiple patterns).
func extractIncludeComments(node d2ast.MapNodeBox, key string) []string {
	if node.Comment == nil {
		return nil
	}

	var patterns []string
	prefix := "include " + key + "="

	// Check each line in the comment (comments can span multiple lines)
	lines := strings.SplitSeq(node.Comment.Value, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if pattern, ok := strings.CutPrefix(trimmed, prefix); ok {
			patterns = append(patterns, strings.ToLower(pattern))
		}
	}

	return patterns
}

// resolveImportPath extracts and resolves the full file path from an import node.
// Returns the absolute path to the imported file, or empty string if unable to resolve.
func resolveImportPath(imp *d2ast.Import) string {
	if imp == nil || len(imp.Path) == 0 {
		return ""
	}

	// Get the source file's directory from the import's range
	sourceDir := filepath.Dir(imp.Range.Path)

	// Build the import path from path components
	var pathParts []string
	for _, part := range imp.Path {
		pathParts = append(pathParts, part.Unbox().ScalarString())
	}
	importPath := strings.Join(pathParts, "/")

	// Prepend the Pre field which contains relative path prefixes like "../"
	if imp.Pre != "" {
		importPath = imp.Pre + importPath
	}

	// Add .d2 extension if not present
	if !strings.HasSuffix(importPath, ".d2") {
		importPath += ".d2"
	}

	// Resolve relative to source file's directory
	fullPath := filepath.Join(sourceDir, importPath)

	return fullPath
}

// parseImportedFile parses a D2 file and returns its AST nodes.
func parseImportedFile(path string) (nodes []d2ast.MapNodeBox, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	graph, _, err := CompileD2(path, file)
	if err != nil {
		return nil, err
	}

	return graph.AST.Nodes, nil
}
