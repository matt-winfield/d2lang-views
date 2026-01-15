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
}
