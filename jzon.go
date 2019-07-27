// Package jzon implements parsing and encoding  of JSON (Javascript Object Notation)
// defined in ECMA-404. The most different feature between jzon and encoding/json is
// that jzon defines a explicit structure to notate a JSON object and supplies some
// utility methods to manipulate JSON objects. Package jzon is NOT compatible with
// encoding/json completely. The main design goal of jzon is enhanced validators.
package jzon

import (
	"errors"
	"fmt"
)

// ValueType is the alias of int
type ValueType = int

// Any is the alias of interface{}
type Any = interface{}

// Jzon defines a JSON node
type Jzon struct {
	Type ValueType
	data Any
}

// Types
const (
	JzTypeStr ValueType = iota
	JzTypeInt
	JzTypeFlt
	JzTypeBol
	JzTypeObj
	JzTypeArr
	JzTypeNul
)

var typeStrings = []string{
	"JzTypeStr",
	"JzTypeFlt",
	"JzTypeInt",
	"JzTypeBol",
	"JzTypeObj",
	"JzTypeArr",
	"JzTypeNul",
}

// New allocates an empty Jzon node on the heap
func New(t ValueType) *Jzon {
	v := Jzon{}
	v.Type = t
	switch t {
	case JzTypeStr:
	case JzTypeInt:
	case JzTypeFlt:
	case JzTypeBol:
	case JzTypeObj:
		v.data = make(map[string]*Jzon)
	case JzTypeArr:
		v.data = make([]*Jzon, 0)
	case JzTypeNul:
	}

	return &v
}

// Parse parses string to Jzon, any errors occurred in the parsing will be thrown out
func Parse(json []byte) (jz *Jzon, err error) {
	// in the implements of function `parse()` we don't handle any
	// exceptions about slice bounds out of range. here we simply
	// throw the error recovered from those unhandled exceptions
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("maybe out of bound: %v", e)
		}
	}()

	pos.col = 0
	pos.row = 0
	jz, rem, err := parse(json)
	if len(rem) == 0 {
		return jz, err
	}

	rem = trimWhiteSpaces(rem)
	if len(rem) == 0 {
		return jz, err
	}
	return nil, expectString("end of file", rem)
}

// Object returns object value, if it's not an object, an error will be thrown out
func (jz *Jzon) Object() (m map[string]*Jzon, err error) {
	if jz.Type != JzTypeObj {
		return m, expectTypeOf(JzTypeObj, jz.Type)
	}

	return jz.data.(map[string]*Jzon), nil
}

// Array returns array value, if it's not an array, an error will be thrown out
func (jz *Jzon) Array() (a []*Jzon, err error) {
	if jz.Type != JzTypeArr {
		return a, expectTypeOf(JzTypeArr, jz.Type)
	}

	return jz.data.([]*Jzon), nil
}

// String returns string value, if it's not a string, an error will be thrown out
func (jz *Jzon) String() (s string, err error) {
	if jz.Type != JzTypeStr {
		return s, expectTypeOf(JzTypeStr, jz.Type)
	}

	return jz.data.(string), nil
}

// Integer returns integer value, if it's not an integer, an error will be thrown out
func (jz *Jzon) Integer() (n int64, err error) {
	if jz.Type != JzTypeInt {
		return n, expectTypeOf(JzTypeInt, jz.Type)
	}

	return jz.data.(int64), nil
}

// Float returns float64 value, if it's not a float, an error will be thrown out
func (jz *Jzon) Float() (f float64, err error) {
	if jz.Type != JzTypeFlt {
		return f, expectTypeOf(JzTypeInt, jz.Type)
	}

	return jz.data.(float64), nil
}

// Null returns nil value, if it's not a boolean, an error will be thrown out
func (jz *Jzon) Null() (n Any, err error) {
	if jz.Type != JzTypeNul {
		return nil, expectTypeOf(JzTypeNul, jz.Type)
	}

	return nil, nil
}

// Bool returns bool value, if it's not a null, an error will be thrown out
func (jz *Jzon) Bool() (b bool, err error) {
	if jz.Type != JzTypeBol {
		return b, expectTypeOf(JzTypeBol, jz.Type)
	}

	return jz.data.(bool), nil
}

// Length returns the length of an object or an array, if it is an object,
// just returns the number of keys, otherwise an error will be thrown out
func (jz *Jzon) Length() (l int, err error) {
	if jz.Type == JzTypeArr {
		return len(jz.data.([]*Jzon)), nil
	}

	if jz.Type == JzTypeObj {
		return len(jz.data.(map[string]*Jzon)), nil
	}

	return -1, errors.New("expect node of type JzTypeArr or JzTypeObj" +
		", but the real type is " + typeStrings[jz.Type])
}

// ValueOf finds the value of the key in an object, if it's not an object
// or the key does not exist in this object, an error will be thrown out
func (jz *Jzon) ValueOf(k string) (v *Jzon, err error) {
	if jz.Type != JzTypeObj {
		return v, expectTypeOf(JzTypeObj, jz.Type)
	}

	v, ok := jz.data.(map[string]*Jzon)[k]
	if !ok {
		err = errors.New("key doesn't exist")
		return
	}

	return v, nil
}

// ValueAt finds the value at the index in an array, if it's not an
// array or the index is out of bound, an error will be thrown out
func (jz *Jzon) ValueAt(i int) (v *Jzon, err error) {
	if jz.Type != JzTypeArr {
		return v, expectTypeOf(JzTypeArr, jz.Type)
	}

	if i < 0 || i >= len(jz.data.([]*Jzon)) {
		err = errors.New("index is out of bound")
		return
	}

	return jz.data.([]*Jzon)[i], nil
}

