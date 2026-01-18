package d2view

import "oss.terrastruct.com/d2/d2graph"

// View represents an abstraction over a D2 layer that includes specific objects and their relationships.
// A view may contain regular objects, references to base layer objects, and edges (relationships) between them.
type View struct {
	Name    string
	Edges   []*d2graph.Edge
	Objects []*Object
}

type Object struct {
	// BaseObject is the original object from the base/root layer of the diagram that this view object references.
	// May be nil for objects that are defined solely within the view.
	BaseObject *d2graph.Object
	// ViewObject is the original object from the view layer of the diagram.
	ViewObject *d2graph.Object
	ID         string
	// Label is the label of the object in the view, which may differ from the base object's label.
	Label    string
	Parent   *Object
	Children []*Object
}
