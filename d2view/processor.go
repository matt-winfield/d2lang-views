package d2view

import (
	"strings"

	"github.com/matt-winfield/d2lang-views/compile"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
)

// ProcessViews processes view layers defined in the D2 graph and returns an array of Views.
// viewLayers is an array of D2 graph layers that represent views.
// graph is the complete D2 graph, which is used to reference base layer objects.
func ProcessViews(viewLayers []*d2graph.Graph, graph *d2graph.Graph) []View {
	views := make([]View, 0)

	// Create import cache once for all views to avoid re-parsing imported files
	importCache := compile.NewImportCache(graph)

	for _, layer := range viewLayers {
		views = append(views, processView(layer, graph, importCache))
	}

	return views
}

// processView processes a single view layer and constructs a View object from it.
func processView(layer *d2graph.Graph, graph *d2graph.Graph, importCache *compile.ImportCache) View {
	includeParentsRefs := compile.GetIncludeParentsReferences(graph, layer.Name, importCache)
	includePatterns := compile.GetIncludePatternReferences(graph, layer.Name, importCache)
	includeClasses := compile.GetIncludeClassReferences(graph, layer.Name, importCache)
	rootObjectIds := compile.ExtractRootObjectIds(graph)
	explicitIds := getExplicitObjectIds(layer, graph, includeParentsRefs, includePatterns, includeClasses, rootObjectIds)

	return View{
		Name:    layer.Name,
		Label:   layer.Root.Label.Value,
		Edges:   processViewEdges(layer, graph, explicitIds, importCache),
		Objects: processViewObjects(layer, graph, explicitIds),
		Layer:   layer,
	}
}

// processViewObjects processes the objects within a view layer and constructs an array of Object instances.
// It filters out implicit parent objects and adjusts the parent chain for remaining objects.
// Objects that are explicit via pattern matching but not in the view layer are also included from the base graph.
func processViewObjects(layer *d2graph.Graph, graph *d2graph.Graph, explicitIds map[string]struct{}) []*Object {
	objects := make([]*Object, 0, len(layer.Objects))
	processedIds := make(map[string]struct{})

	// First, process objects from the view layer
	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		if _, isExplicit := explicitIds[strings.ToLower(absId)]; !isExplicit {
			continue // Skip implicit parent objects
		}
		viewObj := processViewObject(obj, graph, explicitIds)
		objects = append(objects, viewObj)
		processedIds[strings.ToLower(absId)] = struct{}{}
	}

	// Then, process objects from the base graph that are explicit (via pattern matching)
	// but not already processed from the view layer
	for _, obj := range graph.Objects {
		absId := compile.GetAbsoluteId(obj)
		absIdLower := strings.ToLower(absId)

		// Skip if not explicit or already processed
		if _, isExplicit := explicitIds[absIdLower]; !isExplicit {
			continue
		}
		if _, alreadyProcessed := processedIds[absIdLower]; alreadyProcessed {
			continue
		}

		viewObj := processBaseGraphObject(obj, explicitIds)
		objects = append(objects, viewObj)
		processedIds[absIdLower] = struct{}{}
	}

	return objects
}

// processBaseGraphObject processes an object from the base graph (not present in view layer).
// Used for pattern-matched objects that need to be included in the view.
func processBaseGraphObject(obj *d2graph.Object, explicitIds map[string]struct{}) *Object {
	// Find the new parent by walking up the chain until we find an explicit parent or reach root
	newParent := findExplicitParent(obj.Parent, explicitIds)

	// Build the filtered absolute ID
	filteredAbsId := buildFilteredAbsoluteId(obj, explicitIds)

	label := ""
	if obj.Label.Value != obj.ID {
		label = getObjectLabel(obj)
	}

	viewObj := &Object{
		BaseObject:     obj,
		ViewObject:     obj, // Same as base since it's not in view layer
		ID:             obj.ID,
		Label:          label,
		ExplicitParent: newParent,
		IDA:            filteredAbsId,
	}
	return viewObj
}

