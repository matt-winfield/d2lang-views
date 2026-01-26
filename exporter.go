package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/matt-winfield/d2lang-views/compile"
	"github.com/matt-winfield/d2lang-views/d2view"
	"github.com/matt-winfield/d2lang-views/version"
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

	result := applyRangeOperations(source, operations)

	// Add file header comment at the start of the generated output
	return version.GeneratedFileHeader() + result, nil
}

type viewContentRangeResult struct {
	r           d2ast.Range
	indentation string
}

// getViewContentRange finds the range in the source D2 content that corresponds to the given view's layer definition.
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

	// Add generated view header comment
	builder.WriteString("    ")
	builder.WriteString(version.GeneratedViewHeader())

	// Output view-level keywords (direction, etc.)
	builder.WriteString(getViewLevelKeywordsRepresentation(view))

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

// getViewLevelKeywordsRepresentation returns the D2 representation of view-level keywords
// such as direction, grid-rows, grid-columns, etc.
func getViewLevelKeywordsRepresentation(view d2view.View) string {
	if view.Layer == nil || view.Layer.Root == nil {
		return ""
	}

	var builder strings.Builder
	root := view.Layer.Root

	// Direction keyword
	if root.Direction.Value != "" && root.Direction.MapKey != nil {
		builder.WriteString(fmt.Sprintf("    direction: %s\n", root.Direction.Value))
	}

	// Grid-related keywords
	if root.GridRows != nil && root.GridRows.Value != "" {
		builder.WriteString(fmt.Sprintf("    grid-rows: %s\n", root.GridRows.Value))
	}

	if root.GridColumns != nil && root.GridColumns.Value != "" {
		builder.WriteString(fmt.Sprintf("    grid-columns: %s\n", root.GridColumns.Value))
	}

	if root.GridGap != nil && root.GridGap.Value != "" {
		builder.WriteString(fmt.Sprintf("    grid-gap: %s\n", root.GridGap.Value))
	}

	if root.HorizontalGap != nil && root.HorizontalGap.Value != "" {
		builder.WriteString(fmt.Sprintf("    horizontal-gap: %s\n", root.HorizontalGap.Value))
	}

	if root.VerticalGap != nil && root.VerticalGap.Value != "" {
		builder.WriteString(fmt.Sprintf("    vertical-gap: %s\n", root.VerticalGap.Value))
	}

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
func getEdgeD2Representation(edge *d2view.Edge) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s ", edge.Src))

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
	builder.WriteString(fmt.Sprintf(" %s", edge.Dst))

	// LabelOverride takes precedence over D2Edge.Label
	if edge.LabelOverride != "" {
		builder.WriteString(fmt.Sprintf(": \"%s\"", edge.LabelOverride))
	} else if edge.D2Edge != nil && edge.D2Edge.Label.Value != "" {
		builder.WriteString(fmt.Sprintf(": \"%s\"", edge.D2Edge.Label.Value))
	}

	// Add edge attributes if present
	attrs := getEdgeAttributesRepresentation(edge)
	if attrs != "" {
		builder.WriteString(" {\n")
		builder.WriteString(attrs)
		builder.WriteString("    }")
	}

	builder.WriteString("\n")
	return builder.String()
}

// getEdgeAttributesRepresentation returns the D2 representation of the edge's attributes.
// It merges attributes from both the base edge (D2Edge) and view edge (ViewEdge),
// with view edge attributes taking precedence when explicitly set.
// Returns an empty string if the edge has no attributes to output.
func getEdgeAttributesRepresentation(edge *d2view.Edge) string {
	if edge.D2Edge == nil {
		return ""
	}

	var builder strings.Builder
	base := edge.D2Edge
	view := edge.ViewEdge

	// Tooltip - view takes precedence
	tooltip := getMergedScalar(getEdgeTooltip(base), getEdgeTooltip(view))
	if tooltip != "" {
		builder.WriteString(fmt.Sprintf("        tooltip: \"%s\"\n", tooltip))
	}

	// Link - view takes precedence
	link := getMergedScalar(getEdgeLink(base), getEdgeLink(view))
	if link != "" {
		builder.WriteString(fmt.Sprintf("        link: %s\n", link))
	}

	// Classes - merge both sets
	classes := getMergedClasses((*ClassesEdge)(base), (*ClassesEdge)(view))
	for _, class := range classes {
		builder.WriteString(fmt.Sprintf("        class: %s\n", class))
	}

	// Style - merge base and view styles
	styleContent := getMergedStyleRepresentation((*withStyleEdge)(base), (*withStyleEdge)(view))
	if styleContent != "" {
		builder.WriteString("        style: {\n")
		builder.WriteString(styleContent)
		builder.WriteString("        }\n")
	}

	return builder.String()
}

