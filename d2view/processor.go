package d2view

import (
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
	return View{
		Name:    layer.Name,
		Edges:   processViewEdges(layer, graph),
		Objects: processViewObjects(layer, graph),
	}
}

// processViewObjects processes the objects within a view layer and constructs an array of Object instances.
func processViewObjects(layer *d2graph.Graph, graph *d2graph.Graph) []*Object {
	objects := make([]*Object, 0, len(layer.Objects))

	for _, obj := range layer.Objects {
		baseObj, err := compile.FindObjectById(graph, compile.GetAbsoluteId(obj))
		if err != nil {
			baseObj = nil
		}

		viewObj := &Object{
			BaseObject: baseObj,
			ViewObject: obj,
			ID:         obj.ID,
			Label:      GetLabel(obj, baseObj),
		}
		objects = append(objects, viewObj)
	}

	return objects
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
// Only edges where both the source and destination objects are present in the view layer are included.
func processViewEdges(layer *d2graph.Graph, graph *d2graph.Graph) []*d2graph.Edge {
	edges := make([]*d2graph.Edge, 0, len(layer.Edges))
	viewObjectIds := make(map[string]struct{})
	for _, obj := range layer.Objects {
		viewObjectIds[compile.GetAbsoluteId(obj)] = struct{}{}
	}

	for _, edge := range graph.Edges {
		srcId := compile.GetAbsoluteId(edge.Src)
		dstId := compile.GetAbsoluteId(edge.Dst)

		_, srcInView := viewObjectIds[srcId]
		_, dstInView := viewObjectIds[dstId]

		if srcInView && dstInView {
			edges = append(edges, edge)
		}
	}

	edges = append(edges, layer.Edges...)

	return edges
}
