package jzon

import (
	"errors"
)

type ValueType int
type Any interface{}

const (
	JzTypeStr ValueType = iota
	JzTypeNum
	JzTypeBol
	JzTypeObj
	JzTypeArr
	JzTypeNul
)

var typeStrings = []string{
	"JzTypeStr",
	"JzTypeNum",
	"JzTypeBol",
	"JzTypeObj",
	"JzTypeArr",
	"JzTypeNul",
}

// Jzon defines a JSON node
type Jzon struct {
	Type ValueType
	arr  []*Jzon
	obj  map[string]*Jzon
	str  string
	num  int64
	flt  float64
	bol  bool
}

// New allocates a Jzon node on heap
func New(t ValueType) *Jzon {
	v := Jzon{}
	v.Type = t
	switch t {
	case JzTypeStr:
	case JzTypeNum:
	case JzTypeBol:
	case JzTypeObj: v.obj = make(map[string]*Jzon)
	case JzTypeArr: v.arr = make([]*Jzon, 0)
	case JzTypeNul:
	}

	return &v
}

// Parse parses string to Jzon
func Parse(json []byte) (*Jzon, error) {
	pos.col = 0
	pos.row = 0
	jz, _, err := parse(json)
	return jz, err
}

// Number returns number value
func (jz *Jzon) Number() (n int64, err error) {
	if jz.Type != JzTypeNum {
		return n, expectTypeOf(JzTypeNum, jz.Type)
	}

	return jz.num, nil
}

// String returns string value
func (jz *Jzon) String() (s string, err error) {
	if jz.Type != JzTypeStr {
		return s, expectTypeOf(JzTypeStr, jz.Type)
	}

	return jz.str, nil
}

// Bool returns bool value
func (jz *Jzon) Bool() (b bool, err error) {
	if jz.Type != JzTypeBol {
		return b, expectTypeOf(JzTypeBol, jz.Type)
	}

	return jz.bol, nil
}

// Null returns null value (as nil)
func (jz *Jzon) Null() (n Any, err error) {
	if jz.Type != JzTypeNul {
		return nil, expectTypeOf(JzTypeNul, jz.Type)
	}

	return nil, nil
}

// Array returns array value
func (jz *Jzon) Array() (a []*Jzon, err error) {
	if jz.Type != JzTypeArr {
		return a, expectTypeOf(JzTypeArr, jz.Type)
	}

	return jz.arr, nil
}

// Object returns object value
func (jz *Jzon) Object() (m map[string]*Jzon, err error) {
	if jz.Type != JzTypeObj {
		return m, expectTypeOf(JzTypeObj, jz.Type)
	}

	return jz.obj, nil
}

// ValueOf finds the value of the key in an object, if the node isn't
// an object, or the key doesn't exist, an error will be thrown out
func (jz *Jzon) ValueOf(k string) (v *Jzon, err error) {
	if jz.Type != JzTypeObj {
		err = expectTypeOf(JzTypeObj, jz.Type)
		return
	}

	v, ok := jz.obj[k]
	if !ok {
		err = errors.New("key doesn't exist")
		return
	}

	return v, nil
}

// ValueAt finds the value at the index in an array, if the node isn't
// an array, or the index is out of bound, an error will be thrown out
func (jz *Jzon) ValueAt(i int) (v *Jzon, err error) {
	if jz.Type != JzTypeArr {
		err = expectTypeOf(JzTypeArr, jz.Type)
		return
	}

	if i < 0 || i >= len(jz.arr) {
		err = errors.New("index is out of bound")
		return
	}

	return jz.arr[i], nil
}

// Has returns whether the object has the key, if it's not an object
// an error will be thrown out
func (jz *Jzon) Has(k string) (has bool, err error) {
	if jz.Type != JzTypeObj {
		err = expectTypeOf(JzTypeArr, jz.Type)
		return
	}

	_, ok := jz.obj[k]
	return ok, nil
}

// Keys returns the total keys as an string slice in an object,
// if it's not an object, an error will be thrown out
func (jz *Jzon) Keys() (ks []string, err error) {
	if jz.Type != JzTypeObj {
		err = expectTypeOf(JzTypeObj, jz.Type)
		return
	}

	for k, _ := range jz.obj {
		ks = append(ks, k)
	}

	return ks, nil
}

