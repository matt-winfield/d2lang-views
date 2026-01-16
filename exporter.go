package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/fatih/color"
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

	views := getViewsNodes(graph)
	var operations []rangeOperation
	for _, view := range views {
		viewResult := generateViewContent(view, graph, rootObjectIds, source)
		if viewResult.newContent == "" {
			continue
		}

		// Apply indentation to the new content
		indentedContent := applyIndentation(viewResult.newContent, viewResult.indentation)

		// Insert new content at the earliest reference position (insertByte)
		// This is a pure insertion (start == end), not a replacement
		insertRange := d2ast.Range{
			Start: d2ast.Position{Byte: viewResult.insertByte},
			End:   d2ast.Position{Byte: viewResult.insertByte},
		}
		operations = append(operations, rangeOperation{
			r:           insertRange,
			replacement: indentedContent,
		})

		// Remove all replaced ranges (these are separate from the insertion point)
		for _, r := range viewResult.replacedRanges {
			operations = append(operations, rangeOperation{r: r})
		}
	}

	return applyRangeOperations(source, operations), nil
}

type viewReplacementResult struct {
	// insertByte is the byte position in the source where the new content should be inserted.
	insertByte int
	// indentation is the whitespace prefix to prepend to each line of the new content.
	indentation string
	newContent  string
	// replacedRanges holds the ranges in the original source that were replaced with view content and should be removed.
	replacedRanges []d2ast.Range
}

// generateViewContent generates D2 language content for the given view node
//
// view is the D2 graph node representing the view
// graph is the full D2 graph (needed for getting object info that isn't included in the view)
// rootObjectIds is a list of entity IDs from the base layer to include in the view
// source is the original D2 source content (used to find line endings)
func generateViewContent(view *d2graph.Graph, graph *d2graph.Graph, rootObjectIds []string, source string) viewReplacementResult {
	var builder strings.Builder
	replacedRanges := make([]d2ast.Range, 0)

	// insertByte tracks the start of the line containing the earliest reference
	// This is where the new view content will be inserted
	insertByte := len(source)
	// indentation is extracted from the source between line start and reference position
	indentation := ""

	for _, object := range view.Objects {
		objectId := getAbsoluteId(object)
		if slices.Contains(rootObjectIds, objectId) {
			builder.WriteString(getObjectD2Representation(object, graph))
			for _, reference := range object.References {
				// Track the earliest line start position for insertion
				lineStart := findLineStart(source, reference.Key.Range.Start.Byte)
				if lineStart < insertByte {
					insertByte = lineStart
					// Extract indentation (whitespace between line start and reference)
					indentation = source[lineStart:reference.Key.Range.Start.Byte]
				}

				if reference.InEdge() {
					continue
				}

				referenceContainsNonRootObject := false
				ida := reference.Key.StringIDA()
				for i := range ida {
					objectId := strings.Join(ida[:i+1], ".")
					if !slices.Contains(rootObjectIds, objectId) {
						referenceContainsNonRootObject = true
						break
					}
				}

				// Skip references that include non-root objects
				// Since those define new objects implicitly
				if referenceContainsNonRootObject {
					continue
				}

				// Extend the range to the end of the line to capture the full statement
				fullRange := extendRangeToEndOfLine(reference.Key.Range, source)
				replacedRanges = append(replacedRanges, fullRange)
			}
		}
	}

	for _, edge := range graph.Edges {
		src := getAbsoluteId(edge.Src)
		dst := getAbsoluteId(edge.Dst)

		if viewContainsObjectId(view, src) && viewContainsObjectId(view, dst) {
			// Both source and destination are in the view, include the edge
			builder.WriteString(getEdgeD2Representation(edge))
		}
	}

	return viewReplacementResult{
		newContent:     builder.String(),
		replacedRanges: replacedRanges,
		insertByte:     insertByte,
		indentation:    indentation,
	}
}

// getEdgeD2Representation returns the D2 language representation of the given edge.
func getEdgeD2Representation(edge *d2graph.Edge) string {
	var builder strings.Builder
	srcId := getAbsoluteId(edge.Src)
	dstId := getAbsoluteId(edge.Dst)
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

// viewContainsObjectId checks if the given view contains an object with the specified absolute ID.
func viewContainsObjectId(view *d2graph.Graph, objectId string) bool {
	for _, obj := range view.Objects {
		if getAbsoluteId(obj) == objectId {
			return true
		}
	}
	return false
}

// extendRangeToEndOfLine extends a range to the end of the line (newline character)
// This captures the full statement including any label/value after the key
func extendRangeToEndOfLine(r d2ast.Range, source string) d2ast.Range {
	endByte := r.End.Byte
	// Find the next newline from the end position
	for endByte < len(source) && source[endByte] != '\n' {
		endByte++
	}
	return d2ast.Range{
		Path:  r.Path,
		Start: r.Start,
		End: d2ast.Position{
			Line:   r.End.Line,
			Column: r.End.Column + (endByte - r.End.Byte),
			Byte:   endByte,
		},
	}
}

// applyIndentation prepends the given indentation to each line of the content.
// This ensures all inserted content maintains proper indentation within the view.
func applyIndentation(content string, indentation string) string {
	if indentation == "" || content == "" {
		return content
	}
	// Split content into lines, add indentation to each non-empty line
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Don't add indentation to empty lines or the final empty line after trailing newline
		if line != "" {
			lines[i] = indentation + line
		}
	}
	return strings.Join(lines, "\n")
}

// findLineStart finds the byte position of the start of the line containing the given byte position.
// It returns the position right after the previous newline (or 0 if at the start of the file).
func findLineStart(source string, bytePos int) int {
	if bytePos <= 0 {
		return 0
	}
	// Search backwards for the newline
	for i := bytePos - 1; i >= 0; i-- {
		if source[i] == '\n' {
			return i + 1
		}
	}
	return 0
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
