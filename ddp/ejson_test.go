package ddp

import (
	"reflect"
	"testing"
	"time"
)

var ejsonExamples = []struct {
	input interface{}
	want  interface{}
}{
	{nil, nil},
	{1., 1.},
	{3.14, 3.14},
	{"hello", "hello"},
	{true, true},
	{false, false},
	{
		map[string]interface{}{"a": "hello", "b": "world"},
		map[string]interface{}{"a": "hello", "b": "world"},
	},
	{
		map[string]interface{}{"$date": 1592084328413.},
		time.Unix(1592084328, 413000000),
	},
	{
		map[string]interface{}{"$date": nil},
		time.Unix(0, 0),
	},
	{
		map[string]interface{}{"$escape": nil},
		nil,
	},
	{
		map[string]interface{}{"$escape": 5},
		5,
	},
	{
		map[string]interface{}{"$binary": "SGVsbG8sIHdvcmxkIQ=="},
		[]byte("Hello, world!"),
	},
	{
		map[string]interface{}{"$binary": nil},
		[]byte{},
	},
	{
		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$binary": "SGVsbG8sIHdvcmxkIQ==",
			},
		},
		map[string]interface{}{"$binary": "SGVsbG8sIHdvcmxkIQ=="},
	},
	{
		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$binary": "SGVsbG8sIHdvcmxkIQ==",
			},
			"$date": 23555,
		},
		map[string]interface{}{
			"$escape": []byte("Hello, world!"),
			"$date":   23555,
		},
	},
	{
		map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"$date": 592084328513.,
				},
			},
		},
		map[string]interface{}{
			"a": map[string]interface{}{
				"b": time.Unix(592084328, 513000000),
			},
		},
	},
	{
		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$date": 10000.,
			},
		},
		map[string]interface{}{
			"$date": 10000.,
		},
	},
	{
		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$date": map[string]interface{}{
					"$date": 1592086425398.,
				},
				"a": 123,
				"b": nil,
			},
		},
		map[string]interface{}{
			"$date": time.Unix(1592086425, 398000000),
			"a": 123,
			"b": nil,
		},
	},
	{

		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$date": map[string]interface{}{
					"$date": 1592086425398,
					"$binary": "SGVsbG8sIHdvcmxkIQ==",
				},
			},
		},
		map[string]interface{}{
			"$date": map[string]interface{}{
				"$date": 1592086425398,
				"$binary": "SGVsbG8sIHdvcmxkIQ==",
			},
		},
	},
	{
		map[string]interface{}{
			"$escape": map[string]interface{}{
				"$date": map[string]interface{}{
					"$date": 32491.,
				},
			},
		},
		map[string]interface{}{
			"$date": time.Unix(32, 491000000),
		},
	},
	{
		[]interface{}{
			nil,
			1.,
			3.14,
			true,
			false,
			map[string]interface{}{
				"$escape": map[string]interface{}{
					"$date": map[string]interface{}{
						"$date": 32491.,
					},
				},
			},
			map[string]interface{}{"$date": 1592084328413.},
		},
		[]interface{}{
			nil,
			1.,
			3.14,
			true,
			false,
			map[string]interface{}{"$date": time.Unix(32, 491000000)},
			time.Unix(1592084328, 413000000),
		},
	},
	{
		map[string]interface{}{
			"$escape": []interface{}{
				map[string]interface{}{
					"$escape": map[string]interface{}{
						"$date": map[string]interface{}{
							"$date": 32491.,
						},
					},
				},
				map[string]interface{}{"$date": 1592084328413.},
			},
		},
		[]interface{}{
			map[string]interface{}{"$date": time.Unix(32, 491000000)},
			time.Unix(1592084328, 413000000),
		},
	},
}

var badEjsonExamples = []interface{}{
	map[string]interface{}{"$date": "hello"},
	map[string]interface{}{"$binary": 5},
	map[string]interface{}{"$binary": "hello world"},
	map[string]interface{}{
		"a": map[string]interface{}{
			"$date": []int{1, 2, 3},
		},
	},
}

func TestToEjson(t *testing.T) {
	for _, ex := range ejsonExamples {
		got, err := ToEjson(ex.input)
		if err != nil {
			t.Errorf("%v: %v", ex.input, err)
			continue
		}
		if !reflect.DeepEqual(got, ex.want) {
			t.Errorf("got %v, want %v", got, ex.want)
		}
	}

	for _, ex := range badEjsonExamples {
		_, err := ToEjson(ex)
		if err == nil {
			t.Errorf("expected err for %v", ex)
			continue
		}
	}
}

var ejsonBenchExample = []interface{}{
	nil,
	1.,
	3.14,
	true,
	false,
	"hello",
	map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "world",
			},
		},
	},
	map[string]interface{}{
		"x": map[string]interface{}{
			"y": map[string]interface{}{
				"z": map[string]interface{}{
					"$date": 1592086425398.,
				},
				"a": map[string]interface{}{
					"$binary": "SGVsbG8sIHdvcmxkIQ==",
				},
			},
		},
	},
	map[string]interface{}{
		"$escape": map[string]interface{}{
			"$date": map[string]interface{}{
				"$date": 1592086425398.,
			},
			"a": 123,
			"b": nil,
		},
	},
}

func BenchmarkToEjson(b *testing.B) {
	_, err := ToEjson(ejsonBenchExample)
	if err != nil {
		b.Errorf("%v: %v", ejsonBenchExample, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToEjson(ejsonBenchExample)
	}
}
