package sse

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestSendWithBom(t *testing.T) {
	var out bytes.Buffer
	n, err := SendWithBom(&out, Id("123"), Data("hello, world"))
	require.Nil(t, err)
	const want = "\ufeffid:123\ndata:hello, world\n\n"
	assert.Equal(t, len(want), n)
	assert.Equal(t, want, out.String())
}

var sendTests = []struct {
	events []Event
	want   string
}{
	{
		want: "\n",
	},
	{
		events: []Event{Comment("abc  ")},
		want:   ":abc  \n\n",
	},
	{
		events: []Event{Comment("abc\rcba"), Comment("def\r fed")},
		want:   ":abc\n:cba\n:def\n: fed\n\n",
	},
	{
		events: []Event{Id("123"), EventType("abc\r\n cba")},
		want:   "id:123\nevent:abc\nevent:  cba\n\n",
	},
	{
		events: []Event{Data(" a,b,c"), Comment(" Hello,  world!")},
		want:   "data:  a,b,c\n: Hello,  world!\n\n",
	},
	{
		events: []Event{
			Retry(999),
			EventType("abc\r\ncba"),
			Id("  123"),
			Data("hello"),
		},
		want: "retry:999\nevent:abc\nevent:cba\nid:   123\ndata:hello\n\n",
	},
	{
		events: []Event{Retry(0), Data("a,b\r\n,c")},
		want:   "retry:0\ndata:a,b\ndata:,c\n\n",
	},
}

func TestSend(t *testing.T) {
	for i, st := range sendTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var out bytes.Buffer
			n, err := Send(&out, st.events...)
			require.Nil(t, err)
			assert.Equal(t, len(st.want), n)
			assert.Equal(t, st.want, out.String())
		})
	}
}
