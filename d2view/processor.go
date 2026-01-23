package d2view

import (
	"strings"

	"github.com/matt-winfield/d2lang-views/compile"
	"oss.terrastruct.com/d2/d2graph"
)

// ProcessViews processes view layers defined in the D2 graph and returns an array of Views.
// viewLayers is an array of D2 graph layers that represent views.
// graph is the complete D2 graph, which is used to reference base layer objects.
func ProcessViews(viewLayers []*d2graph.Graph, graph *d2graph.Graph) []View {
	views := make([]View, 0)

	for _, layer := range viewLayers {
		views = append(views, processView(layer, graph))
	}

	return views
}

// processView processes a single view layer and constructs a View object from it.
func processView(layer *d2graph.Graph, graph *d2graph.Graph) View {
	explicitIds := getExplicitObjectIds(layer)

	return View{
		Name:    layer.Name,
		Label:   layer.Root.Label.Value,
		Edges:   processViewEdges(layer, graph, explicitIds),
		Objects: processViewObjects(layer, graph, explicitIds),
		Layer:   layer,
	}
}

// processViewObjects processes the objects within a view layer and constructs an array of Object instances.
// It filters out implicit parent objects and adjusts the parent chain for remaining objects.
func processViewObjects(layer *d2graph.Graph, graph *d2graph.Graph, explicitIds map[string]struct{}) []*Object {
	objects := make([]*Object, 0, len(layer.Objects))
	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		if _, isExplicit := explicitIds[strings.ToLower(absId)]; !isExplicit {
			continue // Skip implicit parent objects
		}
		viewObj := processViewObject(obj, graph, explicitIds)
		objects = append(objects, viewObj)
	}

	return objects
}

// getExplicitObjectIds returns a set of absolute IDs for objects that are explicitly referenced in the layer.
// An object is explicit if it has at least one reference where the reference path length equals the object's path depth.
// Keys are stored in lowercase for case-insensitive matching.
func getExplicitObjectIds(layer *d2graph.Graph) map[string]struct{} {
	explicitIds := make(map[string]struct{})

	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		pathLen := getPathDepth(absId)

		for _, ref := range obj.References {
			if len(ref.Key.Path) == pathLen {
				explicitIds[strings.ToLower(absId)] = struct{}{}
				break
			}
		}
	}

	return explicitIds
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
func GetLabel(viewObject *d2graph.Object, baseObject *d2graph.Object) string {
	if viewObject.HasLabel() && viewObject.Label.Value != viewObject.ID {
		return viewObject.Label.Value
	}
	if baseObject != nil && baseObject.HasLabel() && baseObject.Label.Value != baseObject.ID {
		return baseObject.Label.Value
	}
	return ""
}