// getExplicitObjectIds returns a set of absolute IDs for objects that are explicitly referenced in the layer.
// An object is explicit if it has at least one reference where the reference path length equals the object's path depth.
// When a reference has the #include-parents comment, all ancestors in that reference path are also marked as explicit.
// When include patterns are provided, all root objects matching the patterns are also marked as explicit,
// along with all ancestors of the pattern prefix (similar to include-parents behavior).
// Keys are stored in lowercase for case-insensitive matching.
func getExplicitObjectIds(layer *d2graph.Graph, graph *d2graph.Graph, includeParentsRefs map[string]struct{}, includePatterns []string, includeClasses []string, rootObjectIds []string) map[string]struct{} {
	explicitIds := make(map[string]struct{})

	// Process include patterns - add matching root objects and their ancestors as explicit
	for _, pattern := range includePatterns {
		// Mark all matching objects as explicit
		for _, rootId := range rootObjectIds {
			if matchesIncludePattern(strings.ToLower(rootId), pattern) {
				explicitIds[strings.ToLower(rootId)] = struct{}{}
			}
		}
	}

	// Process include classes - add matching root objects and their ancestors as explicit
	for _, pattern := range includeClasses {
		// Mark all matching objects as explicit
		for _, rootId := range rootObjectIds {
			if matchesClass(strings.ToLower(rootId), pattern, graph) {
				explicitIds[strings.ToLower(rootId)] = struct{}{}
			}
		}
	}

	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		pathLen := getPathDepth(absId)

		for _, ref := range obj.References {
			// Convert reference path to string
			refPathParts := make([]string, len(ref.Key.Path))
			for i, part := range ref.Key.Path {
				refPathParts[i] = part.Unbox().ScalarString()
			}
			refPath := strings.Join(refPathParts, ".")

			// Check if this reference has the #include-parents comment
			_, hasIncludeParents := includeParentsRefs[strings.ToLower(refPath)]

			if len(ref.Key.Path) == pathLen {
				explicitIds[strings.ToLower(absId)] = struct{}{}

				// If #include-parents is present, mark all ancestors as explicit
				if hasIncludeParents {
					markAncestorsAsExplicit(obj, explicitIds)
				}
				break
			}
		}
	}

	return explicitIds
}

// matchesIncludePattern checks if an object ID matches the given pattern.
// The pattern format is "prefix.*" which matches the prefix itself and all children.
// For example, "cf.*" matches "cf", "cf.stack1", "cf.stack1.resource1", etc.
func matchesIncludePattern(objectId, pattern string) bool {
	// a.b matches a.b.*
	if strings.HasPrefix(pattern, objectId) {
		return true
	}

	return matchesWildcard(pattern, objectId)
}

// matchesClass checks if an object ID matches the given class pattern.
func matchesClass(objectId, class string, graph *d2graph.Graph) bool {
	obj, err := compile.FindObjectById(graph, objectId)
	if err != nil || obj == nil {
		return false
	}

	// Check if the object has the specified class
	for _, objClass := range obj.Classes {
		if matchesWildcard(strings.ToLower(class), strings.ToLower(objClass)) {
			return true
		}
	}

	return false
}

// matchesWildcard checks if a string matches a wildcard pattern.
// The pattern can contain '*' which matches any sequence of characters (including empty sequences).
//
// Examples:
//   - matchesWildcard("hello*", "helloworld") -> true
//   - matchesWildcard("*world", "helloworld") -> true
//   - matchesWildcard("a*b*c", "aXbYc") -> true
//   - matchesWildcard("*", "anything") -> true
//
// Algorithm:
// 1. Split pattern by '*' to get literal segments that must appear in order
// 2. Check if string starts with first segment (if not empty due to leading '*')
// 3. For each middle segment, find its first occurrence and advance past it
// 4. Check if string ends with last segment (if not empty due to trailing '*')
//
// Time Complexity: O(n*m) where n = len(str), m = number of pattern segments
// Space Complexity: O(m) for storing split pattern parts
func matchesWildcard(pattern, str string) bool {
	// Optimization: single wildcard matches everything
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")

	// No wildcards - require exact match
	if len(parts) == 1 {
		return pattern == str
	}

	// Check prefix (first part before any wildcard)
	if len(parts[0]) > 0 {
		if !strings.HasPrefix(str, parts[0]) {
			return false
		}
		str = str[len(parts[0]):]
	}

	// Check middle parts - each must appear in order
	for i := 1; i < len(parts)-1; i++ {
		// Skip empty parts from consecutive wildcards (e.g., "a**b" creates ["a", "", "b"])
		if len(parts[i]) == 0 {
			continue
		}

		idx := strings.Index(str, parts[i])
		if idx == -1 {
			return false
		}
		// Advance past this match
		str = str[idx+len(parts[i]):]
	}

	// Check suffix (last part after final wildcard)
	lastPart := parts[len(parts)-1]
	if len(lastPart) > 0 {
		return strings.HasSuffix(str, lastPart) && len(str) >= len(lastPart)
	}

	return true
}