// getEdgeTooltip safely extracts the tooltip value from an edge.
func getEdgeTooltip(edge *d2graph.Edge) string {
	if edge != nil && edge.Tooltip != nil && edge.Tooltip.Value != "" {
		return edge.Tooltip.Value
	}
	return ""
}

// getEdgeLink safely extracts the link value from an edge.
func getEdgeLink(edge *d2graph.Edge) string {
	if edge != nil && edge.Link != nil && edge.Link.Value != "" {
		return edge.Link.Value
	}
	return ""
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
		builder.WriteString(getLabelRepresentation(object.Label))
	}

	attrs := getObjectAttributesRepresentation(object)
	if attrs != "" {
		builder.WriteString(" {\n")
		builder.WriteString(attrs)
		builder.WriteString("    }")
	}
	builder.WriteString("\n")

	return builder.String()
}

// getLabelRepresentation returns the D2 representation of the object's label.
func getLabelRepresentation(label string) string {
	if label == "" {
		return ""
	}
	// needs quotes if string does not start and end with |
	needsQuotes := !(strings.HasPrefix(label, "|") && strings.HasSuffix(label, "|"))

	if needsQuotes {
		return fmt.Sprintf(": \"%s\"", label)
	}
	return fmt.Sprintf(": %s", label)
}

// getObjectAttributesRepresentation returns the D2 representation of the object's attributes.
// It merges attributes from both the base object and the view object, with view object
// attributes taking precedence when explicitly set.
// Returns an empty string if the object has no attributes to output.
func getObjectAttributesRepresentation(object *d2view.Object) string {
	base := object.BaseObject
	view := object.ViewObject

	// If neither object has attributes, return empty
	if base == nil && view == nil {
		return ""
	}

	var builder strings.Builder

	// Object-level attributes - view takes precedence over base shape
	if shape := getMergedScalar(getScalarIfExplicit(base, "shape"), getScalarIfExplicit(view, "shape")); shape != "" {
		builder.WriteString(fmt.Sprintf("        shape: %s\n", shape))
	}

	if icon := getMergedIcon(base, view); icon != "" {
		builder.WriteString(fmt.Sprintf("        icon: %s\n", icon))
	}

	if tooltip := getMergedScalar(getTooltip(base), getTooltip(view)); tooltip != "" {
		builder.WriteString(fmt.Sprintf("        tooltip: \"%s\"\n", tooltip))
	}

	if link := getMergedScalar(getLink(base), getLink(view)); link != "" {
		builder.WriteString(fmt.Sprintf("        link: %s\n", link))
	}

	if width := getMergedScalar(getWidth(base), getWidth(view)); width != "" {
		builder.WriteString(fmt.Sprintf("        width: %s\n", width))
	}

	if height := getMergedScalar(getHeight(base), getHeight(view)); height != "" {
		builder.WriteString(fmt.Sprintf("        height: %s\n", height))
	}

	if near := getMergedNear(base, view); near != "" {
		builder.WriteString(fmt.Sprintf("        near: %s\n", near))
	}

	if direction := getMergedScalar(getDirection(base), getDirection(view)); direction != "" {
		builder.WriteString(fmt.Sprintf("        direction: %s\n", direction))
	}

	// Classes - merge both sets
	classes := getMergedClasses((*ClassesObj)(base), (*ClassesObj)(view))
	for _, class := range classes {
		builder.WriteString(fmt.Sprintf("        class: %s\n", class))
	}

	// Style attributes - merge base and view styles
	styleContent := getMergedStyleRepresentation((*withStyleObj)(base), (*withStyleObj)(view))
	if styleContent != "" {
		builder.WriteString("        style: {\n")
		builder.WriteString(styleContent)
		builder.WriteString("        }\n")
	}

	return builder.String()
}

