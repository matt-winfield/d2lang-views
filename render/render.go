package render

import (
	"errors"
	"fmt"
	"os/exec"
)

// RenderD2File compiles a D2 file to SVG using the d2 CLI.
// inputPath is the path to the D2 source file.
// outputPath is the path where the SVG will be written.
// layout is the optional layout engine to use (e.g., "dagre", "elk"). Pass empty string for default.
func RenderD2File(inputPath, outputPath, layout string) error {
	args := buildD2Args(inputPath, outputPath, layout)
	return runD2(args)
}

// RenderD2Layer compiles a specific layer from a D2 file to SVG using the d2 CLI.
// inputPath is the path to the D2 source file.
// outputPath is the path where the SVG will be written.
// layout is the optional layout engine to use (e.g., "dagre", "elk"). Pass empty string for default.
// target is the name of the layer to render.
func RenderD2Layer(inputPath, outputPath, layout, target string) error {
	args := buildD2ArgsWithTarget(inputPath, outputPath, layout, target)
	return runD2(args)
}

// buildD2Args constructs the command line arguments for the d2 CLI.
func buildD2Args(inputPath, outputPath, layout string) []string {
	args := []string{}
	if layout != "" {
		args = append(args, "--layout", layout)
	}
	args = append(args, inputPath, outputPath)
	return args
}

// buildD2ArgsWithTarget constructs the command line arguments for the d2 CLI with a target layer.
func buildD2ArgsWithTarget(inputPath, outputPath, layout, target string) []string {
	args := []string{}
	if layout != "" {
		args = append(args, "--layout", layout)
	}
	args = append(args, "--target", target)
	args = append(args, inputPath, outputPath)
	return args
}

// runD2 executes the d2 CLI with the given arguments.
func runD2(args []string) error {
	cmd := exec.Command("d2", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("d2 compilation failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// RenderOptions contains configuration for rendering D2 files to SVG.
type RenderOptions struct {
	InputPath      string   // Path to the compiled D2 file
	OutputPath     string   // Path for the main SVG output
	Layout         string   // Layout engine (e.g., "dagre", "elk"). Empty for default.
	ViewsOnly      bool     // If true, only render view layers
	ViewLayerNames []string // Names of view layers (required if ViewsOnly is true)
}

// Validate checks if the RenderOptions are valid.
func (o *RenderOptions) Validate() error {
	if o.ViewsOnly && len(o.ViewLayerNames) == 0 {
		return errors.New("no view layers found (layers marked with #view)")
	}
	return nil
}

// RenderResult contains information about what was rendered.
type RenderResult struct {
	RenderedLayers []string // Names of layers that were rendered
	ViewsOnly      bool     // Whether only views were rendered
}

// Render renders a D2 file to SVG based on the provided options.
// If ViewsOnly is true, it renders each view layer separately using --target.
// Otherwise, it renders the entire file including all layers.
func Render(opts RenderOptions) (*RenderResult, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if opts.ViewsOnly {
		return renderViewsOnly(opts)
	}
	return renderAll(opts)
}

// renderViewsOnly renders only the view layers using --target option.
func renderViewsOnly(opts RenderOptions) (*RenderResult, error) {
	outputDir := stripExtension(opts.OutputPath)

	for _, viewName := range opts.ViewLayerNames {
		viewSvgPath := buildOutputPath(outputDir, viewName)
		if err := RenderD2Layer(opts.InputPath, viewSvgPath, opts.Layout, viewName); err != nil {
			return nil, fmt.Errorf("unable to compile view '%s' to SVG: %w", viewName, err)
		}
	}

	return &RenderResult{
		RenderedLayers: opts.ViewLayerNames,
		ViewsOnly:      true,
	}, nil
}

// renderAll renders the entire D2 file including all layers.
func renderAll(opts RenderOptions) (*RenderResult, error) {
	if err := RenderD2File(opts.InputPath, opts.OutputPath, opts.Layout); err != nil {
		return nil, err
	}

	return &RenderResult{
		RenderedLayers: nil, // All layers rendered via d2 CLI
		ViewsOnly:      false,
	}, nil
}

// buildOutputPath constructs the output path for a layer SVG.
func buildOutputPath(outputDir, layerName string) string {
	return outputDir + "/" + layerName + ".svg"
}

// stripExtension removes the file extension from a path.
func stripExtension(path string) string {
	// Find the last slash to get basename start
	basenameStart := 0
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			basenameStart = i + 1
			break
		}
	}

	// Find the last dot in the basename
	dotIdx := -1
	for i := len(path) - 1; i >= basenameStart; i-- {
		if path[i] == '.' {
			dotIdx = i
			break
		}
	}

	if dotIdx == -1 {
		return path
	}
	return path[:dotIdx]
}
