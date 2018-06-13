package jzon

import (
	"fmt"
	"os"
	"strings"
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
	return ""
}

// Compact generates compact text
func (jz *Jzon) Compact() string {
	switch jz.Type {
	case JzTypeArr:
	    var ss []string
		as, _ := jz.AMap(func(v *Jzon) Any { return v.Compact() })
		for _, a := range as { ss = append(ss, a.(string))}
		return "[" + strings.Join(ss, ",") + "]"

	case JzTypeObj:
        var ss []string
		as, _ := jz.OMap(func(k string, v *Jzon) Any { return "\"" + k + "\"" + ":" + v.Compact() })
        for _, a := range as { ss = append(ss, a.(string))}
        return "{" + strings.Join(ss, ",") + "}"

	case JzTypeStr: // maybe should de-escape characters?
		s, _ := jz.String()
		return "\"" + s + "\""

	case JzTypeInt:
		n, _ := jz.Integer()
		return fmt.Sprintf("%d", n)

	case JzTypeFlt:
		f, _ := jz.Float()
		return fmt.Sprintf("%f", f)

	case JzTypeBol:
		b, _ := jz.Bool()
		if b { return "true" } else { return "false" }

	case JzTypeNul:
		return "null"
	}
	return ""
}

// Print prints human-reading JSON text to writer
func (jz *Jzon) Print() {
    fmt.Print(jz.Compact())
}

// Coloring prints colored and formatted JSON text on the terminal. if it's not
// a terminal or doesn't support colors, it just prints raw but formatted text
func (jz *Jzon) Coloring(file os.File) {
	// TODO:
}
