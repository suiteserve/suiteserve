package ddp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"
)

func ToEjson(msg interface{}) (interface{}, error) {
	return toEjson(msg, false)
}

func toEjson(msg interface{}, escape bool) (interface{}, error) {
	a, ok := msg.([]interface{})
	if ok {
		return arrToEjson(a)
	}
	m, ok := msg.(map[string]interface{})
	if ok {
		return mapToEjson(m, escape)
	}
	return msg, nil
}

func arrToEjson(msg []interface{}) (interface{}, error) {
	a := make([]interface{}, len(msg))
	for i, v := range msg {
		v, err := toEjson(v, false)
		if err != nil {
			return nil, err
		}
		a[i] = v
	}
	return a, nil
}

func mapToEjson(msg map[string]interface{}, escape bool) (interface{}, error) {
	if len(msg) != 1 {
		m := make(map[string]interface{}, len(msg))
		for k, v := range msg {
			v, err := toEjson(v, false)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		return m, nil
	}
	if !escape {
		if v, ok := msg["$escape"]; ok {
			return toEjson(v, true)
		}
		if v, ok := msg["$date"]; ok {
			if v == nil {
				return time.Unix(0, 0), nil
			}
			num, ok := v.(json.Number)
			if !ok {
				f64, ok := v.(float64)
				if !ok {
					return nil, errors.New("expected integer for $date type")
				}
				num = json.Number(strconv.FormatFloat(f64, 'f', -1, 64))
			}
			i64, err := num.Int64()
			if err != nil {
				return nil, errors.New("expected integer for $date type")
			}
			return time.Unix(i64/1000, (i64%1000)*1e6), nil
		}
		if v, ok := msg["$binary"]; ok {
			if v == nil {
				return []byte{}, nil
			}
			s, ok := v.(string)
			if !ok {
				return nil, errors.New("expected base64 string for $binary type")
			}
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return nil, errors.New("expected base64 string for $binary type")
			}
			return b, nil
		}
	}
	for k, v := range msg {
		v, err := toEjson(v, false)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{k: v}, nil
	}
	panic("unexpected empty map")
}

func isEjson(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Struct:
		fallthrough
	case reflect.TypeOf(time.Time{}).Kind():
		return true
	case reflect.Map:
		return t.Key().Kind() == reflect.String && isEjson(t.Elem())
	case reflect.Slice:
		fallthrough
	case reflect.Ptr:
		return isEjson(t.Elem())
	default:
		return false
	}
}
