package server

import "github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"

// ErrorHandler - wrapper function to hand off server related errors to insure servers Handler functions maintain correct interface mappings
func ErrorHandler(err error) error {
	core.ErrorColorBold.Printf("\nerror: ")
	core.ErrorColorBold.DisableColor()
	core.ErrorColorBold.Printf("%s\n", err)
	core.ErrorColorBold.EnableColor()
	return err
}
