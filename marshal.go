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
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
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
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.Elem().Kind() == reflect.Slice {
			return unmarshalSlice(r, rv.Elem())
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
		v := reflect.New(v.Type().Elem())
		if err := doUnmarshal(r, v.Interface()); err != nil {
			return err
		}
		s.Index(i).Set(reflect.Indirect(v))
	}
	v.Set(s)
	return nil
}