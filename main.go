package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/fatih/color"
)

var args struct {
	Source      string `arg:"positional"`
	Destination string `arg:"positional"`
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
	ast, err := parseD2(args.Source, reader)

	checkErr(err, "Unable to parse D2 content")

	json, err := json.MarshalIndent(ast, "", "  ")
	checkErr(err, "Unable to marshal AST to JSON")

	outputPath := astOutputPath(args.Source, args.Destination)
	err = os.WriteFile(outputPath, json, 0644)
	checkErr(err, "Unable to write output file")

	layers := getViewsNodes(ast)

	fmt.Printf("Found %d view nodes\n", len(layers))

	color.Green("Successfully wrote output to %s", outputPath)
}

// ensureDirExists checks if a directory exists at the given path,
// and creates it (including any necessary parents) if it does not.
func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
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

// astOutputPath constructs the output file path for the AST JSON
// based on the source file name and destination directory.
// It appends "-ast.json" to the source file name after stripping the directory path and file extension.
func astOutputPath(sourcePath, destinationDir string) string {
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

	return destinationDir + "/" + baseName + "-ast.json"
}
