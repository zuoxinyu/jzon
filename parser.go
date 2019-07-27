package jzon

import (
	"fmt"
)

type position struct {
	row int
	col int
}

// pos is the global variable indicates the parsing step info
// NOTE: it makes the function `parse()` none re-entrant. to
// fix this, move it inner the function as a stack variable
var pos position

// SHORT_STRING_OPTIMIZED_CAP assumes that 16 was the most common
// length among short strings. NOTE: it need profiling to find a
// best capacity, or supply API let users modify it dynamically
const SHORT_STRING_OPTIMIZED_CAP = 16

// firstByteMarkMap is the first-byte-mark table in UTF-8 encoding
var firstByteMarkMap = [...]uint32{0x00, 0x00, 0xC0, 0xE0, 0xF0}

// escapeMap is for fast converting escape-able characters
var escapeMap = map[byte]byte{
	'"':  '"',
	'/':  '/',
	'\\': '\\',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

// nState indicates the inner state of the `parseNumeric` state machine
type nState uint64

const (
	_nStart    nState = 2 << iota
	_nZero            // 0
	_nDot             // .
	_nDigit0          // 0-9 after _nNoneZero
	_nDigit1          // 0-9 after _nDot
	_nDigit2          // 0-9 after _nExp or _nPlus or _Minus
	_nNoneZero        // 1-9
	_nExp             // e E
	_nPlus            // +
	_nMinus           // -
)

var nTypeStrings = map[nState]string{
	_nStart:    "_nStart",
	_nZero:     "_nZero",
	_nDot:      "_nDot",
	_nDigit0:   "_nDigit0",
	_nDigit1:   "_nDigit1",
	_nDigit2:   "_nDigit2",
	_nNoneZero: "_nNoneZero",
	_nExp:      "_nExp",
	_nPlus:     "_nPlus",
	_nMinus:    "_nMinus",
}

var nExStrings = map[nState]string{
	_nStart:    "0123456789-",
	_nZero:     ".eE",
	_nNoneZero: "0123456789.eE",
	_nDigit0:   "0123456789.eE",
	_nDigit1:   "0123456789eE",
	_nDigit2:   "0123456789",
	_nDot:      "0123456789",
	_nPlus:     "0123456789",
	_nMinus:    "0123456789",
	_nExp:      "0123456789+-",
}

func (st nState) match(ex ...nState) bool {
	for _, s := range ex {
		if uint64(st)&uint64(s) > 0 {
			return true
		}
	}
	return false
}

func isNoneZero(b byte) bool { return '1' <= b && b <= '9' }

func isDigit(b byte) bool { return '0' <= b && b <= '9' }

func isNumericChar(b byte) bool {
	return b == '-' || b == '+' || b == 'e' || b == 'E' || b == '.' || isDigit(b)
}

func expect(c uint8, found uint8) error {
	return fmt.Errorf("expect '%c' but found '%c' at [%d:%d]", c, found, pos.row+1, pos.col+1)
}

func expectNState(st nState, ex []nState) error {
	ss := make([]string, 4)
	for _, s := range ex {
		ss = append(ss, nTypeStrings[s])
	}
	return fmt.Errorf("expect state = %s, but the real is %s", ss, nTypeStrings[st])
}

func expectTypeOf(ex ValueType, found ValueType) error {
	return fmt.Errorf("expect node of type %s, but the real type is %s", typeStrings[ex], typeStrings[found])
}

func expectOneOf(pattern string, found byte) error {
	var cs = []rune{}
	for _, c := range pattern {
		cs = append(cs, c, '|')
	}
	return fmt.Errorf("expect one of [%s] but found '%c' at [%d:%d]", string(cs), found, pos.row+1, pos.col+1)
}

func expectString(pattern string, found []byte) error {
	return fmt.Errorf("expect \"%s\" but found \"%s\" at [%d:%d]", pattern, found, pos.row+1, pos.col+1)
}

func expectCodePoint() error {
	return fmt.Errorf("expect \"\\uXXXX\" formed string as valid Unicode codepoint at [%d:%d]", pos.row+1, pos.col+1)
}

func trimWhiteSpaces(str []byte) []byte {
	for {
		switch {
		case len(str) > 1 && str[0] == '\r' && str[1] == '\n':
			pos.row += 1
			pos.col = 0
			str = str[2:]
			continue

		case str[0] == ' ' || str[0] == '\t':
			pos.col += 1
			str = str[1:]
			continue

		case str[0] == '\n' || str[0] == '\r':
			pos.row += 1
			pos.col = 0
			str = str[1:]
			continue
		}

		break
	}

	return str
}

func parse(json []byte) (jz *Jzon, rem []byte, err error) {
	switch json[0] {
	case '{':
		return parseObj(json)
	case '[':
		return parseArr(json)
	case '"':
		return parseStr(json)
	case 't':
		return parseTru(json)
	case 'f':
		return parseFls(json)
	case 'n':
		return parseNul(json)
	case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
		return parseNum(json)
	case ' ', '\t', '\r', '\n':
		return parse(trimWhiteSpaces(json))
	default:
		return nil, json, expectOneOf("{[\"-1234567890ftn", json[0])
	}
}

func parseObj(json []byte) (obj *Jzon, rem []byte, err error) {
	obj = New(JzTypeObj)
	obj.data = make(map[string]*Jzon)
	var k string
	var v *Jzon
	var extraComma bool

	rem = json[1:]
	pos.col++

	for {
		switch rem[0] {
		case ',':
			if len(obj.data.(map[string]*Jzon)) == 0 {
				err = expectOneOf("}\"", rem[0])
				return
			}
			extraComma = true
			pos.col++
			rem = rem[1:]
			continue
		case ' ', '\t':
			pos.col++
			rem = rem[1:]
			continue
		case '\n', '\r':
			pos.row++
			rem = rem[1:]
			continue
		case '}':
			if extraComma {
				err = expectString("value", rem)
				return 
			}
			pos.col++
			rem = rem[1:]
			return
		default:
			extraComma = false
			k, v, rem, err = parseKVPair(rem)
			if err != nil {
				return
			}
			obj.data.(map[string]*Jzon)[k] = v
			continue
		}
	}
}

func parseArr(json []byte) (arr *Jzon, rem []byte, err error) {
	arr = New(JzTypeArr)
	arr.data = make([]*Jzon, 0)
	var v *Jzon
	var extraComma bool // TODO: add extension option 

	rem = json[1:]
	pos.col++

	for {
		switch rem[0] {
		case ',':
			if len(arr.data.([]*Jzon)) == 0 {
				err = expectOneOf("{[\"-1234567890ftn", rem[0])
				return
			}
			extraComma = true
			pos.col++
			rem = rem[1:]
			continue
		case ' ', '\t':
			pos.col++
			rem = rem[1:]
			continue
		case '\n', '\r':
			pos.row++
			rem = rem[1:]
			continue
		case ']':
			if extraComma {
				err = expectString("value", rem)
				return 
			}
			pos.col++
			rem = rem[1:]
			return
		default:
			extraComma = false
			v, rem, err = parse(rem)
			if err != nil {
				return
			}
			arr.data = append(arr.data.([]*Jzon), v)
			continue
		}
	}
}

func parseStr(json []byte) (str *Jzon, rem []byte, err error) {
	str = New(JzTypeStr)
	var raw string

	raw, rem, err = parseKey(json)
	str.data = raw
	return
}

func parseNum(json []byte) (num *Jzon, rem []byte, err error) {
	num = New(JzTypeInt)
	var n int64
	var f float64
	var isInt bool

	n, f, isInt, rem, err = parseNumeric(json)
	if err != nil {
		return
	}

	if isInt {
		num.Type = JzTypeInt
		num.data = n
	} else {
		num.Type = JzTypeFlt
		num.data = f
	}

	return
}

func parseTru(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:4]) == "true" {
		bol.data = true
		pos.col += 4
		return bol, json[4:], nil
	}

	err = expectString("true", json[0:4])
	return
}

