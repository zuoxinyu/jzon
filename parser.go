package jzon

import (
	"fmt"
	"strconv"
	"strings"
)

type position struct {
	row int
	col int
}

var pos position

func expect(c uint8, found uint8) error {
	return fmt.Errorf("expect '%c' but found '%c' at [%d:%d]", c, found, pos.row, pos.col)
}

func expectTypeOf(ex ValueType, real ValueType) error {
	return fmt.Errorf("expect node of type %s, but the real type is %s", typeStrings[ex], typeStrings[real])
}

func expectOneOf(pattern string, found byte) error {
	st := strings.Join(strings.Split(pattern, ""), "|")
	return fmt.Errorf("expect one of [%s] but found '%c' at [%d:%d]", st, found, pos.row, pos.col)
}

func expectString(pattern string, found []byte) error {
	return fmt.Errorf("expect \"%s\" but found \"%s\" at [%d:%d]", pattern, found, pos.row, pos.col)
}

func skipWhiteSpaces(str []byte) []byte {
	for {
		switch str[0] {
		case ' ', '\t':
			pos.col += 1
			str = str[1:]
			continue
		case '\n', '\r':
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
		return parse(skipWhiteSpaces(json))
	}

	err = expectOneOf("{[\"-1234567890ftn", json[0])
	return
}

func parseObj(json []byte) (obj *Jzon, rem []byte, err error) {
	obj = New(JzTypeObj)
	var k string
	var v *Jzon

	// return empty object directly
	oldPos := pos
	try := skipWhiteSpaces(json[1:])
	if try[0] == '}' {
		pos.col++
		return obj, try[1:], nil
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

		rem = skipWhiteSpaces(rem)
		if rem[0] == ',' {
			continue
		}

		break
	}

	rem = skipWhiteSpaces(rem)
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
	oldPos := pos
	try := skipWhiteSpaces(json[1:])
	if try[0] == ']' {
		pos.col++
		return arr, try[1:], nil
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

		rem = skipWhiteSpaces(rem)
		if rem[0] == ',' {
			continue
		}

		break
	}

	rem = skipWhiteSpaces(rem)
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
	digits := "0123456789"
	if json[0] == '-' {
		var neg = New(JzTypeNum)

		pos.col++
		neg, rem, err = parseNum(json[1:])
		neg.num = -neg.num
		return neg, rem, err
	}

	var n int64
	_, err = fmt.Sscanf(string(json), "%d", &n)
	if err != nil {
		err = expectOneOf(digits, json[0])
		return
	}

	nparsed := len(strconv.FormatInt(n, 10))
	num.num = n
	pos.col += nparsed
	return num, json[nparsed:], nil
}

func parseTru(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:4]) == "true" {
		bol.bol = true
		pos.col += 4
		return bol, json[4:], nil
	} else {
		err = expectString("true", json[0:4])
		return
	}
}

func parseFls(json []byte) (bol *Jzon, rem []byte, err error) {
	bol = New(JzTypeBol)
	if string(json[0:5]) == "false" {
		bol.bol = false
		pos.col += 5
		return bol, json[5:], nil
	} else {
		err = expectString("false", json[0:5])
		return
	}
}

func parseNul(json []byte) (nul *Jzon, rem []byte, err error) {
	nul = New(JzTypeNul)
	if string(json[0:4]) == "null" {
		pos.col += 4
		return nul, json[4:], nil
	} else {
		err = expectString("null", json[0:4])
		return
	}
}

func parseKVPair(json []byte) (k string, v *Jzon, rem []byte, err error) {
	json = skipWhiteSpaces(json)

	if json[0] == '"' {
		k, rem, err = parseKey(json)
		if err != nil {
			return
		}

		rem = skipWhiteSpaces(rem)
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
	// TODO: handle unicode and escaped characters
	pos.col++
	rem = json[1:]
	x := 0

	// return empty string directly
	if rem[0] == '"' {
		pos.col++
		return "", rem[1:], nil
	}

	for i, c := range rem {
		if c == '"' && i != 0 && rem[i-1] != '\\' {
			break
		}

		x += 1
	}

	if x == len(rem)-1 {
		err = expect('"', 0)
		return
	}

	pos.col += x + 1
	return string(rem[0:x]), rem[x+1:], nil
}

