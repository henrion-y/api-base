package georedis

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/henrion-y/api-base/geo"
)

func unpackValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		if !v.IsNil() {
			v = v.Elem()
		}
	}
	return v
}

func toString(v reflect.Value) (string, error) {
	if v.Kind() != reflect.Slice {
		return "", fmt.Errorf("to string fail: %v", v.Kind())
	}

	b := v.Bytes()
	return string(b), nil
}

func toFloat64(v reflect.Value) (float64, error) {
	s, err := toString(v)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

func rawToNeighbors(r interface{}, options ...geo.Option) ([]*geo.Neighbor, error) {
	v := reflect.ValueOf(r)

	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("wrong type: %v", v.Kind())
	}

	results := make([]*geo.Neighbor, v.Len())
	var err error
	for i := 0; i < v.Len(); i++ {
		results[i], err = geo.NewNeighbor(unpackValue(v.Index(i)), options...)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
