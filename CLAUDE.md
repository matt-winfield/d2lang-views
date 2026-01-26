# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

d2lang-views is a Go tool that extends D2 (https://d2lang.com/) diagram language to support sub-views. It allows creating focused visualizations of specific parts of larger diagrams by referencing entities from the base diagram within view layers.

The tool scans D2 files for layers marked with a `#view` comment and generates new D2 content where referenced entities from the base diagram are automatically included in the view, with their labels and relationships preserved.

## Commands

### Build

```bash
go build
```

### Install (with automatic version detection)

```bash
# Install from a specific version tag - version is automatically embedded
go install github.com/matt-winfield/d2lang-views@v1.3.0

# Or build locally with manual version override
go build -ldflags "-X github.com/matt-winfield/d2lang-views/version.version=v1.3.0"
```

### Run

```bash
# Run without building
go run . ./path/to/diagram.d2 output/diagram.svg

# Or after building
./d2lang-views ./path/to/diagram.d2 output/diagram.svg
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestName ./...

# Run tests with detailed comparison output (useful for debugging test failures)
go test -v ./... | less
```

## Architecture

The codebase is organized into three main logical components:

### 1. Parser (parser.go)

Handles D2 compilation and extraction of view-related information:

-   `compileD2()` - Uses the D2 compiler to parse D2 source into AST and graph
-   `getViewsNodes()` - Finds all layers marked with `#view` comment
-   `extractRootObjectIds()` - Recursively extracts all entity IDs from the base layer (non-layer content)
-   `getLayersNode()` - Locates the layers node in the AST

The parser uses the D2 library (oss.terrastruct.com/d2) which provides:

-   `d2ast` - Abstract Syntax Tree representation
-   `d2graph` - Compiled graph with objects and edges
-   `d2compiler` - Compilation from source to graph

### 2. Processor (d2view/processor.go)

Processes view layers and constructs View objects with filtered objects and edges:

-   `ProcessViews()` - Main entry point that processes all view layers
-   `getExplicitObjectIds()` - Identifies explicitly referenced objects (vs implicit parents)
-   `processViewObjects()` - Filters implicit parents and builds Object instances with filtered IDs
-   `processViewEdges()` - Extracts relevant edges and maps IDs to filtered hierarchy
-   `buildFilteredAbsoluteId()` - Constructs IDs after removing implicit parent segments

**Implicit Parent Filtering**: When `parent.child` is referenced without `parent`, the parent is filtered out and `child` becomes a root-level entity. Only objects explicitly listed in the view are included.

### 3. Exporter (exporter.go)

Generates view content by replacing view layer references with full entity definitions:

-   `replaceViewLayers()` - Main orchestrator that processes all views
-   `generateViewContent()` - For each view, generates D2 representation of referenced entities
-   `getObjectD2Representation()` - Converts graph objects to D2 syntax (with labels)
-   `getEdgeD2Representation()` - Converts graph edges to D2 syntax
-   `applyRangeOperations()` - Handles text manipulation to insert generated content and remove references

The range operation system allows merging overlapping replacements and insertions while preserving source positions.

### 4. Version (version/version.go)

Provides version information for generated output:

-   `RepoURL` - Constant containing the GitHub repository URL
-   `Version()` - Returns the version, auto-detected from build info or ldflags
-   `GeneratedFileHeader()` - Returns the comment for the start of generated D2 files
-   `GeneratedViewHeader()` - Returns the comment for the start of each generated view

The version is automatically detected using `runtime/debug.ReadBuildInfo()`:
1. Module version when installed via `go install @version`
2. VCS revision when building from a clean git repository
3. Ldflags override: `-X github.com/matt-winfield/d2lang-views/version.version=v1.0.0`
4. Falls back to "dev" during development

The generated D2 files include comments indicating they were auto-generated, including the tool version. This helps users understand the file should not be manually edited.

### 5. Main (main.go)

Entry point and CLI handling:

-   Uses `github.com/alexflint/go-arg` for argument parsing
-   Requires source D2 file path and SVG output path
-   Outputs:
    -   `<source-dir>/<filename>-compiled.d2` - Compiled D2 file in the same directory as the source (preserves relative import paths)
    -   `<svg-output-path>` - SVG files in the specified output location
-   Uses `github.com/fatih/color` for colored terminal output

**Output behavior:**

The tool generates two types of output:

1. **D2 file**: Always placed in the same directory as the source file with `-compiled` suffix (e.g., `source.d2` → `source-compiled.d2`). This preserves relative import paths that may be used in the diagram.

2. **SVG files**: Placed in the location specified by the destination argument. The D2 CLI is invoked to render the compiled D2 file into SVG format, creating a folder structure with an `index.svg` and individual layer SVG files.

## Key Concepts

### Root Objects

"Root objects" are entities defined in the base diagram (outside of layers). The tool tracks their full hierarchical IDs (e.g., "first.second.third") because D2 allows nested entities.

### View References

When a view layer contains a line like `entityName`, it's a reference to an entity from the base diagram. The tool replaces these references with the full entity definition including its label.

### Absolute IDs

Objects are identified by their full path from root (e.g., "parent.child.grandchild"). The `getAbsoluteId()` function traverses the parent chain to build this path.

### Explicit vs Implicit References

An object is "explicit" if it appears as the final element of a reference path. When you write `parent.child`:
- `child` is explicit (it's the target of the reference)
- `parent` is implicit (created only as a container)

The processor detects this by comparing each object's reference path length to its depth. Implicit objects are filtered out, and their children's IDs are adjusted accordingly.

### Range Operations

The AST provides byte ranges for every element. The exporter uses these ranges to:

-   Find where to insert generated content (at the earliest reference)
-   Remove reference lines that were replaced
-   Merge overlapping operations to avoid double-processing

## Testing

Tests use table-driven patterns with inline D2 content as strings. The `go-cmp` library is used for detailed difference reporting in test failures.

Test files:

-   `compile/parser_test.go` - Tests view detection and entity extraction
-   `d2view/processor_test.go` - Tests view processing and implicit parent filtering
-   `exporter_test.go` - Tests view content generation and replacement
-   `main_test.go` - Integration tests
-   `version/version_test.go` - Tests version header generation

When tests fail, `go-cmp` output shows the exact differences between expected and actual D2 output, making it easy to identify issues.

## Development

You must always use Test-Driven Development (TDD) when adding features, making changes, or fixing bugs. Write tests first to define the expected behaviour, run them to verify they fail, then implement the code to make them pass.
