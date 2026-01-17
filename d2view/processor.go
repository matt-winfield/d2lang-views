package d2view

import "oss.terrastruct.com/d2/d2graph"

// ProcessViews processes view layers defined in the D2 graph and returns an array of Views
func ProcessViews(viewLayers []*d2graph.Graph) []View {
	views := make([]View, 0)

	for _, layer := range viewLayers {
		views = append(views, processView(layer))
	}

	return views
}

// processView processes a single view layer and constructs a View object from it.
func processView(layer *d2graph.Graph) View {
	return View{
		Name:    layer.Name,
		Edges:   layer.Edges,
		Objects: processViewObjects(layer),
	}
}

// processViewObjects processes the objects within a view layer and constructs an array of Object instances.
func processViewObjects(layer *d2graph.Graph) []*Object {
	objects := make([]*Object, 0)

	for _, obj := range layer.Objects {
		viewObj := &Object{
			BaseObject: obj,
			Name:       obj.ID,
			Label:      obj.Label.Value,
		}
		objects = append(objects, viewObj)
	}

	return objects
}
