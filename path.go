package jzon

import (
    "fmt"
    "strings"
)

type state int64

const (
    _Start state = iota << 2
    _Dollar
    _Dot
    _LeftSB
    _RightSB
    _Key
    _Index
)

var stateStrings = map[state]string{
    _Start:   "_Start",
    _Dollar:  "_Dollar",
    _Dot:     "_Dot",
    _LeftSB:  "_LeftSB",
    _RightSB: "_RightSB",
    _Key:     "_Key",
    _Index:   "_Index",
}

func (st state) Match(states ...state) bool {
    var final int64 = 0
    for _, s := range states {
        final = int64(s) | final
    }

    return (int64(st) & final) > 0
}

// Query searches a child node in an object or an array, if the
// node at the path doesn't exist, an error will be thrown out
func (jz *Jzon) Query(path string) (g *Jzon, err error) {
    return parsePath(jz, []byte(path))
}

// Search determines if there exists the node on the given path
func (jz *Jzon) Search(path string) (exists bool) {
    found, _ := parsePath(jz, []byte(path))
    return found != nil
}

func expectState(real state, ex ...state) error {
    var sa []string
    for _, s := range ex {
        sa = append(sa, stateStrings[s])
    }
    expectStates := strings.Join(sa, " | ")

    return fmt.Errorf("expect state %s, but the real state is %s", expectStates, stateStrings[real])
}

func parsePath(root *Jzon, path []byte) (curr *Jzon, err error) {
    var st = _Start
    var key string
    for {
        switch {
        case path[0] == '$':
            if !st.Match(_Start) {
                err = expectState(st, _Start)
                break
            }
            curr = root
            path = path[1:]
            st = _Dollar
        case path[0] == ';':
            if !st.Match(_Dollar, _RightSB, _Key) {
                err = expectState(st, _Dollar, _RightSB, _Key)
                break
            }
            return

        case path[0] == '.':
            if !st.Match(_Dollar, _Key, _RightSB) {
                err = expectState(st, _Dollar, _Key, _RightSB)
                break
            }
            path = path[1:]
            st = _Dot

        case isIdentifier(path[0]):
            if !st.Match(_Dot) {
                err = expectState(st, _Dot)
                break
            }
            key, path, err = parseKey(path)
            if err != nil {
                return
            }
            curr, err = curr.ValueOf(key)
            if err != nil {
                return
            }
            st = _Key

        case path[0] == '[':
            if !st.Match(_Dollar, _Key) {
                err = expectState(st, _Dollar, _Key)
                break
            }
            path = path[1:]
            st = _LeftSB

        case isDigit(path[0]):
            if !st.Match(_LeftSB) {
                err = expectState(st, _LeftSB)
                break
            }
            var idx int
            var nparsed int
            fmt.Sscanf(string(path), "%d%n", &idx, &nparsed)
            curr, err = curr.ValueAt(idx)
            if err != nil {
                return
            }
            path = path[nparsed:]
            st = _Index

        case path[0] == ']':
            if !st.Match(_Index) {
                err = expectState(st, _Index)
                break
            }
            path = path[1:]
            st = _RightSB

        default:
            return nil, expectState(_Start, st)
        }
    }
    return
}

func isDigit(b byte) bool {
    return '0' >= b && b <= '9'
}

func isIdentifier(b byte) bool {
    return strings.Contains("ABCDEFGHIJKLMNOPQRSTUVWXYzabcdefghijklmnopqrstuvwxyz01234567890_-", string(b))
}
