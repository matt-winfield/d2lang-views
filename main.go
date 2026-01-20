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

	err = ensureDirExists(args.Destination)
	checkErr(err, "Unable to create destination directory")

	reader := bytes.NewReader(content)
	graph, _, err := compile.CompileD2(args.Source, reader)
	checkErr(err, "Unable to compile D2 content")

	views := compile.GetViewsNodes(graph)
	fmt.Printf("Found %d view nodes\n", len(views))

	rootObjectIds := compile.ExtractRootObjectIds(graph)

	sourceReader := bytes.NewReader(content)
	viewContent, err := replaceViewLayers(sourceReader, graph, rootObjectIds)
	checkErr(err, "Unable to replace view content")

	viewOutputPath := getOutputFilePath(args.Source, args.Destination, "-views.d2")
	err = os.WriteFile(viewOutputPath, []byte(viewContent), 0644)
	checkErr(err, "Unable to write output file with views")

	color.Green("Successfully wrote D2 output to %s", viewOutputPath)

	// Get all layer names for rendering (not just views)
	var layerNames []string
	for _, layer := range graph.Layers {
		layerNames = append(layerNames, layer.Name)
	}

	// Create subfolder for SVG output (same name as output d2 file without extension)
	svgOutputDir := getOutputFilePath(args.Source, args.Destination, "-views")
	err = ensureDirExists(svgOutputDir)
	checkErr(err, "Unable to create SVG output directory")

	// Compile each layer to SVG using d2 CLI
	err = render.RenderD2File(viewOutputPath, svgOutputDir+".svg", args.Layout)
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

// getOutputFilePath constructs the output file path for the AST JSON
// based on the source file name and destination directory.
// It appends extension to the source file name after stripping the directory path and file extension.
func getOutputFilePath(sourcePath, destinationDir string, extension string) string {
	baseName := sourcePath
	if idx := len(sourcePath) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if sourcePath[i] == '/' || sourcePath[i] == '\\' {
				baseName = sourcePath[i+1:]
				break
			}
		}
	}

	if dotIdx := len(baseName) - 1; dotIdx >= 0 {
		for i := dotIdx; i >= 0; i-- {
			if baseName[i] == '.' {
				baseName = baseName[:i]
				break
			}
		}
	}

	return destinationDir + "/" + baseName + extension
}
