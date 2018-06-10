package jzon

import (
	"io"
	"os"
)

// Format converts raw compact JSON to human-reading text
func Format(compact string, indent int, useTab bool) (formatted string, err error) {
	jz, err := Parse([]byte(compact))
	if err != nil {
	    return
	}

	return jz.Format(indent, useTab), nil
}

// Compact converts formatted JSON to compact text
func Compact(formatted string) (compact string, err error) {
	jz, err := Parse([]byte(formatted))
	if err != nil {
	    return
	}

	return jz.Compact(), nil
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

// Coloring prints colored and formatted JSON text on the terminal. if it's not
// a terminal or doesn't support colors, it just prints raw but formatted text
func (jz *Jzon) Coloring(file os.File) {
	// TODO:
}
