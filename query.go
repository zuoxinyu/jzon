package jzon

import (
	"fmt"
	"strconv"
	"strings"
)

type state int64

const (
	_Start state = 2 << iota
	_Dollar
	_Dot
	_LeftSB
	_RightSB
	_Key
	_Index
	_Comma
)

var stateStrings = map[state]string{
	_Start:   "_Start",
	_Dollar:  "_Dollar",
	_Dot:     "_Dot",
	_LeftSB:  "_LeftSB",
	_RightSB: "_RightSB",
	_Key:     "_Key",
	_Index:   "_Index",
	_Comma:   "_Comma",
}

func (st state) match(states ...state) bool {
	var final int64 = 0
	for _, s := range states {
		final = int64(s) | final
	}

	return (int64(st) & final) > 0
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
	return parsePath(jz, []byte(path))
}

// Search determines whether there exists the node on the given path
func (jz *Jzon) Search(path string) (exists bool) {
	found, _ := parsePath(jz, []byte(path))
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
	var isDigit = func(b byte) bool {
		return '0' >= b && b <= '9'
	}
	var isIdentifier = func(b byte) bool {
		return strings.Contains("ABCDEFGHIJKLMNOPQRSTUVWXYzabcdefghijklmnopqrstuvwxyz01234567890_-", string(b))
	}
	var st = _Start
	var ex = []state{_Dollar}
	var key string

	// a typical state machine model
	for {
		switch {
		case path[0] == '$' && st.match(_Start):
			curr = root
			path = path[1:]
			st = _Dollar
			ex = []state{_Dot, _LeftSB, _Comma}

		case path[0] == ';' && st.match(_Dollar, _RightSB, _Key):
			st = _Comma
			ex = []state{}
			return

		case path[0] == '.' && st.match(_Dollar, _Key, _RightSB):
			path = path[1:]
			st = _Dot
			ex = []state{_Key}

		case isIdentifier(path[0]) && st.match(_Dot):
			key, path, err = parsePathKey(path)
			if err != nil {
				return
			}
			curr, err = curr.ValueOf(key)
			if err != nil {
				return
			}
			st = _Key
			ex = []state{_Dot, _LeftSB, _Comma}

		case path[0] == '[' && st.match(_Dollar, _Key):
			path = path[1:]
			st = _LeftSB
			ex = []state{_Index}

		case isDigit(path[0]) && st.match(_LeftSB):
			var idx int
			_, err = fmt.Sscanf(string(path), "%d", &idx)
			if err != nil {
				return
			}
			nBytes := len(strconv.Itoa(idx))
			curr, err = curr.ValueAt(idx)
			if err != nil {
				return
			}
			path = path[nBytes:]
			st = _Index
			ex = []state{_RightSB}

		case path[0] == ']' && st.match(_Index):
			path = path[1:]
			st = _RightSB
			ex = []state{_Dot, _Comma}

		default:
			return nil, expectState(st, ex)
		}
	}
	return
}

func parsePathKey(json []byte) (k string, rem []byte, err error) {
	var parsed = make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
	var c byte

	rem = json

	for {
		switch {
		case rem[0] == '\\' && rem[1] == 'u':
			utf8str := make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
			utf8str, rem, err = parseUnicode(rem)
			for _, c := range utf8str {
				parsed = append(parsed, c)
			}
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
