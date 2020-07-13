package httpheader

import (
	"reflect"
	"testing"
)

var connectionTokensTests = []struct {
	input []string
	want  []string
}{
	{
		want: []string{},
	},
	{
		input: []string{},
		want:  []string{},
	},
	{
		input: []string{"foo", "bar"},
		want:  []string{"foo", "bar"},
	},
	{
		input: []string{"foo, bar", "baz"},
		want:  []string{"foo", "bar", "baz"},
	},
	{
		input: []string{" hello,, \r\n\t,  world ,", " hi-there,  @]\"(),,"},
		want:  []string{"hello", "world", "hi-there"},
	},
	{
		input: []string{"  \r\n  \r\n  , a", " 5"},
		want:  []string{"a", "5"},
	},
	{
		input: []string{"  ,c", " 5"},
		want:  []string{"c", "5"},
	},
}

func TestConnectionTokens(t *testing.T) {
	for _, test := range connectionTokensTests {
		got := ConnectionTokens(test.input)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("want %q, got %q", test.want, got)
		}
	}
}

var upgradeTokensTests = []struct {
	input []string
	want  []string
}{
	{
		want: []string{},
	},
	{
		input: []string{},
		want:  []string{},
	},
	{
		input: []string{"foo", "bar"},
		want:  []string{"foo", "bar"},
	},
	{
		input: []string{"foo, bar", "baz"},
		want:  []string{"foo", "bar", "baz"},
	},
	{
		input: []string{" hello,, \r\n\t,  world ,", " hi-there/2.0,  @]\"(),,"},
		want:  []string{"hello", "world", "hi-there/2.0"},
	},
	{
		input: []string{"/ \r\n  \r\n  , a", " 5"},
		want:  []string{"/", "a", "5"},
	},
	{
		input: []string{" / ,c", " 5"},
		want:  []string{"/", "c", "5"},
	},
}

func TestUpgradeTokens(t *testing.T) {
	for _, test := range upgradeTokensTests {
		got := UpgradeTokens(test.input)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("want %q, got %q", test.want, got)
		}
	}
}
