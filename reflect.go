package jzon

import (
	"fmt"
	"reflect"
)

// Serializable is the interface makes those types which implemented
// it can be serialize from user custom operations
type Serializable interface {
	Serialize() string
}

// Deserializable makes
type Deserializable interface {
	Deserialize() string
}

// TAG_NAME is the default leading tag when tag an structure field
const TAG_NAME = "jzon"

// Serialize parses a tagged structure to Jzon
func Serialize(s Any) (jz *Jzon, err error) {
	v := reflect.ValueOf(s)
	return serialize(v)
}

func serialize(v reflect.Value) (jz *Jzon, err error) {
	t := v.Type()
	k := v.Kind()

	// TODO: serialize those type which implements interface jzon.Serializable
	// method := v.MethodByName("Serialize")
	// if method.IsValid() {
	//     rs := method.Call(nil)
	//     if len(rs) == 1 && rs[0].Kind() == reflect.String { }
	// }

	switch k {
	case reflect.Struct:
		jz = New(JzTypeObj)
		var val *Jzon
		for i := 0; i < t.NumField(); i++ {
			key := t.Field(i).Tag.Get(TAG_NAME)
			val, err = serialize(v.Field(i))
			if err != nil {
				return
			}
			// t1 := v.Field(i).Type()
			// k1 := v.Field(i).Kind()
			// fmt.Printf("type: %-12s | kind: %-10s | tag: %s\n", t1, k1, key)
			// ignore field without `jzon` tag or was marked as ignorable
			// TODO: implement the `omitempty` and `string` tag
			if key == "," {
				key = t.Field(i).Name
			}

			if key != "" && key != "-" {
				jz.Insert(key, val)
			}
		}

	case reflect.Map:
		jz = New(JzTypeObj)
		var val *Jzon
		var keys = v.MapKeys()
		for _, key := range keys {
			if key.Kind() != reflect.String {
				err = fmt.Errorf("only type map[string]T can be serialized")
				return
			}

			val, err = serialize(v.MapIndex(key))
			if err != nil {
				return
			}
			jz.Insert(key.String(), val)
		}

	case reflect.Slice:
		if v.IsNil() {
			jz = New(JzTypeNul)
			return
		}
		fallthrough

	case reflect.Array:
		jz = New(JzTypeArr)
		var val *Jzon
		for i := 0; i < v.Len(); i++ {
			val, err = serialize(v.Index(i))
			if err != nil {
				return
			}
			jz.Append(val)
		}

	case reflect.String:
		jz = NewFromAny(v.String())

	case reflect.Float32, reflect.Float64:
		jz = NewFromAny(v.Float())

	case reflect.Bool:
		jz = NewFromAny(v.Bool())

	case reflect.Ptr:
		if v.IsNil() {
			jz = New(JzTypeNul)
		} else {
			err = fmt.Errorf("only nil ptr can be serialized")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		jz = NewFromAny(v.Int())

	default:
		err = fmt.Errorf("can not serialize variable of kind [%s] to Jzon", k)
	}

	return
}

// Deserialize parses JSON string to a structure of arbitrary type
func Deserialize(json []byte, ptr Any) (err error) {
	jz, err := Parse(json)
	if err != nil {
		return
	}
	p := reflect.ValueOf(ptr)
	v := reflect.Indirect(p)
	t := v.Type()
	k := t.Kind()
	fmt.Printf("type: %-18s | kind: %-18s\n", t, k)

	if k != reflect.Ptr || v.IsNil() {
		err = fmt.Errorf("expect nono-nil pointer, but the given value is of kind %s", k)
	}

	return deserialize(jz, &v)
}

func deserialize(jz *Jzon, v *reflect.Value) (err error) {
	t := v.Type()
	k := t.Kind()
	// fmt.Printf("type: %-18s | kind: %-10s | tag: \n", t, k)

	switch {
	case jz.Type == JzTypeObj && k == reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			v1 := v.Field(i)
			t1 := v.Field(i).Type()
			k1 := t1.Kind()
			tag := t.Field(i).Tag.Get(TAG_NAME)

			if tag == "," {
				tag = t.Field(i).Name
			}

			fmt.Printf("type: %-18s | kind: %-10s | tag: %s\n", t1, k1, tag)

			if tag == "" || tag == "-" {
				continue
			}

			var jv *Jzon
			jv, err = jz.ValueOf(tag)

			if err != nil {
				return
			}

			err = deserialize(jv, &v1)
		}

	case jz.Type == JzTypeObj && k == reflect.Map:
		// TODO: deserialize those types which implemented `Deserialize()`
		vks := v.MapKeys()
		m, _ := jz.Object()

		for _, vk := range vks {
			v1 := v.MapIndex(vk)

			k1 := vk.String()
			jv, ok := m[k1]

			if !ok {
				return fmt.Errorf("no such key")
			} else {
				err = deserialize(jv, &v1)
			}
		}

	case jz.Type == JzTypeArr && (k == reflect.Slice || k == reflect.Array):
		l := v.Len()
		a, _ := jz.Array()

		for i := 0; i < l; i++ {
			v1 := v.Index(i)
			deserialize(a[i], &v1)
		}

	case jz.Type == JzTypeStr && k == reflect.String:
		str, _ := jz.String()
		v.SetString(str)

	case jz.Type == JzTypeInt && k == reflect.Int:
		n, _ := jz.Integer()
		v.SetInt(n)

	case jz.Type == JzTypeFlt && k == reflect.Float64:
		f, _ := jz.Float()
		v.SetFloat(f)

	case jz.Type == JzTypeBol && k == reflect.Bool:
		b, _ := jz.Bool()
		v.SetBool(b)

	case jz.Type == JzTypeNul && k == reflect.Slice:
		v.SetLen(0)
	}
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
	case JzTypeStr:
		v = jz.str
	case JzTypeInt:
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
		jz.Type = JzTypeFlt
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
