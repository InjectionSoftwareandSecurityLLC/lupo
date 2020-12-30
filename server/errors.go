package server

import (
	"github.com/fatih/color"
)

// Define custom colors for text output
var errorPrint = color.New(color.FgRed).Add(color.Bold)

// ErrorHandler - wrapper function to hand off server related errors to insure servers Handler functions maintain correct interface mappings
func ErrorHandler(err error) error {
	errorPrint.Printf("\nerror: ")
	errorPrint.DisableColor()
	errorPrint.Printf("%s\n", err)
	errorPrint.EnableColor()
	return err
}
