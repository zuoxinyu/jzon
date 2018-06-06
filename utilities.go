package jzon

import (
	"io"
	"os"
)

// Format converts raw compact JSON to human-reading
func Format(compact string, indent int, useTab bool) string {
	// TODO:

	return ""
}

// Compact converts indented JSON to compact text
func Compact(indented string) string {
	// TODO:

	return ""
}

// Print prints human-reading JSON text to reader
func (jz *Jzon) Print(reader io.Reader) {
	// TODO:
}

// Colored prints colored JSON text on terminal
// if it's not a terminal or doesn't support colors,
// it just print raw but formatted text
func (jz *Jzon) Colored(file os.File) {
	// TODO:
}