// Length returns the length of an object or an array, if it's an object, then
// just return the number of keys, otherwise an error will be thrown out
func (jz *Jzon) Length() (l int, err error) {
	if jz.Type == JzTypeArr {
		return len(jz.arr), nil
	}

	if jz.Type == JzTypeObj {
		return len(jz.obj), nil
	}

	err = errors.New("expect node of type JzTypeArr or JzTypeObj" +
		", but the real type is " + typeStrings[jz.Type])
	l = -1
	return
}

// IsNull returns whether it equals to null
func (jz *Jzon) IsNull() bool {
	return jz.Type == JzTypeNul
}

// Insert inserts a key with a value in an object, if it's not
// an object, an error will be thrown out
func (jz *Jzon) Insert(k string, v *Jzon) (err error) {
	if jz.Type != JzTypeObj {
		err = expectTypeOf(JzTypeObj, jz.Type)
		return
	}

	jz.obj[k] = v
	return nil
}

// Append appends a node in an array, if it's not an array, an error
// will be thrown out
func (jz *Jzon) Append(v *Jzon) (err error) {
	if jz.Type != JzTypeArr {
		err = expectTypeOf(JzTypeArr, jz.Type)
		return
	}

	jz.arr = append(jz.arr, v)
	return nil
}

// Delete removes a key in an object, it's safe to delete a node which
// doesn't exist, if it's not an object, an error will be thrown out
func (jz *Jzon) Delete(k string) (err error) {
	if jz.Type != JzTypeObj {
		err = expectTypeOf(JzTypeObj, jz.Type)
		return
	}

	delete(jz.obj, k)
	return nil
}

// Remove removes a node at index in an array, it's safe to delete an index
// which doesn't exist, if it's not an object, an error will be thrown out
func (jz *Jzon) Remove(i int) (err error) {
	if jz.Type != JzTypeArr {
		err = expectTypeOf(JzTypeArr, jz.Type)
		return
	}

	if i > len(jz.arr) || i < 0 {
		err = errors.New("index is out of bounds")
		return
	}

	newArr := jz.arr[0:i]

	for _, v := range jz.arr[i:] {
		newArr = append(newArr, v)
	}

	jz.arr = newArr

	return nil
}

// AMap is just map for array
func (jz *Jzon) AMap(itFunc func(g *Jzon) []Any) []Any {
	if jz.Type != JzTypeArr {
	}

	var res []Any

	for _, node := range jz.arr {
		res = append(res, itFunc(node))
	}

	return res
}

// AFilter is just filter for array
func (jz *Jzon) AFilter(predictFunc func(g *Jzon) bool) []*Jzon {
	var res []*Jzon

	for _, node := range jz.arr {
		if predictFunc(node) {
			res = append(res, node)
		}
	}

	return res
}

// AReduce is just reduce for array
func (jz *Jzon) Reduce(init Any, acc func(a *Jzon, b Any) Any) Any {
	var res = init

	for _, node := range jz.arr {
		res = acc(node, res)
	}

	return res
}

// OMap is just map for object
func (jz *Jzon) OMap(itFunc func(key string, g *Jzon) Any) []Any {
	if jz.Type != JzTypeObj {

	}

	var res []Any = nil

	for k, v := range jz.obj {
		res = append(res, itFunc(k, v))
	}

	return res
}

// OFilter is just filter for object
func (jz *Jzon) OFilter(predictFunc func(key string, value *Jzon) bool) Jzon {
	var res = *jz

	for k, v := range res.obj {
		if !predictFunc(k, v) {
			res.Delete(k)
		}
	}

	return res
}

// Map is just flat map which retrieves on each children node in the Jzon
func (jz *Jzon) Map(mapFunc func(string, *Jzon) Any) Any {
	switch jz.Type {
	case JzTypeArr:
		return jz.AMap(func(j *Jzon) (res []Any) {
			return append(res, mapFunc("", j))
		})
	case JzTypeObj:
		return jz.OMap(mapFunc)
	default:
		return mapFunc("", jz)
	}
}
