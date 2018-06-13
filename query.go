package jzon

import (
	"fmt"
	"strings"
)

// state indicates the inner state of the `parsePath` state machine
type state int64

const (
	// $.key1[1].big-array[1:4]
	_Start     state = 2 << iota
	_Dollar          // $        root
	_Dot             // .        key mark
	_LeftSB          // [        index mark
	_RightSB         // ]        index end
	_Key             // .*       object key
	_Index           // [1-9]\d+ array index
	_Colon           // :        slice mark
	_Semicolon       // ;        line tail
)

var stateStrings = map[state]string{
	_Start:     "_Start",
	_Dollar:    "_Dollar",
	_Dot:       "_Dot",
	_LeftSB:    "_LeftSB",
	_RightSB:   "_RightSB",
	_Key:       "_Key",
	_Index:     "_Index",
	_Colon:     "_Colon",
	_Semicolon: "_Semicolon",
}

func (st state) match(states ...state) bool {
	for _, s := range states {
		if uint64(st)&uint64(s) > 0 {
			return true
		}
	}
	return false
}

// Query searches a child node in an object or an array, if the
// node at the path doesn't exist, an error will be thrown out
func (jz *Jzon) Query(path string) (g *Jzon, err error) {
	// in the implements of function `parsePath()` we don't handle
	// exceptions about slice bounds out of range. here we simply
	// throw the error recovered from those unhandled exceptions
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("maybe out of bound: %v", e)
		}
	}()
	return parsePath(jz, append([]byte(path), ';'))
}

// Search determines whether there exists the node on the given path
func (jz *Jzon) Search(path string) (exists bool) {
	found, _ := jz.Query(path)
	return found != nil
}

func expectState(real state, ex []state) error {
	var sa []string
	for _, s := range ex {
		sa = append(sa, stateStrings[s])
	}
	expectStates := strings.Join(sa, " | ")

	return fmt.Errorf("expect state %s, but the real state is %s", expectStates, stateStrings[real])
}

func parsePath(root *Jzon, path []byte) (curr *Jzon, err error) {
	var st = _Start
	var ex = []state{_Dollar}
	var key string

	// a typical state machine model
	for {
		switch {
		case path[0] == '$' && st.match(_Start):
			ex = []state{_Dot, _LeftSB, _Semicolon}
            st = _Dollar

            curr = root
            path = path[1:]

		case path[0] == ';' && st.match(_Dollar, _RightSB, _Key):
			ex = []state{}
            st = _Semicolon

			return

		case path[0] == '.' && st.match(_Dollar, _Key, _RightSB):
			ex = []state{_Key}
            st = _Dot

            path = path[1:]

		case path[0] == '[' && st.match(_Dollar, _Key):
			ex = []state{_Index}
            st = _LeftSB

            path = path[1:]

		case isDigit(path[0]) && st.match(_LeftSB):
            ex = []state{_RightSB}
            st = _Index

            var n int64
            var f float64
            var isInt bool
            n, f, isInt, path, err = parseNumeric(path)
            if err != nil {
                return
            }

            if !isInt {
                err = fmt.Errorf("expect an integer index, but found float: %v", f)
                return
            }

			curr, err = curr.ValueAt(int(n))
			if err != nil {
				return
			}

		case path[0] == ']' && st.match(_Index):
            ex = []state{_Dot, _Semicolon}
            st = _RightSB

			path = path[1:]

		case st.match(_Dot):
            ex = []state{_Dot, _LeftSB, _Semicolon}
            st = _Key

			key, path, err = parsePathKey(path)
			if err != nil {
				return
			}
			curr, err = curr.ValueOf(key)
			if err != nil {
				return
			}

		default:
			return nil, expectState(st, ex)
		}
	}
}

// parsePathKey parses as `parseKey()`, except that the given string
// isn't surrounded with ", and it will escape some more characters
func parsePathKey(path []byte) (k string, rem []byte, err error) {
	var parsed = make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
	var c byte

	rem = path

	for {
		switch {
		case rem[0] == '\\' && rem[1] == 'u':
			utf8str := make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
			utf8str, rem, err = parseUnicode(rem)
			for _, c := range utf8str {
				parsed = append(parsed, c)
			}
			continue

		case rem[0] == '\\' && rem[1] == '.':
			parsed = append(parsed, '.')
			rem = rem[2:]
			continue

		case rem[0] == '\\' && rem[1] == '[':
			parsed = append(parsed, '[')
			rem = rem[2:]
			continue

		case rem[0] == '\\' && rem[1] == ']':
			parsed = append(parsed, ']')
			rem = rem[2:]
			continue

		case rem[0] == '\\' && rem[1] == ';':
			parsed = append(parsed, ';')
			rem = rem[2:]
			continue

		case rem[0] == '\\' && rem[1] != 'u':
			c, rem, err = parseEscaped(rem)
			if err != nil {
				return
			}
			parsed = append(parsed, c)
			continue

		case rem[0] == '.' || rem[0] == '[' || rem[0] == ';':
			goto End

		default:
			parsed = append(parsed, rem[0])
			pos.col += 1
			rem = rem[1:]
			continue
		}
	}
End:
	return string(parsed), rem, nil
}
