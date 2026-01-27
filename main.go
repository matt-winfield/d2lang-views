package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/fatih/color"
	"github.com/matt-winfield/d2lang-views/compile"
	"github.com/matt-winfield/d2lang-views/render"
	"github.com/matt-winfield/d2lang-views/version"
	"github.com/matt-winfield/d2lang-views/watch"
	"oss.terrastruct.com/d2/d2graph"
)

type argsType struct {
	Source      string `arg:"positional" help:"path to source D2 file with views"`
	Destination string `arg:"positional" help:"path to output SVG file"`
	Layout      string `arg:"-l,--layout" help:"layout engine to use (e.g., dagre, elk)"`
	Debug       bool   `arg:"-d,--debug" help:"enable debug output - output intermediate AST files"`
	ViewsOnly   bool   `arg:"--views-only" help:"only output SVGs for view layers (marked with #view)"`
	Watch       bool   `arg:"-w,--watch" help:"watch source file and imports for changes, automatically recompile"`
}

// Version returns the version string for the --version flag
func (argsType) Version() string {
	return "d2lang-views " + version.Version()
}

var args argsType

func main() {
	arg.MustParse(&args)

	if args.Source == "" || args.Destination == "" {
		color.Red("ERR: Missing required arguments")
		color.Yellow("Usage: program <source> <destination>")
		os.Exit(1)
	}

	// Check for output directory conflict before any file operations
	err := checkOutputConflict(args.Source, args.Destination)
	checkErr(err, "Output path conflict detected")

	// Run initial compilation
	if err := runCompilation(); err != nil {
		color.Red("ERR: %v", err)
		os.Exit(1)
	}

	// If watch mode is enabled, start watching for changes
	if args.Watch {
		runWatchMode()
	}
}

// runCompilation performs a single compilation of the D2 source file.
// Returns an error if compilation fails, allowing watch mode to continue.
func runCompilation() error {
	content, err := os.ReadFile(args.Source)
	if err != nil {
		return fmt.Errorf("unable to read source file: %w", err)
	}

	reader := bytes.NewReader(content)
	graph, _, err := compile.CompileD2(args.Source, reader)
	if err != nil {
		return fmt.Errorf("unable to compile D2 content: %w", err)
	}

	rootObjectIds := compile.ExtractRootObjectIds(graph)

	sourceReader := bytes.NewReader(content)
	viewContent, err := replaceViewLayers(sourceReader, graph, rootObjectIds)
	if err != nil {
		return fmt.Errorf("unable to replace view content: %w", err)
	}

	// Output D2 file to same directory as source (to preserve relative imports)
	viewOutputPath := getD2OutputPath(args.Source)

	// Ensure output directory exists
	outputDir := getDirectory(viewOutputPath)
	if outputDir != "" {
		if err := ensureDirExists(outputDir); err != nil {
			return fmt.Errorf("unable to create output directory: %w", err)
		}
	}

	if args.Debug {
		debugOutputPath := viewOutputPath + ".debug.json"
		jsonContent, err := json.MarshalIndent(graph, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to marshall graph as JSON: %w", err)
		}
		if err := os.WriteFile(debugOutputPath, jsonContent, 0644); err != nil {
			return fmt.Errorf("unable to write debug output file: %w", err)
		}
		color.Green("Wrote debug D2 output to %s", debugOutputPath)
	}

	if err := os.WriteFile(viewOutputPath, []byte(viewContent), 0644); err != nil {
		return fmt.Errorf("unable to write output file with views: %w", err)
	}

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
		if err := ensureDirExists(svgParentDir); err != nil {
			return fmt.Errorf("unable to create SVG output directory: %w", err)
		}
	}

	// Create subfolder for layer SVGs (same name as output file without extension)
	if err := ensureDirExists(svgOutputDir); err != nil {
		return fmt.Errorf("unable to create SVG layers directory: %w", err)
	}

	// Compile to SVG using d2 CLI
	renderOpts := render.RenderOptions{
		InputPath:      viewOutputPath,
		OutputPath:     svgOutputPath,
		Layout:         args.Layout,
		ViewsOnly:      args.ViewsOnly,
		ViewLayerNames: getViewLayerNames(graph),
	}

	result, err := render.Render(renderOpts)
	if err != nil {
		return fmt.Errorf("unable to compile to SVG: %w", err)
	}

	if result.ViewsOnly {
		for _, viewName := range result.RenderedLayers {
			color.Green("Successfully compiled view '%s' to %s/%s.svg", viewName, svgOutputDir, viewName)
		}
		fmt.Printf("Compiled %d view layers to SVG\n", len(result.RenderedLayers))
	} else {
		for _, layerName := range layerNames {
			color.Green("Successfully compiled layer '%s' to %s/%s.svg", layerName, svgOutputDir, layerName)
		}
		fmt.Printf("Compiled %d layers to SVG\n", len(layerNames))
	}

	return nil
}

