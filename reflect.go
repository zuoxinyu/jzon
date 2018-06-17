package jzon

import (
	"fmt"
	"reflect"
)

// Serializable makes those types which implemented
// it can be serialized from user custom operation
type Serializable interface {
	Serialize() []byte
}

// Deserializable makes
type Deserializable interface {
	Deserialize([]byte, *Any)
}

// TAG_NAME is the default leading tag for tagging a structure field
var TAG_NAME = "jzon"

// SetTagName lets users use custom tag name at serializing and deserializing
func SetTagName(tag string) {
	TAG_NAME = tag
}

// Serialize parses a tagged structure to Jzon
func Serialize(s Any) (jz *Jzon, err error) {
	v := reflect.ValueOf(s)
	return serialize(v)
}

func serialize(v reflect.Value) (jz *Jzon, err error) {
	t := v.Type()
	k := v.Kind()

	// TODO: serialize those types which implemented interface `jzon.Serializable`
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
		var toString func(v *reflect.Value) string
		kt := t.Key()
		_, ok := kt.MethodByName("SerializeJzon")

		if kt.Kind() == reflect.String {
			toString = func(v *reflect.Value) string {
				return v.String()
			}
		} else if ok {
			toString = func(v *reflect.Value) string {
				mt := (*v).MethodByName("SerializeJzon")
				rs := mt.Call(nil)
				if rs[0].IsValid() && rs[0].Type().Kind() == reflect.String {
					return rs[0].String()
				}

				return ""
			}
		} else {
			err = fmt.Errorf("key must be of type string or implement the interface jjzon.Serializable")
			return
		}

		var keys = v.MapKeys()
		var str string
		for _, key := range keys {
			val, err = serialize(v.MapIndex(key))
			if err != nil {
				return
			}

			str = toString(&key)

			jz.Insert(str, val)
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
	//fmt.Printf("type: %-18s | kind: %-18s\n", t, k)

	if k != reflect.Ptr || v.IsNil() {
		err = fmt.Errorf("expect nono-nil pointer, but the given value is of kind %s", k)
	}

	return deserialize(jz, &v)
}

func deserialize(jz *Jzon, v *reflect.Value) (err error) {
	t := v.Type()
	k := t.Kind()

	switch {
	case jz.Type == JzTypeObj && k == reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			v1 := v.Field(i)
			n1 := t.Field(i).Name
			t1 := v.Field(i).Type()
			k1 := t1.Kind()
			tag := t.Field(i).Tag.Get(TAG_NAME)

			if tag == "," {
				tag = t.Field(i).Name
			}

			fmt.Printf("name: %-18s | type: %-18s | kind: %-10s | tag: %s\n", n1, t1, k1, tag)

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
		m, _ := jz.Object()
		vt := t.Elem()
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}

		for jk, jv := range m {
			pv1 := reflect.New(vt)
			v1 := reflect.Indirect(pv1)
			k1 := reflect.ValueOf(jk)
			err = deserialize(jv, &v1)
			if err != nil {
				return
			}

			v.SetMapIndex(k1, v1)
		}

	case jz.Type == JzTypeArr && k == reflect.Slice:
		a, _ := jz.Array()
		if v.IsNil() {
			v.Set(reflect.MakeSlice(t, len(a), len(a)))
		}

		for i, jv := range a {
			v1 := v.Index(i)
			deserialize(jv, &v1)
			v1.Set(v1)
		}

	case jz.Type == JzTypeArr && k == reflect.Array:
		a, _ := jz.Array()
		l := len(a)
		for i := 0; i < l; i++ {
			v1 := v.Index(i)
			deserialize(a[i], &v1)
		}

	case jz.Type == JzTypeStr && k == reflect.String:
		str, _ := jz.String()
		v.SetString(str)

	case jz.Type == JzTypeInt && (k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64):
		n, _ := jz.Integer()
		v.SetInt(n)

	case jz.Type == JzTypeInt && (k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32):
		n, _ := jz.Integer()
		v.SetInt(n)

	case jz.Type == JzTypeFlt && (k == reflect.Float64 || k == reflect.Float32):
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