// markAncestorsAsExplicit walks up the parent chain and marks all ancestors as explicit.
func markAncestorsAsExplicit(obj *d2graph.Object, explicitIds map[string]struct{}) {
	current := obj.Parent
	for current != nil {
		absId := compile.GetAbsoluteId(current)
		if absId == "" {
			break // Reached root
		}
		explicitIds[strings.ToLower(absId)] = struct{}{}
		current = current.Parent
	}
}

// getPathDepth returns the depth of a dot-separated path (number of segments).
func getPathDepth(absId string) int {
	if absId == "" {
		return 0
	}
	return len(strings.Split(absId, "."))
}

// processViewObject processes a single view object and creates an Object instance.
// It also adjusts the parent chain to skip any filtered implicit parents.
func processViewObject(obj *d2graph.Object, graph *d2graph.Graph, explicitIds map[string]struct{}) *Object {
	baseObj, err := compile.FindObjectById(graph, compile.GetAbsoluteId(obj))
	if err != nil {
		baseObj = nil
	}

	// Find the new parent by walking up the chain until we find an explicit parent or reach root
	newParent := findExplicitParent(obj.Parent, explicitIds)

	// Build the filtered absolute ID
	filteredAbsId := buildFilteredAbsoluteId(obj, explicitIds)

	viewObj := &Object{
		BaseObject:     baseObj,
		ViewObject:     obj,
		ID:             obj.ID,
		Label:          GetLabel(obj, baseObj),
		ExplicitParent: newParent,
		IDA:            filteredAbsId,
	}
	return viewObj
}

// findExplicitParent walks up the parent chain to find the first explicit parent.
// Returns nil if no explicit parent is found.
func findExplicitParent(parent *d2graph.Object, explicitIds map[string]struct{}) *d2graph.Object {
	for parent != nil {
		absId := compile.GetAbsoluteId(parent)
		if absId == "" {
			return nil // Reached the root
		}
		if _, isExplicit := explicitIds[strings.ToLower(absId)]; isExplicit {
			return parent
		}
		parent = parent.Parent
	}
	return nil
}

// GetLabel returns the label to use for the view object.
// It prefers the view object's label, falling back to the base object's label if necessary.
// If neither has a label, it returns the view object's ID.
func GetLabel(viewObject, baseObject *d2graph.Object) string {
	if viewObject.Label.Value != viewObject.ID {
		return getObjectLabel(viewObject)
	}
	if baseObject != nil && baseObject.Label.Value != baseObject.ID {
		return getObjectLabel(baseObject)
	}
	return ""
}

func getObjectLabel(obj *d2graph.Object) string {
	if obj.Language != "" {
		alias := d2compiler.FullToShortLanguageAliases[obj.Language]
		if alias != "" {
			return "|" + alias + "\n" + obj.Label.Value + "\n|"
		}
		return "|" + obj.Language + "\n" + obj.Label.Value + "\n|"
	}
	return obj.Label.Value
}