func parseFls(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:5]) == "false" {
		bol.data = false
		pos.col += 5
		return bol, json[5:], nil
	}

	err = expectString("false", json[0:5])
	return
}

func parseNul(json []byte) (nul *Jzon, rem []byte, err error) {
	nul = New(JzTypeNul)
	if string(json[0:4]) == "null" {
		pos.col += 4
		return nul, json[4:], nil
	}

	err = expectString("null", json[0:4])
	return
}

func parseKVPair(json []byte) (k string, v *Jzon, rem []byte, err error) {
	k, rem, err = parseKey(json)
	if err != nil {
		return
	}

	rem = trimWhiteSpaces(rem)
	if rem[0] != ':' {
		err = expect(':', rem[0])
		return
	}

	pos.col++
	v, rem, err = parse(rem[1:])

	return
}

func parseKey(json []byte) (k string, rem []byte, err error) {
	var parsed = make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
	var c byte

	pos.col++
	rem = json[1:]

	for {
		switch {
		case rem[0] == '"':
			pos.col += 1
			rem = rem[1:]
			goto End

		case rem[0] != '\\' && rem[1] == '"':
			parsed = append(parsed, rem[0])
			pos.col += 2
			rem = rem[2:]
			goto End

		case rem[0] == '\\' && rem[1] == 'u':
			utf8str := make([]byte, 0, SHORT_STRING_OPTIMIZED_CAP)
			utf8str, rem, err = parseUnicode(rem)
			if err != nil {
				return
			}
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

		case rem[0] >= 0 && rem[0] < 32:
			err = expectOneOf("non-control", rem[0])
			return

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

func parseEscaped(json []byte) (escaped byte, rem []byte, err error) {
	rem = json
	escaped, ok := escapeMap[rem[1]]
	if !ok {
		err = expectOneOf("\"\\/bfnrtu", rem[1])
		return
	}

	pos.col += 2
	rem = rem[2:]
	return escaped, rem, nil
}

func parseUnicode(json []byte) (parsed []byte, rem []byte, err error) {
	// a valid UTF-8 code point is consisted of n bytes (1 < n < 5):
	// a leading byte begins with n * 1 and n-1 bytes begin with 10
	// 0000 0000 - 0000 007F | 0xxxxxxx
	// 0000 0080 - 0000 07FF | 110xxxxx 10xxxxxx
	// 0000 0800 - 0000 FFFF | 1110xxxx 10xxxxxx 10xxxxxx
	// 0001 0000 - 0010 FFFF | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx

	var uc, uc2 uint32
	var isInvalidCodePoint = func(cp uint32) bool {
		return 0xDC00 <= cp && cp <= 0xDFFF || cp == 0
	}

	rem = json[2:]
	pos.col += 2

	uc, rem, err = parseHex4(rem)
	if err != nil {
		return
	}

	if isInvalidCodePoint(uc) {
		err = expectCodePoint()
		return
	}

	if 0xD800 <= uc && uc <= 0xDBFF {
		if !(rem[0] == '\\' && rem[1] == 'u') {
			err = expectCodePoint()
			return
		}

		rem = rem[2:]
		pos.col += 2

		uc2, rem, err = parseHex4(rem)
		if err != nil {
			return
		}

		if isInvalidCodePoint(uc2) {
			err = expectCodePoint()
			return
		}

		uc = 0x10000 + (((uc&0x3FF)<<10 | uc2) & 0x3FF)
	}

	var nBytes int
	switch {
	case uc < 0x80:
		nBytes = 1
	case uc < 0x800:
		nBytes = 2
	case uc < 0x10000:
		nBytes = 3
	default:
		nBytes = 4
	}

	parsed = []byte{0, 0, 0, 0}

	// `c | 0x80` : set the 8th bit (0th from the lowest bit) to 1
	// `c & 0xBF` : reserve the 8th bit, and get the lower 6 bits
	// `c >> 0x6` : erase the lower 6 bits for the next one step
	// `c | firstByteMarkMap[n]`:set the highest n-1 bytes to 1

	switch nBytes {
	case 4:
		parsed[3] = byte((uc | 0x80) & 0xBF)
		uc >>= 6
		fallthrough
	case 3:
		parsed[2] = byte((uc | 0x80) & 0xBF)
		uc >>= 6
		fallthrough
	case 2:
		parsed[1] = byte((uc | 0x80) & 0xBF)
		uc >>= 6
		fallthrough
	case 1:
		parsed[0] = byte(uc | firstByteMarkMap[nBytes])
	}

	var realParsed []byte
	for _, c := range parsed {
		if c != 0 {
			realParsed = append(realParsed, c)
		}
	}

	return realParsed, rem, nil
}

func parseHex4(json []byte) (hex uint32, rem []byte, err error) {
	rem = json
	for i := uint32(0); i < 4; i++ {
		hc := uint32(0)
		ex := uint32(0x1000 >> (i * 4))
		switch {
		case '0' <= rem[i] && rem[i] <= '9':
			hc = uint32(0 + rem[i] - '0')
		case 'A' <= rem[i] && rem[i] <= 'F':
			hc = uint32(10 + rem[i] - 'A')
		case 'a' <= rem[i] && rem[i] <= 'f':
			hc = uint32(10 + rem[i] - 'a')
		default:
			return hex, nil, expectOneOf("0123456789ABCDEF", rem[i])
		}

		hex += hc * ex
		pos.col += 1
	}

	return hex, rem[4:], nil
}

func parseNumeric(json []byte) (n int64, f float64, isInt bool, rem []byte, err error) {
	var st = _nStart
	var ex = []nState{_nZero, _nNoneZero}
	var metExpPlus = false
	var nAfterDot = 1.0
	var nAfterExp = 0

	isInt = true
	rem = json

	// since the leading '-' should just occur less than once
	if rem[0] == '-' {
		if len(rem) > 1 && !('0' <= rem[1] && rem[1] <= '9') {
			err = expectOneOf("0123456789", rem[1])
			return
		}
		pos.col++
		n, f, isInt, rem, err = parseNumeric(rem[1:])
		n = -n
		return
	}

	for {
		switch {
		case len(rem) == 0: // Must be the first condition, avoiding illegal memory access
			if isInt {
				f = 0
			} else {
				n = 0
			}
			return

		case rem[0] == '0' && st.match(_nStart):
			ex = []nState{_nDot, _nExp}
			st = _nZero

		case rem[0] == '.' && st.match(_nZero, _nDigit0, _nNoneZero):
			ex = []nState{_nDigit1}
			st = _nDot

			isInt = false

		case isDigit(rem[0]) && st.match(_nDot, _nDigit1):
			ex = []nState{_nDigit1, _nExp}
			st = _nDigit1

			nAfterDot *= 10
			f += float64(float64(rem[0]-'0') / nAfterDot)

		case isNoneZero(rem[0]) && st.match(_nStart):
			ex = []nState{_nDot, _nDigit0, _nExp}
			st = _nNoneZero

			n = int64(rem[0] - '0')
			f = float64(rem[0] - '0')

		case isDigit(rem[0]) && st.match(_nDigit0, _nNoneZero):
			ex = []nState{_nDigit0, _nExp, _nDot}
			st = _nDigit0

			n = n*10 + int64(rem[0]-'0')
			f = f*10 + float64(rem[0]-'0')

		case (rem[0] == 'e' || rem[0] == 'E') && st.match(_nZero, _nNoneZero, _nDigit0, _nDigit1):
			ex = []nState{_nPlus, _nMinus, _nDigit2}
			st = _nExp

			isInt = false
			metExpPlus = true

		case rem[0] == '+' && st.match(_nExp):
			ex = []nState{_nDigit2}
			st = _nPlus

			metExpPlus = true

		case rem[0] == '-' && st.match(_nExp):
			ex = []nState{_nDigit2}
			st = _nMinus

			metExpPlus = false

		case isDigit(rem[0]) && st.match(_nExp, _nPlus, _nMinus, _nDigit2):
			ex = []nState{_nDigit2}
			st = _nDigit2

			nAfterExp = nAfterExp*10 + int(rem[0]-'0')
			if metExpPlus {
				for i := 0; i < nAfterExp; i++ {
					f *= 10
				}
			} else {
				for i := 0; i < nAfterExp; i++ {
					f /= 10
				}
			}

		case !isNumericChar(rem[0]) && st.match(_nZero, _nNoneZero, _nDigit0, _nDigit1, _nDigit2):
			if isInt {
				f = 0
			} else {
				n = 0
			}
			return

		default:
			if st.match(ex...) {
				err = expectOneOf(nExStrings[st], rem[0])
			} else {
				err = expectNState(st, ex)
			}

			return
		}
		rem = rem[1:]
		pos.col++

	}
}
