package render

import (
	"fmt"
	"os/exec"
)

// RenderD2File compiles a D2 file to SVG using the d2 CLI.
// inputPath is the path to the D2 source file.
// outputPath is the path where the SVG will be written.
// layout is the optional layout engine to use (e.g., "dagre", "elk"). Pass empty string for default.
func RenderD2File(inputPath, outputPath, layout string) error {
	args := []string{}
	if layout != "" {
		args = append(args, "--layout", layout)
	}
	args = append(args, inputPath, outputPath)

	cmd := exec.Command("d2", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("d2 compilation failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