// processViewEdges extracts the edges from the base layer that are relevant to the view layer.
// Only edges where both the source and destination objects are explicitly present in the view layer are included.
// Edge source/destination IDs are adjusted to reflect the filtered object hierarchy.
// Edges with #override in the view layer override the labels of matching base edges.
func processViewEdges(layer *d2graph.Graph, graph *d2graph.Graph, explicitIds map[string]struct{}, importCache *compile.ImportCache) []*Edge {
	// Build a mapping from original absolute IDs to filtered absolute IDs
	idMapping := buildFilteredIdMapping(layer, graph, explicitIds)

	// Get the remove edges for this view
	removeEdges := compile.GetRemoveEdges(graph, layer.Name, importCache)

	// Get the override edges for this view - we'll track which ones get applied
	overrideEdges := compile.GetOverrideEdges(graph, layer.Name, importCache)
	appliedOverrides := make(map[compile.OverrideEdgeKey]struct{})

	edges := make([]*Edge, 0, len(layer.Edges))

	// Process edges from the base graph
	for _, edge := range graph.Edges {
		srcId := compile.GetAbsoluteId(edge.Src)
		dstId := compile.GetAbsoluteId(edge.Dst)

		filteredSrcId, srcInView := idMapping[strings.ToLower(srcId)]
		filteredDstId, dstInView := idMapping[strings.ToLower(dstId)]

		if srcInView && dstInView {
			// Check if this edge should be removed
			if isRemovedEdge(srcId, dstId, edge.SrcArrow, edge.DstArrow, removeEdges) {
				continue
			}

			newEdge := &Edge{
				Src:      filteredSrcId,
				Dst:      filteredDstId,
				SrcArrow: edge.SrcArrow,
				DstArrow: edge.DstArrow,
				D2Edge:   edge,
			}

			// Check if there's an override for this edge
			overrideKey, hasOverride := getOverrideKey(srcId, dstId, edge.SrcArrow, edge.DstArrow, overrideEdges)
			if hasOverride {
				// Find and store the view edge for property merging
				viewEdge := findViewEdge(layer, srcId, dstId, edge.SrcArrow, edge.DstArrow)
				newEdge.ViewEdge = viewEdge

				// Apply label override from view edge if it has a different label
				if viewEdge != nil && viewEdge.Label.Value != "" {
					newEdge.LabelOverride = viewEdge.Label.Value
				}

				// Mark this override as applied
				appliedOverrides[overrideKey] = struct{}{}
				// Remove the override entry so it's only applied once (to the first matching edge)
				delete(overrideEdges, overrideKey)
			}

			edges = append(edges, newEdge)
		}
	}

	// Process edges defined in the view layer
	allOverrideEdges := compile.GetOverrideEdges(graph, layer.Name, importCache)
	allRemoveEdges := compile.GetRemoveEdges(graph, layer.Name, importCache)
	for _, edge := range layer.Edges {
		srcId := compile.GetAbsoluteId(edge.Src)
		dstId := compile.GetAbsoluteId(edge.Dst)

		filteredSrcId, srcExists := idMapping[strings.ToLower(srcId)]
		filteredDstId, dstExists := idMapping[strings.ToLower(dstId)]

		if srcExists && dstExists {
			// Skip edges marked with #remove (they should not be added as new edges)
			if isRemovedEdge(srcId, dstId, edge.SrcArrow, edge.DstArrow, allRemoveEdges) {
				continue
			}

			// Check if this is an override edge
			overrideKey, isOverride := getOverrideKey(srcId, dstId, edge.SrcArrow, edge.DstArrow, allOverrideEdges)
			if isOverride {
				// Only skip if this override was applied to a base edge
				if _, wasApplied := appliedOverrides[overrideKey]; wasApplied {
					continue
				}
				// Otherwise, add it as a new edge with the override as its label
				edges = append(edges, &Edge{
					Src:           filteredSrcId,
					Dst:           filteredDstId,
					SrcArrow:      edge.SrcArrow,
					DstArrow:      edge.DstArrow,
					D2Edge:        edge,
					LabelOverride: overrideKey.Label,
				})
				continue
			}

			edges = append(edges, &Edge{
				Src:      filteredSrcId,
				Dst:      filteredDstId,
				SrcArrow: edge.SrcArrow,
				DstArrow: edge.DstArrow,
				D2Edge:   edge,
			})
		}
	}

	return edges
}

