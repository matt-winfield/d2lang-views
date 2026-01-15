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
	checkErr(err, "Unable to read source file")

	err = ensureDirExists(args.Destination)
	checkErr(err, "Unable to create destination directory")

	color.Green("Done!" + string(content))
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