// getScalarIfExplicit returns the scalar value if it was explicitly set (has a MapKey).
func getScalarIfExplicit(obj *d2graph.Object, attrType string) string {
	if obj == nil {
		return ""
	}
	switch attrType {
	case "shape":
		if obj.Shape.Value != "" && obj.Shape.MapKey != nil {
			return obj.Shape.Value
		}
	case "direction":
		if obj.Direction.Value != "" && obj.Direction.MapKey != nil {
			return obj.Direction.Value
		}
	}
	return ""
}

// getMergedScalar returns the view value if set, otherwise the base value.
func getMergedScalar(baseVal, viewVal string) string {
	if viewVal != "" {
		return viewVal
	}
	return baseVal
}

func getTooltip(obj *d2graph.Object) string {
	if obj != nil && obj.Tooltip != nil && obj.Tooltip.Value != "" {
		return obj.Tooltip.Value
	}
	return ""
}

func getLink(obj *d2graph.Object) string {
	if obj != nil && obj.Link != nil && obj.Link.Value != "" {
		return obj.Link.Value
	}
	return ""
}

func getWidth(obj *d2graph.Object) string {
	if obj != nil && obj.WidthAttr != nil && obj.WidthAttr.Value != "" {
		return obj.WidthAttr.Value
	}
	return ""
}

func getHeight(obj *d2graph.Object) string {
	if obj != nil && obj.HeightAttr != nil && obj.HeightAttr.Value != "" {
		return obj.HeightAttr.Value
	}
	return ""
}

func getDirection(obj *d2graph.Object) string {
	if obj != nil && obj.Direction.Value != "" && obj.Direction.MapKey != nil {
		return obj.Direction.Value
	}
	return ""
}

// getMergedIcon returns the icon string, preferring the view object's icon if set, falling back to the base object's icon.
func getMergedIcon(base, view *d2graph.Object) string {
	if view != nil && view.Icon != nil {
		return view.Icon.String()
	}
	if base != nil && base.Icon != nil {
		return base.Icon.String()
	}
	return ""
}

func getMergedNear(base, view *d2graph.Object) string {
	// View takes precedence
	if view != nil && view.NearKey != nil {
		var nearParts []string
		for _, part := range view.NearKey.Path {
			nearParts = append(nearParts, part.Unbox().ScalarString())
		}
		return strings.Join(nearParts, ".")
	}
	if base != nil && base.NearKey != nil {
		var nearParts []string
		for _, part := range base.NearKey.Path {
			nearParts = append(nearParts, part.Unbox().ScalarString())
		}
		return strings.Join(nearParts, ".")
	}
	return ""
}

type withClasses interface {
	classes() []string
}

type ClassesObj d2graph.Object

func (obj *ClassesObj) classes() []string {
	if obj == nil {
		return []string{}
	}
	return obj.Classes
}

type ClassesEdge d2graph.Edge

func (obj *ClassesEdge) classes() []string {
	if obj == nil {
		return []string{}
	}
	return obj.Classes
}

func getMergedClasses(base, view withClasses) []string {
	classSet := make(map[string]struct{})
	var classes []string

	// Add base classes first
	if base != nil {
		for _, class := range base.classes() {
			if _, exists := classSet[class]; !exists {
				classSet[class] = struct{}{}
				classes = append(classes, class)
			}
		}
	}
	// Add view classes (may override or add)
	if view != nil {
		for _, class := range view.classes() {
			if _, exists := classSet[class]; !exists {
				classSet[class] = struct{}{}
				classes = append(classes, class)
			}
		}
	}
	return classes
}

type styleProvider interface {
	style() d2graph.Style
}

type withStyleObj d2graph.Object

func (obj *withStyleObj) style() d2graph.Style {
	if obj == nil {
		return d2graph.Style{}
	}
	return obj.Style
}

type withStyleEdge d2graph.Edge