// isRemovedEdge checks if the given edge parameters match a remove edge.
func isRemovedEdge(srcId, dstId string, srcArrow, dstArrow bool, removeEdges map[compile.RemoveEdgeKey]struct{}) bool {
	for key := range removeEdges {
		if strings.ToLower(srcId) == key.SrcID &&
			strings.ToLower(dstId) == key.DstID &&
			srcArrow == key.SrcArrow &&
			dstArrow == key.DstArrow {
			return true
		}
	}
	return false
}

// getOverrideKey checks if the given edge parameters match an override edge and returns the key.
func getOverrideKey(srcId, dstId string, srcArrow, dstArrow bool, overrideEdges map[compile.OverrideEdgeKey]struct{}) (compile.OverrideEdgeKey, bool) {
	for key := range overrideEdges {
		if strings.ToLower(srcId) == key.SrcID &&
			strings.ToLower(dstId) == key.DstID &&
			srcArrow == key.SrcArrow &&
			dstArrow == key.DstArrow {
			return key, true
		}
	}
	return compile.OverrideEdgeKey{}, false
}

// findViewEdge finds the view layer edge that matches the given edge parameters.
// Used to retrieve the view edge for property merging when an override is applied.
func findViewEdge(layer *d2graph.Graph, srcId, dstId string, srcArrow, dstArrow bool) *d2graph.Edge {
	srcIdLower := strings.ToLower(srcId)
	dstIdLower := strings.ToLower(dstId)

	for _, edge := range layer.Edges {
		edgeSrcId := strings.ToLower(compile.GetAbsoluteId(edge.Src))
		edgeDstId := strings.ToLower(compile.GetAbsoluteId(edge.Dst))

		if edgeSrcId == srcIdLower &&
			edgeDstId == dstIdLower &&
			edge.SrcArrow == srcArrow &&
			edge.DstArrow == dstArrow {
			return edge
		}
	}
	return nil
}

// buildFilteredIdMapping creates a mapping from original absolute IDs to filtered absolute IDs.
// Only explicit objects are included in the mapping.
// Keys are stored in lowercase for case-insensitive matching.
// This function considers both view layer objects and base graph objects (for pattern matching).
func buildFilteredIdMapping(layer *d2graph.Graph, graph *d2graph.Graph, explicitIds map[string]struct{}) map[string]string {
	idMapping := make(map[string]string)
	processedIds := make(map[string]struct{})

	// Process view layer objects
	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		absIdLower := strings.ToLower(absId)
		if _, isExplicit := explicitIds[absIdLower]; !isExplicit {
			continue
		}

		filteredId := buildFilteredAbsoluteId(obj, explicitIds)
		idMapping[absIdLower] = filteredId
		processedIds[absIdLower] = struct{}{}
	}

	// Process base graph objects that are explicit but not in view layer
	for _, obj := range graph.Objects {
		absId := compile.GetAbsoluteId(obj)
		absIdLower := strings.ToLower(absId)

		// Skip if not explicit or already processed
		if _, isExplicit := explicitIds[absIdLower]; !isExplicit {
			continue
		}
		if _, alreadyProcessed := processedIds[absIdLower]; alreadyProcessed {
			continue
		}

		filteredId := buildFilteredAbsoluteId(obj, explicitIds)
		idMapping[absIdLower] = filteredId
	}

	return idMapping
}

// buildFilteredAbsoluteId constructs the absolute ID for an object after filtering out implicit parents.
func buildFilteredAbsoluteId(obj *d2graph.Object, explicitIds map[string]struct{}) string {
	var parts []string
	parts = append(parts, obj.ID)

	current := obj.Parent
	for current != nil {
		absId := compile.GetAbsoluteId(current)
		if absId == "" {
			break // Reached root
		}
		if _, isExplicit := explicitIds[strings.ToLower(absId)]; isExplicit {
			parts = append([]string{current.ID}, parts...)
		}
		current = current.Parent
	}

	return strings.Join(parts, ".")
}