// processViewEdges extracts the edges from the base layer that are relevant to the view layer.
// Only edges where both the source and destination objects are explicitly present in the view layer are included.
// Edge source/destination IDs are adjusted to reflect the filtered object hierarchy.
// Edges with #relabel in the view layer override the labels of matching base edges.
func processViewEdges(layer *d2graph.Graph, graph *d2graph.Graph, explicitIds map[string]struct{}) []*Edge {
	// Build a mapping from original absolute IDs to filtered absolute IDs
	idMapping := buildFilteredIdMapping(layer, explicitIds)

	// Get the relabel edges for this view - we'll track which ones get applied
	relabelEdges := compile.GetRelabelEdges(graph, layer.Name)
	appliedRelabels := make(map[compile.RelabelEdgeKey]struct{})

	edges := make([]*Edge, 0, len(layer.Edges))

	// Process edges from the base graph
	for _, edge := range graph.Edges {
		srcId := compile.GetAbsoluteId(edge.Src)
		dstId := compile.GetAbsoluteId(edge.Dst)

		filteredSrcId, srcInView := idMapping[strings.ToLower(srcId)]
		filteredDstId, dstInView := idMapping[strings.ToLower(dstId)]

		if srcInView && dstInView {
			newEdge := &Edge{
				Src:      filteredSrcId,
				Dst:      filteredDstId,
				SrcArrow: edge.SrcArrow,
				DstArrow: edge.DstArrow,
				D2Edge:   edge,
			}

			// Check if there's a relabel for this edge
			relabelKey, labelOverride := findRelabelForEdge(srcId, dstId, edge.SrcArrow, edge.DstArrow, relabelEdges)
			if labelOverride != "" {
				newEdge.LabelOverride = labelOverride
				// Mark this relabel as applied
				appliedRelabels[relabelKey] = struct{}{}
				// Remove the relabel entry so it's only applied once (to the first matching edge)
				delete(relabelEdges, relabelKey)
			}

			edges = append(edges, newEdge)
		}
	}

	// Process edges defined in the view layer
	allRelabelEdges := compile.GetRelabelEdges(graph, layer.Name)
	for _, edge := range layer.Edges {
		srcId := compile.GetAbsoluteId(edge.Src)
		dstId := compile.GetAbsoluteId(edge.Dst)

		filteredSrcId, srcExists := idMapping[strings.ToLower(srcId)]
		filteredDstId, dstExists := idMapping[strings.ToLower(dstId)]

		if srcExists && dstExists {
			// Check if this is a relabel edge
			relabelKey, isRelabel := getRelabelKey(srcId, dstId, edge.SrcArrow, edge.DstArrow, allRelabelEdges)
			if isRelabel {
				// Only skip if this relabel was applied to a base edge
				if _, wasApplied := appliedRelabels[relabelKey]; wasApplied {
					continue
				}
				// Otherwise, add it as a new edge with the relabel as its label
				edges = append(edges, &Edge{
					Src:           filteredSrcId,
					Dst:           filteredDstId,
					SrcArrow:      edge.SrcArrow,
					DstArrow:      edge.DstArrow,
					D2Edge:        edge,
					LabelOverride: relabelKey.Label,
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

// findRelabelForEdge checks if there's a matching relabel entry for the given edge.
// Returns the key and override label if found, empty key and string otherwise.
func findRelabelForEdge(srcId, dstId string, srcArrow, dstArrow bool, relabelEdges map[compile.RelabelEdgeKey]struct{}) (compile.RelabelEdgeKey, string) {
	for key := range relabelEdges {
		if strings.ToLower(srcId) == key.SrcID &&
			strings.ToLower(dstId) == key.DstID &&
			srcArrow == key.SrcArrow &&
			dstArrow == key.DstArrow {
			return key, key.Label
		}
	}
	return compile.RelabelEdgeKey{}, ""
}

// getRelabelKey checks if the given edge parameters match a relabel edge and returns the key.
func getRelabelKey(srcId, dstId string, srcArrow, dstArrow bool, relabelEdges map[compile.RelabelEdgeKey]struct{}) (compile.RelabelEdgeKey, bool) {
	for key := range relabelEdges {
		if strings.ToLower(srcId) == key.SrcID &&
			strings.ToLower(dstId) == key.DstID &&
			srcArrow == key.SrcArrow &&
			dstArrow == key.DstArrow {
			return key, true
		}
	}
	return compile.RelabelEdgeKey{}, false
}

// buildFilteredIdMapping creates a mapping from original absolute IDs to filtered absolute IDs.
// Only explicit objects are included in the mapping.
// Keys are stored in lowercase for case-insensitive matching.
func buildFilteredIdMapping(layer *d2graph.Graph, explicitIds map[string]struct{}) map[string]string {
	idMapping := make(map[string]string)

	for _, obj := range layer.Objects {
		absId := compile.GetAbsoluteId(obj)
		if _, isExplicit := explicitIds[strings.ToLower(absId)]; !isExplicit {
			continue
		}

		// Build the filtered ID by walking up the parent chain and only including explicit parents
		filteredId := buildFilteredAbsoluteId(obj, explicitIds)
		idMapping[strings.ToLower(absId)] = filteredId
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