// Keys returns all keys as an string slice in object,
// if it's not an object, an error will be thrown out
func (jz *Jzon) Keys() (ks []string, err error) {
	if jz.Type != JzTypeObj {
		return ks, expectTypeOf(JzTypeObj, jz.Type)
	}

	for k := range jz.data.(map[string]*Jzon) {
		ks = append(ks, k)
	}

	return ks, nil
}

// Has returns if this object has the given key, if
// it's not an object, an error will be thrown out
func (jz *Jzon) Has(k string) (has bool, err error) {
	if jz.Type != JzTypeObj {
		return has, expectTypeOf(JzTypeArr, jz.Type)
	}

	_, ok := jz.data.(map[string]*Jzon)[k]
	return ok, nil
}

// IsNull returns whether it equals to null
func (jz *Jzon) IsNull() bool {
	return jz.Type == JzTypeNul
}

// Insert inserts a key with a node in an object, or replaces the value for the key
// when the key already exists. if it's not an object, an error will be thrown out
func (jz *Jzon) Insert(k string, v *Jzon) (err error) {
	if jz.Type != JzTypeObj {
		return expectTypeOf(JzTypeObj, jz.Type)
	}

	jz.data.(map[string]*Jzon)[k] = v
	return nil
}

// Append appends a node after an array, if it's not an array, an error will be thrown out
func (jz *Jzon) Append(v *Jzon) (err error) {
	if jz.Type != JzTypeArr {
		return expectTypeOf(JzTypeArr, jz.Type)
	}

	jz.data = append(jz.data.([]*Jzon), v)
	return nil
}

// Delete removes a key in an object, it's safe to delete a key which
// doesn't exist, if it's not an object, an error will be thrown out
func (jz *Jzon) Delete(k string) (err error) {
	if jz.Type != JzTypeObj {
		return expectTypeOf(JzTypeObj, jz.Type)
	}

	delete(jz.data.(map[string]*Jzon), k)
	return nil
}

// Remove removes an index in an array, it's safe to delete an index
// doesn't exist, if it's not an array, an error will be thrown out
func (jz *Jzon) Remove(i int) (err error) {
	if jz.Type != JzTypeArr {
		return expectTypeOf(JzTypeArr, jz.Type)
	}

	if i > len(jz.data.([]*Jzon)) || i < 0 {
		return errors.New("index is out of bounds")
	}

	newArr := jz.data.([]*Jzon)[0:i]

	for _, v := range jz.data.([]*Jzon)[i:] {
		newArr = append(newArr, v)
	}

	jz.data = newArr

	return nil
}

// AMap is just map for array, if it's not an array, an error will be thrown out
func (jz *Jzon) AMap(itFunc func(g *Jzon) Any) (res []Any, err error) {
	if jz.Type != JzTypeArr {
		return res, expectTypeOf(JzTypeArr, jz.Type)
	}

	res = make([]Any, 0)

	for _, node := range jz.data.([]*Jzon) {
		res = append(res, itFunc(node))
	}

	return res, nil
}

// AFilter is just filter for array, if it's not an array, an error will be thrown out
func (jz *Jzon) AFilter(predictFunc func(g *Jzon) bool) (res []*Jzon, err error) {
	if jz.Type != JzTypeArr {
		return res, expectTypeOf(JzTypeArr, jz.Type)
	}

	res = make([]*Jzon, 0)

	for _, node := range jz.data.([]*Jzon) {
		if predictFunc(node) {
			res = append(res, node)
		}
	}

	return res, nil
}

// AReduce is just reduce for array, if it's not an object, an error will be thrown out
func (jz *Jzon) AReduce(init Any, acc func(a *Jzon, b Any) Any) (res Any, err error) {
	if jz.Type != JzTypeArr {
		return res, expectTypeOf(JzTypeArr, jz.Type)
	}

	res = init

	for _, node := range jz.data.([]*Jzon) {
		res = acc(node, res)
	}

	return res, nil
}

// OMap is just map for object, if it's not an object, an error will be thrown out
func (jz *Jzon) OMap(itFunc func(key string, g *Jzon) Any) (res []Any, err error) {
	if jz.Type != JzTypeObj {
		return res, expectTypeOf(JzTypeObj, jz.Type)
	}

	res = make([]Any, 0)

	for k, v := range jz.data.(map[string]*Jzon) {
		res = append(res, itFunc(k, v))
	}

	return res, nil
}

// OFilter is just filter for object, if it's not an object, an error will be thrown out
func (jz *Jzon) OFilter(predictFunc func(key string, value *Jzon) bool) (res *Jzon, err error) {
	if jz.Type != JzTypeObj {
		return res, expectTypeOf(JzTypeObj, jz.Type)
	}

	var tmp = *jz

	for k, v := range tmp.data.(map[string]*Jzon) {
		if !predictFunc(k, v) {
			tmp.Delete(k)
		}
	}

	res = &tmp
	return res, nil
}

// Map is flat map which retrieves on each children node of itself
func (jz *Jzon) Map(mapFunc func(string, *Jzon) Any) (res Any, err error) {
	switch jz.Type {
	case JzTypeArr:
		return jz.AMap(func(j *Jzon) (res Any) { return mapFunc("", j) })

	case JzTypeObj:
		return jz.OMap(mapFunc)

	default:
		return mapFunc("", jz), err
	}
}
