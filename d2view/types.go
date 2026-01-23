package d2view

import (
	"strings"

	"oss.terrastruct.com/d2/d2graph"
)

// View represents an abstraction over a D2 layer that includes specific objects and their relationships.
// A view may contain regular objects, references to base layer objects, and edges (relationships) between them.
type View struct {
	Name    string
	Label   string
	Edges   []*Edge
	Objects []*Object
	Layer   *d2graph.Graph
}

// Edge represents an edge in the view with filtered source and destination IDs.
type Edge struct {
	// Src is the source object's filtered absolute ID (after removing implicit parents).
	Src string
	// Dst is the destination object's filtered absolute ID (after removing implicit parents).
	Dst string
	// SrcArrow indicates if there's an arrow on the source side.
	SrcArrow bool
	// DstArrow indicates if there's an arrow on the destination side.
	DstArrow bool
	// D2Edge is the original D2 edge.
	D2Edge *d2graph.Edge
	// LabelOverride is set when the view uses #relabel to override the edge's label.
	// When set, this takes precedence over D2Edge.Label.
	LabelOverride string
}

type Object struct {
	// BaseObject is the original object from the base/root layer of the diagram that this view object references.
	// May be nil for objects that are defined solely within the view.
	BaseObject *d2graph.Object
	// ViewObject is the original object from the view layer of the diagram.
	ViewObject *d2graph.Object
	// ExplicitParent is the nearest explicit parent in the view hierarchy.
	// May be nil if no explicit parent exists (object is at root level after filtering).
	ExplicitParent *d2graph.Object
	// IDA is the absolute ID (dot-separated path) after filtering out implicit parents.
	IDA string
	ID  string
	// Label is the label of the object in the view, which may differ from the base object's label.
	Label string
}

// StringIDA returns the IDA representation of the object, which is the absolute identifier as a slice of strings.
// For example, an object with ID "a" in a parent object with ID "b" would return []string{"b", "a"}.
// This uses the filtered absolute ID (after filtering implicit parents) rather than the original parent chain.
func (o *Object) StringIDA() []string {
	if o.IDA == "" {
		return []string{o.ID}
	}
	return strings.Split(o.IDA, ".")
}
