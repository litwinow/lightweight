package lightweight

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := []byte{}
	return doMarshal(buf, v)
}

func doMarshal(buf []byte, v interface{}) ([]byte, error) {
	switch tv := v.(type) {
	case int:
		return binary.AppendVarint(buf, int64(tv)), nil
	case uint:
		return binary.AppendUvarint(buf, uint64(tv)), nil
	case string:
		return marshalString(buf, tv)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return marshalSlice(buf, rv)
	}
	return nil, fmt.Errorf("bad type %T", v)
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

func marshalString(buf []byte, s string) ([]byte, error) {
	asBytes := []byte(s)
	buf, err := doMarshal(buf, len(asBytes))
	if err != nil {
		return nil, err
	}
	return append(buf, asBytes...), nil
}

func Unmarshal(b []byte, v interface{}) error {
	r := bytes.NewReader(b)
	return doUnmarshal(r, v)
}

func doUnmarshal(r io.ByteReader, v interface{}) error {
	switch tv := v.(type) {
	case *int:
		vv, err := binary.ReadVarint(r)
		if err != nil {
			return err
		}
		*tv = int(vv)
		return nil
	case *uint:
		vv, err := binary.ReadUvarint(r)
		if err != nil {
			return err
		}
		*tv = uint(vv)
		return nil
	case *string:
		return unmarshalString(r, tv)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.Elem().Kind() == reflect.Slice {
			return unmarshalSlice(r, rv.Elem())
		} else if rv.Elem().Kind() == reflect.Array {
			return unmarshalArray(r, rv.Elem())
		}
	}

	return fmt.Errorf("bad type %T", v)
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

func unmarshalString(r io.ByteReader, v *string) error {
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
	*v = string(asBytes)
	return nil
}
