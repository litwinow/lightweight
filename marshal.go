package lightweight

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := []byte{}
	return doMarshal(buf, v)
}

func doMarshal(buf []byte, v interface{}) ([]byte, error) {
	switch rv := reflect.ValueOf(v); {
	case rv.Kind() == reflect.Bool:
		return marshalBool(buf, rv), nil
	case rv.CanUint():
		return marshalUint(buf, rv), nil
	case rv.CanInt():
		return marshalInt(buf, rv), nil
	case rv.Kind() == reflect.Float32:
		return marshalFloat32(buf, rv), nil
	case rv.Kind() == reflect.Float64:
		return marshalFloat64(buf, rv), nil
	case rv.Kind() == reflect.String:
		return marshalString(buf, rv)
	case rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array:
		return marshalSlice(buf, rv)
	case rv.Kind() == reflect.Map:
		return marshalMap(buf, rv)
	case rv.Kind() == reflect.Struct:
		return marshalStruct(buf, rv)
	default:
		return nil, fmt.Errorf("bad type %T", v)
	}
}

func marshalSlice(buf []byte, v reflect.Value) ([]byte, error) {
	buf, err := doMarshal(buf, v.Len())
	if err != nil {
		return nil, err
	}
	for i := 0; i < v.Len(); i++ {
		var err error
		buf, err = doMarshal(buf, v.Index(i).Interface())
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func marshalMap(buf []byte, v reflect.Value) ([]byte, error) {
	buf, err := doMarshal(buf, v.Len())
	if err != nil {
		return nil, err
	}
	for _, key := range v.MapKeys() {
		var err error
		buf, err = doMarshal(buf, key.Interface())
		if err != nil {
			return nil, err
		}
		val := v.MapIndex(key)
		buf, err = doMarshal(buf, val.Interface())
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func marshalString(buf []byte, v reflect.Value) ([]byte, error) {
	asBytes := []byte(v.Interface().(string))
	buf, err := doMarshal(buf, len(asBytes))
	if err != nil {
		return nil, err
	}
	return append(buf, asBytes...), nil
}

func marshalInt(buf []byte, v reflect.Value) []byte {
	return binary.AppendVarint(buf, v.Int())
}

func marshalUint(buf []byte, v reflect.Value) []byte {
	return binary.AppendUvarint(buf, v.Uint())
}

func marshalBool(buf []byte, v reflect.Value) []byte {
	vByte := byte(0)
	if v.Interface().(bool) {
		vByte = 1
	}
	return append(buf, vByte)
}

func marshalFloat32(buf []byte, v reflect.Value) []byte {
	return binary.AppendUvarint(buf, uint64(math.Float32bits(v.Interface().(float32))))
}

func marshalFloat64(buf []byte, v reflect.Value) []byte {
	return binary.AppendUvarint(buf, math.Float64bits(v.Interface().(float64)))
}

func marshalStruct(buf []byte, v reflect.Value) ([]byte, error) {
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			var err error
			buf, err = doMarshal(buf, v.Field(i).Interface())
			if err != nil {
				return nil, err
			}
		}
	}
	return buf, nil
}

func Unmarshal(b []byte, v interface{}) error {
	r := bytes.NewReader(b)
	return doUnmarshal(r, v)
}

func doUnmarshal(r io.ByteReader, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("v must be pointer")
	}

	switch rv := rv.Elem(); {
	case rv.Kind() == reflect.Bool:
		return unmarshalBool(r, rv)
	case rv.CanUint():
		return unmarshalUint(r, rv)
	case rv.CanInt():
		return unmarshalInt(r, rv)
	case rv.Kind() == reflect.Float32:
		return unmarshalFloat32(r, rv)
	case rv.Kind() == reflect.Float64:
		return unmarshalFloat64(r, rv)
	case rv.Kind() == reflect.String:
		return unmarshalString(r, rv)
	case rv.Kind() == reflect.Slice:
		return unmarshalSlice(r, rv)
	case rv.Kind() == reflect.Array:
		return unmarshalArray(r, rv)
	case rv.Kind() == reflect.Map:
		return unmarshalMap(r, rv)
	case rv.Kind() == reflect.Struct:
		return unmarshalStruct(r, rv)
	default:
		return fmt.Errorf("bad type %T", v)
	}

}

func unmarshalSlice(r io.ByteReader, v reflect.Value) error {
	var len int
	if err := doUnmarshal(r, &len); err != nil {
		return err
	}
	s := reflect.MakeSlice(v.Type(), len, len)
	for i := 0; i < len; i++ {
		val := reflect.New(v.Type().Elem())
		if err := doUnmarshal(r, val.Interface()); err != nil {
			return err
		}
		s.Index(i).Set(reflect.Indirect(val))
	}
	v.Set(s)
	return nil
}

func unmarshalArray(r io.ByteReader, v reflect.Value) error {
	var len int
	if err := doUnmarshal(r, &len); err != nil {
		return err
	}
	for i := 0; i < len; i++ {
		val := reflect.New(v.Type().Elem())
		if err := doUnmarshal(r, val.Interface()); err != nil {
			return err
		}
		reflect.Indirect(v).Index(i).Set(reflect.Indirect(val))
	}
	return nil
}

func unmarshalMap(r io.ByteReader, v reflect.Value) error {
	var len int
	if err := doUnmarshal(r, &len); err != nil {
		return err
	}
	m := reflect.MakeMapWithSize(v.Type(), len)
	for i := 0; i < len; i++ {
		key := reflect.New(v.Type().Key())
		if err := doUnmarshal(r, key.Interface()); err != nil {
			return err
		}
		value := reflect.New(v.Type().Elem())
		if err := doUnmarshal(r, value.Interface()); err != nil {
			return err
		}
		m.SetMapIndex(reflect.Indirect(key), reflect.Indirect(value))
	}
	v.Set(m)
	return nil
}

func unmarshalString(r io.ByteReader, v reflect.Value) error {
	var len int
	if err := doUnmarshal(r, &len); err != nil {
		return err
	}
	asBytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		asBytes[i] = b
	}
	v.Set(reflect.ValueOf(string(asBytes)))
	return nil
}

func unmarshalInt(r io.ByteReader, v reflect.Value) error {
	vv, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}
	v.SetInt(vv)
	return nil
}

func unmarshalUint(r io.ByteReader, v reflect.Value) error {
	vv, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	v.SetUint(vv)
	return nil
}

func unmarshalBool(r io.ByteReader, v reflect.Value) error {
	vv, err := r.ReadByte()
	if err != nil {
		return err
	}
	val := false
	if vv == 1 {
		val = true
	}
	v.Set(reflect.ValueOf(val))
	return nil
}

func unmarshalFloat32(r io.ByteReader, v reflect.Value) error {
	vv, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(math.Float32frombits(uint32(vv))))
	return nil
}

func unmarshalFloat64(r io.ByteReader, v reflect.Value) error {
	vv, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	v.Set(reflect.ValueOf(math.Float64frombits(uint64(vv))))
	return nil
}

func unmarshalStruct(r io.ByteReader, v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			if err := doUnmarshal(r, v.Field(i).Addr().Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}