// runWatchMode starts watching the source file and its imports for changes.
func runWatchMode() {
	color.Cyan("\nWatching for changes... (Ctrl+C to stop)")

	watcher, err := watch.NewWatcher(watch.WatcherOptions{
		SourcePath:    args.Source,
		DebounceDelay: 200 * time.Millisecond,
		OnChange: func() error {
			fmt.Println() // Add newline before recompilation output
			color.Cyan("Change detected, recompiling...")
			if err := runCompilation(); err != nil {
				color.Red("Compilation error: %v", err)
				// Return nil to continue watching even after errors
				return nil
			}
			color.Cyan("Watching for changes... (Ctrl+C to stop)")
			return nil
		},
	})
	if err != nil {
		color.Red("ERR: Failed to create watcher: %v", err)
		os.Exit(1)
	}

	if err := watcher.Start(); err != nil {
		color.Red("ERR: Failed to start watcher: %v", err)
		os.Exit(1)
	}

	// Print watched files
	watchedFiles := watcher.GetWatchedFiles()
	if len(watchedFiles) > 1 {
		color.Yellow("Watching %d files (source + %d imports)", len(watchedFiles), len(watchedFiles)-1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println() // Add newline after ^C
	color.Yellow("Stopping watcher...")
	defer func() {
		if err := watcher.Stop(); err != nil {
			color.Red("ERR: Failed to stop watcher: %v", err)
			os.Exit(1)
		}
	}()
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

// checkOutputConflict verifies that the SVG output directory won't conflict with the source directory.
// This prevents the d2 CLI from deleting source files when it creates the layer output folder.
// The d2 CLI creates a folder with the same name as the output file (minus extension) for layer SVGs.
func checkOutputConflict(sourcePath, destination string) error {
	// Get the directory containing the source file
	sourceDir := getDirectory(sourcePath)
	if sourceDir == "" {
		sourceDir = "."
	}

	// Get the absolute path of the source directory
	absSourceDir, err := absolutePath(sourceDir)
	if err != nil {
		return fmt.Errorf("unable to resolve source directory: %w", err)
	}

	// The d2 CLI creates a folder for layer SVGs based on the output filename
	// e.g., output.svg creates output/ folder
	svgOutputDir := stripExtension(destination)
	absSvgOutputDir, err := absolutePath(svgOutputDir)
	if err != nil {
		// If we can't resolve the path, it might not exist yet - that's okay
		// But we should still check if it WOULD conflict
		absSvgOutputDir = svgOutputDir
	}

	// Check if the SVG output directory is the same as or a parent of the source directory
	// This would cause d2 to overwrite/delete source files
	if absSourceDir == absSvgOutputDir || isSubpath(absSourceDir, absSvgOutputDir) {
		return fmt.Errorf(
			"SVG output directory '%s' conflicts with source directory '%s'. "+
				"The d2 CLI would overwrite your source files. "+
				"Please use a different output path (e.g., '%s-output.svg')",
			svgOutputDir, sourceDir, stripExtension(destination))
	}

	return nil
}

// absolutePath returns the absolute path, resolving any relative components.
func absolutePath(path string) (string, error) {
	// Use filepath.Abs for proper path resolution
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	// Clean the path to remove any . or .. components
	return filepath.Clean(abs), nil
}

// isSubpath checks if child is a subdirectory of parent.
func isSubpath(child, parent string) bool {
	// Ensure paths are clean and use consistent separators
	child = filepath.Clean(child)
	parent = filepath.Clean(parent)

	// Add trailing separator to parent to ensure we match directory boundaries
	if !strings.HasSuffix(parent, string(filepath.Separator)) {
		parent = parent + string(filepath.Separator)
	}

	return strings.HasPrefix(child, parent)
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

// getViewLayerNames returns the names of all layers marked with #view.
func getViewLayerNames(graph *d2graph.Graph) []string {
	views := compile.GetViewsNodes(graph)
	names := make([]string, 0, len(views))
	for _, view := range views {
		names = append(names, view.Name)
	}
	return names
}