func (obj *withStyleEdge) style() d2graph.Style {
	if obj == nil {
		return d2graph.Style{}
	}
	return obj.Style
}

// getMergedStyleRepresentation returns the D2 representation of merged style attributes.
// View style attributes take precedence over base style attributes.
func getMergedStyleRepresentation(base, view styleProvider) string {
	var builder strings.Builder

	// Helper to get style scalar with view precedence
	getStyleValue := func(baseStyle, viewStyle *d2graph.Scalar) string {
		if viewStyle != nil && viewStyle.Value != "" {
			return viewStyle.Value
		}
		if baseStyle != nil && baseStyle.Value != "" {
			return baseStyle.Value
		}
		return ""
	}

	var baseStyle, viewStyle d2graph.Style
	if base != nil {
		baseStyle = base.style()
	}
	if view != nil {
		viewStyle = view.style()
	}

	if val := getStyleValue(baseStyle.Opacity, viewStyle.Opacity); val != "" {
		builder.WriteString(fmt.Sprintf("            opacity: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Stroke, viewStyle.Stroke); val != "" {
		builder.WriteString(fmt.Sprintf("            stroke: \"%s\"\n", val))
	}
	if val := getStyleValue(baseStyle.Fill, viewStyle.Fill); val != "" {
		builder.WriteString(fmt.Sprintf("            fill: \"%s\"\n", val))
	}
	if val := getStyleValue(baseStyle.FillPattern, viewStyle.FillPattern); val != "" {
		builder.WriteString(fmt.Sprintf("            fill-pattern: %s\n", val))
	}
	if val := getStyleValue(baseStyle.StrokeWidth, viewStyle.StrokeWidth); val != "" {
		builder.WriteString(fmt.Sprintf("            stroke-width: %s\n", val))
	}
	if val := getStyleValue(baseStyle.StrokeDash, viewStyle.StrokeDash); val != "" {
		builder.WriteString(fmt.Sprintf("            stroke-dash: %s\n", val))
	}
	if val := getStyleValue(baseStyle.BorderRadius, viewStyle.BorderRadius); val != "" {
		builder.WriteString(fmt.Sprintf("            border-radius: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Shadow, viewStyle.Shadow); val != "" {
		builder.WriteString(fmt.Sprintf("            shadow: %s\n", val))
	}
	if val := getStyleValue(baseStyle.ThreeDee, viewStyle.ThreeDee); val != "" {
		builder.WriteString(fmt.Sprintf("            3d: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Multiple, viewStyle.Multiple); val != "" {
		builder.WriteString(fmt.Sprintf("            multiple: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Font, viewStyle.Font); val != "" {
		builder.WriteString(fmt.Sprintf("            font: %s\n", val))
	}
	if val := getStyleValue(baseStyle.FontSize, viewStyle.FontSize); val != "" {
		builder.WriteString(fmt.Sprintf("            font-size: %s\n", val))
	}
	if val := getStyleValue(baseStyle.FontColor, viewStyle.FontColor); val != "" {
		builder.WriteString(fmt.Sprintf("            font-color: \"%s\"\n", val))
	}
	if val := getStyleValue(baseStyle.Animated, viewStyle.Animated); val != "" {
		builder.WriteString(fmt.Sprintf("            animated: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Bold, viewStyle.Bold); val != "" {
		builder.WriteString(fmt.Sprintf("            bold: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Italic, viewStyle.Italic); val != "" {
		builder.WriteString(fmt.Sprintf("            italic: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Underline, viewStyle.Underline); val != "" {
		builder.WriteString(fmt.Sprintf("            underline: %s\n", val))
	}
	if val := getStyleValue(baseStyle.Filled, viewStyle.Filled); val != "" {
		builder.WriteString(fmt.Sprintf("            filled: %s\n", val))
	}
	if val := getStyleValue(baseStyle.DoubleBorder, viewStyle.DoubleBorder); val != "" {
		builder.WriteString(fmt.Sprintf("            double-border: %s\n", val))
	}
	if val := getStyleValue(baseStyle.TextTransform, viewStyle.TextTransform); val != "" {
		builder.WriteString(fmt.Sprintf("            text-transform: %s\n", val))
	}

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
