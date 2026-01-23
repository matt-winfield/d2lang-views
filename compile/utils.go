package compile

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
)

// FindObjectById searches for an object with the given ID in the D2 graph.
// The search is case-insensitive to allow flexible referencing of objects.
func FindObjectById(graph *d2graph.Graph, id string) (*d2graph.Object, error) {
	for _, obj := range graph.Objects {
		if strings.EqualFold(GetAbsoluteId(obj), id) {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("unable to find object with id %s", id)
}

// GetAbsoluteId returns the full dot-separated path of an object by traversing its parent chain.
func GetAbsoluteId(object *d2graph.Object) string {
	var parts []string
	current := object
	for current != nil && current.Parent != nil {
		parts = append([]string{current.ID}, parts...)
		current = current.Parent
	}
	return strings.Join(parts, ".")
}
