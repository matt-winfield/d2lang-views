package main

import (
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

	if err != nil {
		color.Red("ERR: Unable to read source file: %v", err)
		os.Exit(1)
	}

	err = ensureDirExists(args.Destination)

	if err != nil {
		color.Red("ERR: Unable to create destination directory: %v", err)
		os.Exit(1)
	}

	color.Green("Done!" + string(content))
}

func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}
