package jzon

import (
	"os"
	"fmt"
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

// Print prints human-reading JSON text to writer
func (jz *Jzon) Print() {
	switch jz.Type {
	case JzTypeArr:
		fmt.Printf("[")
		jz.AMap(func(v *Jzon) []Any {
			v.Print()
			fmt.Printf(",")
			return nil
		})
        if l, _ := jz.Length(); l > 0 {
            fmt.Printf("\b]")
        } else {
            fmt.Printf("]")
        }

    case JzTypeObj:
        fmt.Printf("{")
        jz.OMap(func(k string, v *Jzon) Any {
            fmt.Printf("\"%s\":", k)
            v.Print()
            fmt.Printf(",")
            return nil
        })
        if l, _ := jz.Length(); l > 0 {
            fmt.Printf("\b}")
        } else {
            fmt.Printf("}")
        }

    case JzTypeStr:
        s, _ := jz.String()
        fmt.Printf("\"%s\"", s)

    case JzTypeNum:
        n, _ := jz.Number()
        fmt.Printf("%d", n)

    case JzTypeBol:
        b, _ := jz.Bool()
        fmt.Printf("%v", b)

    case JzTypeNul:
        fmt.Printf("null")
	}
}

// Coloring prints colored and formatted JSON text on the terminal. if it's not
// a terminal or doesn't support colors, it just prints raw but formatted text
func (jz *Jzon) Coloring(file os.File) {
	// TODO:
}
