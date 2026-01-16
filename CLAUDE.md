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

### Run

```bash
# Run without building
go run . ./path/to/diagram.d2 output/directory

# Or after building
./d2lang-views ./path/to/diagram.d2 output/directory
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

### 2. Exporter (exporter.go)

Generates view content by replacing view layer references with full entity definitions:

-   `replaceViewLayers()` - Main orchestrator that processes all views
-   `generateViewContent()` - For each view, generates D2 representation of referenced entities
-   `getObjectD2Representation()` - Converts graph objects to D2 syntax (with labels)
-   `getEdgeD2Representation()` - Converts graph edges to D2 syntax
-   `applyRangeOperations()` - Handles text manipulation to insert generated content and remove references

**Key algorithm**: For each view, the exporter:

1. Identifies which objects in the view exist in the base layer (rootObjectIds)
2. For referenced objects, generates their D2 representation including labels from the base graph
3. Finds relationships between view objects in the base graph and includes them
4. Inserts the generated content at the earliest reference position
5. Removes the original reference lines that were replaced

The range operation system allows merging overlapping replacements and insertions while preserving source positions.

### 3. Main (main.go)

Entry point and CLI handling:

-   Uses `github.com/alexflint/go-arg` for argument parsing
-   Requires source D2 file path and output directory
-   Outputs:
    -   `<filename>-graph.json` - JSON representation of the compiled graph
    -   `<filename>-with-views.d2` - Modified D2 file with expanded views
-   Uses `github.com/fatih/color` for colored terminal output

## Key Concepts

### Root Objects

"Root objects" are entities defined in the base diagram (outside of layers). The tool tracks their full hierarchical IDs (e.g., "first.second.third") because D2 allows nested entities.

### View References

When a view layer contains a line like `entityName`, it's a reference to an entity from the base diagram. The tool replaces these references with the full entity definition including its label.

### Absolute IDs

Objects are identified by their full path from root (e.g., "parent.child.grandchild"). The `getAbsoluteId()` function traverses the parent chain to build this path.

### Range Operations

The AST provides byte ranges for every element. The exporter uses these ranges to:

-   Find where to insert generated content (at the earliest reference)
-   Remove reference lines that were replaced
-   Merge overlapping operations to avoid double-processing

## Testing

Tests use table-driven patterns with inline D2 content as strings. The `go-cmp` library is used for detailed difference reporting in test failures.

Test files:

-   `parser_test.go` - Tests view detection and entity extraction
-   `exporter_test.go` - Tests view content generation and replacement
-   `main_test.go` - Integration tests

When tests fail, `go-cmp` output shows the exact differences between expected and actual D2 output, making it easy to identify issues.

## Development

You must always use Test-Driven Development (TDD) when adding features, making changes, or fixing bugs. Write tests first to define the expected behaviour, run them to verify they fail, then implement the code to make them pass.
