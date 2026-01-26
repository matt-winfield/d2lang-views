package render

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		layerName  string
		expected   string
	}{
		{
			name:      "simple path",
			outputDir: "output",
			layerName: "frontend",
			expected:  "output/frontend.svg",
		},
		{
			name:      "nested output dir",
			outputDir: "path/to/output",
			layerName: "backend",
			expected:  "path/to/output/backend.svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOutputPath(tt.outputDir, tt.layerName)
			if result != tt.expected {
				t.Errorf("buildOutputPath(%q, %q) = %q, want %q", tt.outputDir, tt.layerName, result, tt.expected)
			}
		})
	}
}

func TestRenderResult(t *testing.T) {
	result := RenderResult{
		RenderedLayers: []string{"frontend", "backend"},
		ViewsOnly:      true,
	}

	if len(result.RenderedLayers) != 2 {
		t.Errorf("expected 2 rendered layers, got %d", len(result.RenderedLayers))
	}
	if !result.ViewsOnly {
		t.Error("expected ViewsOnly to be true")
	}
}

func TestRenderOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        RenderOptions
		expectError bool
	}{
		{
			name: "valid options - full render",
			opts: RenderOptions{
				InputPath:  "input.d2",
				OutputPath: "output.svg",
				ViewsOnly:  false,
			},
			expectError: false,
		},
		{
			name: "valid options - views only with layers",
			opts: RenderOptions{
				InputPath:       "input.d2",
				OutputPath:      "output.svg",
				ViewsOnly:       true,
				ViewLayerNames:  []string{"frontend"},
			},
			expectError: false,
		},
		{
			name: "views only with no view layers",
			opts: RenderOptions{
				InputPath:       "input.d2",
				OutputPath:      "output.svg",
				ViewsOnly:       true,
				ViewLayerNames:  []string{},
			},
			expectError: true,
		},
		{
			name: "views only with nil view layers",
			opts: RenderOptions{
				InputPath:       "input.d2",
				OutputPath:      "output.svg",
				ViewsOnly:       true,
				ViewLayerNames:  nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestStripExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"output.svg", "output"},
		{"path/to/output.svg", "path/to/output"},
		{"noextension", "noextension"},
		{"file.name.svg", "file.name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripExtension(tt.input)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("stripExtension(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func TestBuildD2Args_NoOptions(t *testing.T) {
	args := buildD2Args("input.d2", "output.svg", "")
	expected := []string{"input.d2", "output.svg"}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestBuildD2Args_WithLayout(t *testing.T) {
	args := buildD2Args("input.d2", "output.svg", "elk")
	expected := []string{"--layout", "elk", "input.d2", "output.svg"}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestBuildD2ArgsWithTarget_NoOptions(t *testing.T) {
	args := buildD2ArgsWithTarget("input.d2", "output.svg", "", "myview")
	expected := []string{"--target", "myview", "input.d2", "output.svg"}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestBuildD2ArgsWithTarget_WithLayout(t *testing.T) {
	args := buildD2ArgsWithTarget("input.d2", "output.svg", "dagre", "myview")
	expected := []string{"--layout", "dagre", "--target", "myview", "input.d2", "output.svg"}

	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %v", len(expected), len(args), args)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}
