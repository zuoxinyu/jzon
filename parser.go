package jzon

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
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

// isWindows decides that the `\r\n` is indicating one or two new line(s)
var isWindows = runtime.GOOS == "windows"

func expect(c uint8, found uint8) error {
	return fmt.Errorf("expect '%c' but found '%c' at [%d:%d]", c, found, pos.row+1, pos.col+1)
}

func expectTypeOf(ex ValueType, found ValueType) error {
	return fmt.Errorf("expect node of type %s, but the real type is %s", typeStrings[ex], typeStrings[found])
}

func expectOneOf(pattern string, found byte) error {
	st := strings.Join(strings.Split(pattern, ""), "|")
	return fmt.Errorf("expect one of [%s] but found '%c' at [%d:%d]", st, found, pos.row+1, pos.col+1)
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
		case string(str[0:2]) == "\r\n" && isWindows:
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
	var k string
	var v *Jzon

	// return empty object directly
	var oldPos = pos
	trimmed := trimWhiteSpaces(json[1:])
	if trimmed[0] == '}' {
		pos.col++
		return obj, trimmed[1:], nil
	}
	// recover
	pos = oldPos
	rem = json

	for {
		pos.col++
		k, v, rem, err = parseKVPair(rem[1:])
		if err != nil {
			return
		}

		obj.obj[k] = v

		rem = trimWhiteSpaces(rem)
		if rem[0] == ',' {
			continue
		}

		break
	}

	rem = trimWhiteSpaces(rem)
	if rem[0] != '}' {
		err = expectOneOf("},", rem[0])
		return
	}

	pos.col++
	return obj, rem[1:], nil
}

func parseArr(json []byte) (arr *Jzon, rem []byte, err error) {
	arr = New(JzTypeArr)
	var v *Jzon

	// return empty array directly
	var oldPos = pos
	trimmed := trimWhiteSpaces(json[1:])
	if trimmed[0] == ']' {
		pos.col++
		return arr, trimmed[1:], nil
	}
	// recover
	pos = oldPos
	rem = json

	for {
		pos.col++
		v, rem, err = parse(rem[1:])
		if err != nil {
			return
		}

		arr.arr = append(arr.arr, v)

		rem = trimWhiteSpaces(rem)
		if rem[0] == ',' {
			continue
		}

		break
	}

	rem = trimWhiteSpaces(rem)
	if rem[0] != ']' {
		err = expectOneOf("],", rem[0])
		return
	}

	pos.col++
	return arr, rem[1:], nil
}

func parseStr(json []byte) (str *Jzon, rem []byte, err error) {
	str = New(JzTypeStr)
	var raw string

	raw, rem, err = parseKey(json)
	str.str = raw
	return
}

func parseNum(json []byte) (num *Jzon, rem []byte, err error) {
	num = New(JzTypeNum)
	var n int64
	var f float64
    type state uint64
	const (
	    _Start state = 2 << iota
	    _Zero                    // 0
	    _Dot                     // .
	    _Digit                   // 0-9
	    _NoneZero                // 1-9
	    _Exp                     // e E
	    _Plus                    // +
        _Minus                   // -
	    _End
    )

	var st = _Start
	var pst = &st
	var ex = []state{_Zero, _Digit}

	var match = func(ex ... state) bool {
	    for _, s := range ex {
	        if uint64(*pst) | uint64(s) > 0 {
	            return true
            }
        }
        return false
    }

    var isNoneZero = func(b byte) bool {
        return '0' <= b && b <= '9'
    }

    var isDigit = func(b byte) bool {
        return '1' <= b && b <= '9'
    }

	if json[0] == '-' {
		var neg = New(JzTypeNum)

		pos.col++
		neg, rem, err = parseNum(json[1:])
		neg.num = -neg.num
		return neg, rem, err
	}

	if json[0] == '0' && json[1] != '.' {
		err = expect('.', json[1])
		return
	}

	rem = json

	for {
		switch {
		case rem[0] == '0' && match(_Start):
		    ex = []state{_Dot}
		    st = _Zero

        case isNoneZero(rem[0]) && match(_Start):
            ex = []state{_Dot, _Digit}
            st = _NoneZero

        case rem[0] == '.' && match(_Zero, _Digit, _NoneZero):
            ex = []state{_Digit}
            st = _Dot

        case isDigit(rem[0]) && match(_Digit, _NoneZero):
            ex = []state{_Digit, _Exp, _Dot, _End}
            st = _Digit

        case isDigit(rem[0]) && match(_Dot):

        case rem[0] == 'e' && match(_Digit):

        case isDigit(rem[0]) && match(_Exp, _Plus, _Minus):


		}
	}


	_, err = fmt.Sscanf(string(json), "%d", &n)
	if err != nil {
		err = expectOneOf("0123456789.e+-", json[0])
		return
	}

	nParsed := len(strconv.FormatInt(n, 10))
	num.num = n
	pos.col += nParsed
	return num, json[nParsed:], nil
}

func parseTru(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:4]) == "true" {
		bol.bol = true
		pos.col += 4
		return bol, json[4:], nil
	}

	err = expectString("true", json[0:4])
	return
}

func parseFls(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:5]) == "false" {
		bol.bol = false
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
	json = trimWhiteSpaces(json)

	if json[0] == '"' {
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
		if err != nil {
			return
		}

		return
	}

	err = expect('"', json[0])
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

		uc = 0x10000 + (((uc & 0x3FF) << 10 | uc2) & 0x3FF)
	}

	var nBytes int
	switch {
	case uc < 0x80:    nBytes = 1
	case uc < 0x800:   nBytes = 2
	case uc < 0x10000: nBytes = 3
	default:           nBytes = 4
	}

	parsed = []byte{0, 0, 0, 0}

	// `c | 0x80` : set the 8th bit (0th from the lowest bit) to 1
	// `c & 0xBF` : reserve the 8th bit, and get the lower 6 bits
	// `c >> 0x6` : erase the lower 6 bits for the next one step
	// `c | firstByteMarkMap[n]`:set the highest n-1 bytes to 1

	switch nBytes {
	case 4: parsed[3] = byte((uc | 0x80) & 0xBF); uc >>= 6; fallthrough
	case 3: parsed[2] = byte((uc | 0x80) & 0xBF); uc >>= 6; fallthrough
	case 2: parsed[1] = byte((uc | 0x80) & 0xBF); uc >>= 6; fallthrough
	case 1: parsed[0] = byte( uc | firstByteMarkMap[nBytes])
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
		case '0' <= rem[i] && rem[i] <= '9': hc = uint32(0  + rem[i] - '0')
		case 'A' <= rem[i] && rem[i] <= 'F': hc = uint32(10 + rem[i] - 'A')
		case 'a' <= rem[i] && rem[i] <= 'f': hc = uint32(10 + rem[i] - 'a')
		default: return hex, nil, expectOneOf("0123456789ABCDEF", rem[i])
		}

		hex += hc * ex
		pos.col += 1
	}

	return hex, rem[4:], nil
}
