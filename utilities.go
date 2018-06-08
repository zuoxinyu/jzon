package jzon

import (
	"io"
	"os"
)

// Format converts raw compact JSON to human-reading text
func Format(compact string, indent int, useTab bool) string {
	jz, err := Parse([]byte(compact))
	if err != nil {
		panic(err)
	}

	return jz.Format(indent, useTab)
}

// Compact converts formatted JSON to compact text
func Compact(formatted string) string {
	jz, err := Parse([]byte(formatted))
	if err != nil {
		panic(err)
	}

	return jz.Compact()
}

// Format generates human-reading text
func (jz *Jzon) Format(indent int, useTab bool) string {
	// TODO:

	return ""
}

// Compact generates compact text
func (jz *Jzon) Compact() string {
	// TODO:
	return ""
}

// Print prints human-reading JSON text to reader
func (jz *Jzon) Print(reader io.Reader) {
	// TODO:
}

// Colored prints colored and formatted JSON text on terminal.
// if it's not a terminal or doesn't support colors,
// it just prints raw but formatted text
func (jz *Jzon) Colored(file os.File) {
	// TODO:
}
