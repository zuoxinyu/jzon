package jzon

import (
	"fmt"
	"os"
	"strings"
)

// *nix TTY colors
const (
	BLACK     = "\033[0;30m"
	RED       = "\033[0;31m"
	GREEN     = "\033[0;32m"
	YELLOW    = "\033[0;33m"
	BLUE      = "\033[0;34m"
	PURPLE    = "\033[0;35m"
	SKY       = "\033[0;36m"
	WHITE     = "\033[0;37m"
	RESET     = "\033[0m"
	HIGHLIGHT = "\033[1m"
	UNDERLINE = "\033[4m"
	BLINK     = "\033[5m"
	REVERSE   = "\033[7m"
	FADEOUT   = "\033[8m"
)

// Format converts raw compact JSON to human-reading text
func Format(compact string) (formatted string, err error) {
	jz, err := Parse([]byte(compact))
	if err != nil {
		return
	}

	return jz.Format(0, 2), nil
}

// Compact converts formatted JSON to compact text
func Compact(formatted string) (compact string, err error) {
	jz, err := Parse([]byte(formatted))
	if err != nil {
		return
	}

	return jz.Compact(), nil
}

// Format generates human-readable text
func (jz *Jzon) Format(indent int, step int) string {
	return jz.render(indent, step, false, false)
}

// Compact generates compact text
func (jz *Jzon) Compact() string {
	switch jz.Type {
	case JzTypeArr:
		var ss []string
		as, _ := jz.AMap(func(v *Jzon) Any { return v.Compact() })
		for _, a := range as {
			ss = append(ss, a.(string))
		}
		return "[" + strings.Join(ss, ",") + "]"

	case JzTypeObj:
		var ss []string
		as, _ := jz.OMap(func(k string, v *Jzon) Any { return "\"" + k + "\"" + ":" + v.Compact() })
		for _, a := range as {
			ss = append(ss, a.(string))
		}
		return "{" + strings.Join(ss, ",") + "}"

	case JzTypeStr: // FIXME: escaped characters
		s, _ := jz.String()
		var buf []byte
		for _, ch := range []byte(s) {
			switch ch {
			case '\n':
				buf = append(buf, '\\', 'n')
			case '\b':
				buf = append(buf, '\\', 'b')
			case '\f':
				buf = append(buf, '\\', 'f')
			case '\t':
				buf = append(buf, '\\', 't')
			case '\r':
				buf = append(buf, '\\', 'r')
			case '\\':
				buf = append(buf, '\\', '\\')
			case '"':
				buf = append(buf, '\\', '"')
			case '/':
				buf = append(buf, '\\', '/')
			default: 
				buf = append(buf, byte(ch))
			}
		}
		return "\"" + string(buf) + "\""

	case JzTypeInt:
		n, _ := jz.Integer()
		return fmt.Sprintf("%d", n)

	case JzTypeFlt:
		f, _ := jz.Float()
		return fmt.Sprintf("%f", f)

	case JzTypeBol:
		b, _ := jz.Bool()
		if b {
			return "true"
		}
		return "false"

	case JzTypeNul:
		return "null"
	}
	return ""
}

// GoString implements the `GoString` interface
func (jz *Jzon) GoString() string {
	return jz.Compact();
}

// Formatter implements the `Formmater` interface
func (jz *Jzon) Formmater(f fmt.State, c rune) {

}

// Print prints human-reading JSON text to writer
func (jz *Jzon) Print() {
	fmt.Print(jz.Compact())
}

// Coloring prints colored and formatted JSON text on the terminal. if it's not
// a terminal or doesn't support colors, it just prints raw but formatted text
func (jz *Jzon) Coloring(file *os.File) {
	if file != os.Stdout {
		fmt.Fprint(file, jz.Format(0, 2))
	}

	fmt.Fprint(file, jz.render(0, 2, false, true))
}

func (jz *Jzon) render(indent int, step int, useTab bool, useColor bool) string {
	colorify := func(c string, s string) string {
		if useColor {
			return c + s + RESET
		}

		return s
	}

	indentf := func(i, s int) string {
		if useTab {
			return strings.Repeat("\t", i/s)
		}

		return strings.Repeat(" ", i)
	}

	switch jz.Type {
	case JzTypeArr:
		var ss []string
		vs, _ := jz.AMap(func(v *Jzon) Any { return v.render(indent+step, step, useTab, useColor) })
		for _, v := range vs {
			ss = append(ss, v.(string))
		}
		return "[" + strings.Join(ss, ", ") + "]"

	case JzTypeObj:
		if l, _ := jz.Length(); l == 0 {
			return "{}"
		}

		vs, _ := jz.OMap(func(k string, v *Jzon) Any {
			return indentf(indent+step, step) + YELLOW + "\"" + k + "\": " + RESET + v.render(indent+step, step, useTab, useColor)
		})
		var ss []string
		for _, v := range vs {
			ss = append(ss, v.(string))
		}
		return "{\n" + strings.Join(ss, ",\n") + "\n" + indentf(indent, step) + "}"

	case JzTypeNul:
		return colorify(RED, jz.Compact())

	case JzTypeBol:
		return colorify(GREEN, jz.Compact())

	case JzTypeFlt, JzTypeInt:
		return colorify(BLUE, jz.Compact())

	case JzTypeStr:
		return colorify(PURPLE, jz.Compact())
	}

	return ""
}
