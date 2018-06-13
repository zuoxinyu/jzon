package jzon

import "reflect"

// Serialize parses a tagged structure to Jzon
func Serialize(s Any) (jz Jzon) {
	// TODO:
	return
}

// Deserialize parses JSON string to a structure of arbitrary type
func Deserialize(json string, s Any) (err error) {
	// TODO:
	return
}

// Value returns value of type interface{}, for maps
// it's a map[string]*Jzon, for arrays it's []*Jzon
func (jz *Jzon) Value(t ValueType) (v Any, err error) {
	if jz.Type != t {
		err = expectTypeOf(t, jz.Type)
		return
	}

	switch t {
	case JzTypeStr: v = jz.str
	case JzTypeInt: v = jz.num
	case JzTypeBol: v = jz.bol
	case JzTypeObj: v = jz.obj
	case JzTypeArr: v = jz.arr
	case JzTypeNul: v = nil
	}

	return v, nil
}

// NewFromAny allocates and assigns a Jzon node on the heap, if the given `v` is of type
// `*Jzon`, it performs as deep clone, if `v` is of type `Jzon`, it performs as shallow
// clone, otherwise it converts value of built-in types to an appropriate `Jzon` value
func NewFromAny(v Any) *Jzon {
	jz := Jzon{}

	if v == nil {
		jz.Type = JzTypeNul
		return &jz
	}

	switch v.(type) {
	case int, int16, int32, int64, uint, uint16, uint32:
		jz.Type = JzTypeInt
		jz.num = reflect.ValueOf(v).Int()

	case float32, float64:
		jz.Type = JzTypeInt
		jz.flt = reflect.ValueOf(v).Float()

	case string:
		jz.Type = JzTypeStr
		jz.str = v.(string)

	case []byte:
		jz.Type = JzTypeStr
		jz.str = string(v.([]byte))

	case bool:
		jz.Type = JzTypeBol
		jz.bol = v.(bool)

	case []*Jzon:
		jz.Type = JzTypeArr
		jz.arr = v.([]*Jzon)

	case map[string]*Jzon:
		jz.Type = JzTypeObj
		jz.obj = v.(map[string]*Jzon)

	case *Jzon:
		// TODO: deep clone
		jz = *(v.(*Jzon))

	case Jzon:
		// TODO: shallow clone
		jz = v.(Jzon)

	default:
		return nil
	}

	return &jz
}

