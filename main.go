package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/fatih/color"
	"github.com/matt-winfield/d2lang-views/compile"
	"github.com/matt-winfield/d2lang-views/render"
)

var args struct {
	Source      string `arg:"positional"`
	Destination string `arg:"positional"`
	Layout      string `arg:"-l,--layout" help:"layout engine to use (e.g., dagre, elk)"`
}

func main() {
	arg.MustParse(&args)

	if args.Source == "" || args.Destination == "" {
		color.Red("ERR: Missing required arguments")
		color.Yellow("Usage: program <source> <destination>")
		os.Exit(1)
	}

	content, err := os.ReadFile(args.Source)
	checkErr(err, "Unable to read source file")

	reader := bytes.NewReader(content)
	graph, _, err := compile.CompileD2(args.Source, reader)
	checkErr(err, "Unable to compile D2 content")

	rootObjectIds := compile.ExtractRootObjectIds(graph)

	sourceReader := bytes.NewReader(content)
	viewContent, err := replaceViewLayers(sourceReader, graph, rootObjectIds)
	checkErr(err, "Unable to replace view content")

	// Output D2 file to same directory as source (to preserve relative imports)
	viewOutputPath := getD2OutputPath(args.Source)

	// Ensure output directory exists
	outputDir := getDirectory(viewOutputPath)
	if outputDir != "" {
		err = ensureDirExists(outputDir)
		checkErr(err, "Unable to create output directory")
	}

	err = os.WriteFile(viewOutputPath, []byte(viewContent), 0644)
	checkErr(err, "Unable to write output file with views")

	color.Green("Successfully wrote D2 output to %s", viewOutputPath)

	// Get all layer names for rendering (not just views)
	var layerNames []string
	for _, layer := range graph.Layers {
		layerNames = append(layerNames, layer.Name)
	}

	// SVG output goes to the destination specified by user
	svgOutputPath := args.Destination
	svgOutputDir := stripExtension(svgOutputPath)

	// Ensure SVG output directory exists
	svgParentDir := getDirectory(svgOutputPath)
	if svgParentDir != "" {
		err = ensureDirExists(svgParentDir)
		checkErr(err, "Unable to create SVG output directory")
	}

	// Create subfolder for layer SVGs (same name as output file without extension)
	err = ensureDirExists(svgOutputDir)
	checkErr(err, "Unable to create SVG layers directory")

	// Compile each layer to SVG using d2 CLI
	err = render.RenderD2File(viewOutputPath, svgOutputPath, args.Layout)
	checkErr(err, "Unable to compile layers to SVG")

	for _, layerName := range layerNames {
		color.Green("Successfully compiled layer '%s' to %s/%s.svg", layerName, svgOutputDir, layerName)
	}

	fmt.Printf("Compiled %d layers to SVG\n", len(layerNames))
}

// ensureDirExists checks if a directory exists at the given path,
// and creates it (including any necessary parents) if it does not.
// Returns an error if a file (not a directory) exists at the path.
func ensureDirExists(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}
	return nil
}

// checkErr is a helper function that checks for an error and exits the program if one is found.
func checkErr(err error, msg string) {
	if err != nil {
		color.Red("ERR: %s: %v", msg, err)
		os.Exit(1)
	}
}

// getD2OutputPath constructs the output file path for the compiled D2 output.
// It places the output file in the same directory as the source file (to preserve relative imports).
// The output is always named <source_basename>-compiled.d2 to avoid overwriting the source.
func getD2OutputPath(sourcePath string) string {
	sourceDir := getDirectory(sourcePath)
	sourceBasename := getBasename(sourcePath)

	// Create compiled filename by inserting "-compiled" before extension
	compiledBasename := insertSuffixBeforeExtension(sourceBasename, "-compiled")

	if sourceDir == "" {
		return compiledBasename
	}
	return sourceDir + "/" + compiledBasename
}

// getDirectory extracts the directory portion of a path.
// Returns empty string if path has no directory component.
func getDirectory(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return ""
}

// getBasename extracts the filename from a path (including extension).
func getBasename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// insertSuffixBeforeExtension inserts a suffix before the file extension.
// If no extension exists, the suffix is appended at the end.
func insertSuffixBeforeExtension(path, suffix string) string {
	basename := getBasename(path)
	dir := getDirectory(path)

	// Find the last dot in the basename
	dotIdx := -1
	for i := len(basename) - 1; i >= 0; i-- {
		if basename[i] == '.' {
			dotIdx = i
			break
		}
	}

	var newBasename string
	if dotIdx == -1 {
		newBasename = basename + suffix
	} else {
		newBasename = basename[:dotIdx] + suffix + basename[dotIdx:]
	}

	if dir == "" {
		return newBasename
	}
	return dir + "/" + newBasename
}

// stripExtension removes the file extension from a path.
func stripExtension(path string) string {
	basename := getBasename(path)
	dir := getDirectory(path)

	// Find the last dot in the basename
	dotIdx := -1
	for i := len(basename) - 1; i >= 0; i-- {
		if basename[i] == '.' {
			dotIdx = i
			break
		}
	}

	var newBasename string
	if dotIdx == -1 {
		newBasename = basename
	} else {
		newBasename = basename[:dotIdx]
	}

	if dir == "" {
		return newBasename
	}
	return dir + "/" + newBasename
}
