package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/matt-winfield/d2lang-views/compile"
	"github.com/matt-winfield/d2lang-views/d2view"
	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2graph"
)

// replaceViewLayers replaces layers marked as views in the D2 graph with the provided view contents.
func replaceViewLayers(reader io.Reader, graph *d2graph.Graph, rootObjectIds []string) (string, error) {
	contentBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	source := string(contentBytes)

	viewLayers := compile.GetViewsNodes(graph)
	views := d2view.ProcessViews(viewLayers, graph)
	var operations []rangeOperation
	for _, view := range views {
		viewContentRange, err := getViewContentRange(view.Layer, graph, source)
		if err != nil {
			return "", err
		}
		newContent := generateViewContent(view)
		operations = append(operations, rangeOperation{
			r:           viewContentRange.r,
			replacement: applyIndentation(newContent, viewContentRange.indentation, true),
		})
	}

	return applyRangeOperations(source, operations), nil
}

type viewContentRangeResult struct {
	r           d2ast.Range
	indentation string
}

func getViewContentRange(view *d2graph.Graph, graph *d2graph.Graph, source string) (viewContentRangeResult, error) {
	layersNode := getLayersAstNode(graph)
	if layersNode == nil || layersNode.MapKey == nil || layersNode.MapKey.Value.Map == nil {
		return viewContentRangeResult{}, fmt.Errorf("unable to find layers node in graph AST")
	}

	for _, node := range layersNode.MapKey.Value.Map.Nodes {
		mapKey := node.MapKey
		if mapKey == nil || mapKey.Key == nil {
			continue
		}

		if mapKey.Key.StringIDA()[0] == view.Name {
			indentation := getIndentationAtByte(source, node.MapKey.Range.Start.Byte)

			return viewContentRangeResult{
				r:           node.MapKey.Range,
				indentation: indentation,
			}, nil
		}
	}

	return viewContentRangeResult{}, fmt.Errorf("unable to find view content range for view %s", view.Name)
}

func getIndentationAtByte(source string, byteIndex int) string {
	indentation := ""
	if byteIndex < len(source) {
		// Find the start of the line
		lineStart := byteIndex
		for lineStart > 0 && source[lineStart-1] != '\n' {
			lineStart--
		}
		// Collect whitespace from lineStart to startByte
		for i := lineStart; i < byteIndex && i < len(source); i++ {
			if source[i] == ' ' || source[i] == '\t' {
				indentation += string(source[i])
			} else {
				break
			}
		}
	}
	return indentation
}

func getLayersAstNode(graph *d2graph.Graph) *d2ast.MapNodeBox {
	if graph == nil || graph.AST == nil {
		return nil
	}

	for _, node := range graph.AST.Nodes {
		if node.MapKey != nil && node.MapKey.Key != nil && node.MapKey.Key.StringIDA()[0] == "layers" {
			return &node
		}
	}

	return nil
}

// generateViewContent generates D2 language content for the given view node
func generateViewContent(view d2view.View) string {
	var builder strings.Builder

	builder.WriteString(getLayerDefinition(view))
	for _, object := range view.Objects {
		builder.WriteString("    ")
		builder.WriteString(getObjectD2Representation(object))
	}

	for _, edge := range view.Edges {
		builder.WriteString("    ")
		builder.WriteString(getEdgeD2Representation(edge))
	}

	builder.WriteString("}")

	return builder.String()
}

func getLayerDefinition(view d2view.View) string {
	var builder strings.Builder

	builder.WriteString(view.Name + ": ")

	if (view.Label != "") && (view.Label != view.Name) {
		builder.WriteString("\"" + view.Label + "\" ")
	}

	builder.WriteString("{\n")
	return builder.String()
}

// getEdgeD2Representation returns the D2 language representation of the given edge.
func getEdgeD2Representation(edge *d2graph.Edge) string {
	var builder strings.Builder
	srcId := compile.GetAbsoluteId(edge.Src)
	dstId := compile.GetAbsoluteId(edge.Dst)
	builder.WriteString(fmt.Sprintf("%s ", srcId))

	if edge.SrcArrow {
		builder.WriteString("<")
	}

	if edge.DstArrow || edge.SrcArrow {
		builder.WriteString("-")
	} else {
		// No arrows, use double dash for undirected edge
		builder.WriteString("--")
	}

	if edge.DstArrow {
		builder.WriteString(">")
	}
	builder.WriteString(fmt.Sprintf(" %s", dstId))

	if edge.Label.Value != "" {
		builder.WriteString(fmt.Sprintf(": \"%s\"", edge.Label.Value))
	}

	builder.WriteString("\n")
	return builder.String()
}

// applyIndentation prepends the given indentation to each line of the content.
// This ensures all inserted content maintains proper indentation within the view.
func applyIndentation(content string, indentation string, skipFirstLine bool) string {
	if indentation == "" || content == "" {
		return content
	}
	// Split content into lines, add indentation to each non-empty line
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if skipFirstLine && i == 0 {
			continue
		}

		// Don't add indentation to empty lines or the final empty line after trailing newline
		if line != "" {
			lines[i] = indentation + line
		}
	}
	return strings.Join(lines, "\n")
}

// getObjectD2Representation returns the D2 language representation of the given object.
func getObjectD2Representation(object *d2view.Object) string {
	var builder strings.Builder

	objectId := strings.Join(object.StringIDA(), ".")
	builder.WriteString(objectId)
	if object.Label != "" {
		builder.WriteString(": \"")
		builder.WriteString(object.Label)
		builder.WriteString("\"")
	}
	builder.WriteString("\n")

	return builder.String()
}

// rangeOperation represents an operation to perform on a range in the source.
// If replacement is empty, the range is removed. Otherwise, it's replaced with the content.
type rangeOperation struct {
	r           d2ast.Range
	replacement string
}

// applyRangeOperations applies a set of range operations (replacements and removals) to the source.
// Overlapping operations are merged. Operations are processed in reverse byte order to preserve indices.
func applyRangeOperations(source string, ops []rangeOperation) string {
	if len(ops) == 0 {
		return source
	}

	// Sort operations by start byte in ascending order for merging
	sortedOps := make([]rangeOperation, len(ops))
	copy(sortedOps, ops)
	slices.SortFunc(sortedOps, func(a, b rangeOperation) int {
		return a.r.Start.Byte - b.r.Start.Byte
	})

	// Merge overlapping operations
	merged := []rangeOperation{sortedOps[0]}
	for i := 1; i < len(sortedOps); i++ {
		last := &merged[len(merged)-1]
		current := sortedOps[i]
		if current.r.Start.Byte <= last.r.End.Byte {
			// Overlapping or adjacent - extend the range and combine replacements
			if current.r.End.Byte > last.r.End.Byte {
				last.r.End = current.r.End
			}
			// Keep existing replacement content (first one wins for the insertion point)
			if last.replacement == "" && current.replacement != "" {
				last.replacement = current.replacement
			}
		} else {
			merged = append(merged, current)
		}
	}

	// Apply merged operations in reverse order (end to start) to preserve indices
	result := source
	for i := len(merged) - 1; i >= 0; i-- {
		op := merged[i]
		start := op.r.Start.Byte
		end := op.r.End.Byte
		if start < 0 || end < 0 || start > len(result) || end > len(result) {
			continue
		}
		result = result[:start] + op.replacement + result[end:]
	}

	return unindentEmptyLines(result)
}

// unindentEmptyLines removes leading indentation from all whitespace-only lines in the given string.
func unindentEmptyLines(s string) string {
	lines := strings.Split(s, "\n")
	var newLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			newLines = append(newLines, "")
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
}
