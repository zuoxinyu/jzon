package jzon

// Serialize parses a tagged structure to Jzon
func Serialize(s Any) Jzon {
	// TODO:
	return Jzon{}
}

// Deserialize parses JSON string to a structure of arbitrary type
func Deserialize(json string, s Any) error {
	// TODO:
	return nil
}

// Value returns any value of type interface{},
// for maps its map[string]*Jzon, for arrays it's []*Jzon
func (jz *Jzon) Value(t ValueType) (v Any, err error) {
	if jz.Type != t {
		err = expectTypeOf(t, jz.Type)
		return
	}

	switch t {
	case JzTypeStr:
		v = jz.str
	case JzTypeNum:
		v = jz.num
	case JzTypeBol:
		v = jz.bol
	case JzTypeObj:
		v = jz.obj
	case JzTypeArr:
		v = jz.arr
	case JzTypeNul:
		v = nil
	}

	return v, nil
}

