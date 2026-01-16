package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/fatih/color"
	"oss.terrastruct.com/d2/d2graph"
)

// replaceViewLayers replaces layers marked as views in the D2 graph with the provided view contents.
func replaceViewLayers(reader io.Reader, graph *d2graph.Graph, rootObjectIds []string) (string, error) {
	contentBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	source := string(contentBytes)
	var builder strings.Builder
	builder.WriteString(source)

	views := getViewsNodes(graph)
	for _, view := range views {
		viewContent := generateViewContent(view, graph, rootObjectIds)
		builder.WriteString("\n\n")
		builder.WriteString("# View: ")
		builder.WriteString(view.Name)
		builder.WriteString("\n")
		builder.WriteString(viewContent)
	}

	return builder.String(), nil
}

// generateViewContent generates D2 language content for the given view node
//
// view is the D2 graph node representing the view
// graph is the full D2 graph (needed for getting object info that isn't included in the view)
// rootObjectIds is a list of entity IDs from the base layer to include in the view
func generateViewContent(view *d2graph.Graph, graph *d2graph.Graph, rootObjectIds []string) string {
	var builder strings.Builder
	processedIds := make(map[string]bool)

	for _, object := range view.Objects {
		// Skip container objects that have children - only include leaf nodes
		if len(object.ChildrenArray) > 0 {
			continue
		}
		objectId := getAbsoluteId(object)
		if processedIds[objectId] {
			continue
		}
		if slices.Contains(rootObjectIds, objectId) {
			builder.WriteString(getObjectD2Representation(object, graph))
			processedIds[objectId] = true
		}
	}

	return builder.String()
}

// getObjectD2Representation returns the D2 language representation of the given object.
// object is the object from the view to represent.
// graph is the full D2 graph (needed for getting the base layer object attributes, such as label).
func getObjectD2Representation(object *d2graph.Object, graph *d2graph.Graph) string {
	var builder strings.Builder

	baseObject, err := findObjectById(graph, getAbsoluteId(object))
	if err != nil {
		// This object doesn't exist in the base graph
		color.Yellow("%s", err)
		return ""
	}

	objectId := getAbsoluteId(object)
	label := getLabel(object, baseObject)
	builder.WriteString(objectId)
	if label != "" {
		builder.WriteString(": \"")
		builder.WriteString(label)
		builder.WriteString("\"")
	}
	builder.WriteString("\n")

	return builder.String()
}

// findObjectById searches for an object with the given ID in the D2 graph.
func findObjectById(graph *d2graph.Graph, id string) (*d2graph.Object, error) {
	for _, obj := range graph.Objects {
		if getAbsoluteId(obj) == id {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("unable to find object with id %s", id)
}

// getAbsoluteId returns the full dot-separated path of an object by traversing its parent chain.
func getAbsoluteId(object *d2graph.Object) string {
	var parts []string
	current := object
	for current != nil && current.Parent != nil {
		parts = append([]string{current.ID}, parts...)
		current = current.Parent
	}
	return strings.Join(parts, ".")
}

// getLabel returns the label to use for the view object.
// It prefers the view object's label, falling back to the base object's label if necessary.
// If neither has a custom label, it returns an empty string.
func getLabel(viewObject *d2graph.Object, baseObject *d2graph.Object) string {
	if viewObject.HasLabel() && viewObject.Label.Value != viewObject.ID {
		return viewObject.Label.Value
	}
	if baseObject != nil && baseObject.HasLabel() && baseObject.Label.Value != baseObject.ID {
		return baseObject.Label.Value
	}
	return ""
}
