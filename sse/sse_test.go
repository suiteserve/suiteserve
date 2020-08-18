package sse_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suiteserve/suiteserve/sse"
	"io/ioutil"
	"strconv"
	"testing"
)

func TestSendWithBom(t *testing.T) {
	var out bytes.Buffer
	n, err := sse.SendWithBom(&out, sse.WithId("123"),
		sse.WithData("hello, world"))
	require.Nil(t, err)
	const want = "\ufeffid:123\ndata:hello, world\n\n"
	assert.Equal(t, len(want), n)
	assert.Equal(t, want, out.String())
}

var sendTests = []struct {
	events []sse.Event
	want   string
}{
	{
		want: "\n",
	},
	{
		events: []sse.Event{sse.WithComment("abc  ")},
		want:   ":abc  \n\n",
	},
	{
		events: []sse.Event{
			sse.WithComment("abc\rcba"),
			sse.WithComment("def\r fed"),
		},
		want: ":abc\n:cba\n:def\n: fed\n\n",
	},
	{
		events: []sse.Event{
			sse.WithId("123"),
			sse.WithEventType("abc\r\n c\n\rba"),
		},
		want: "id:123\nevent:abc\nevent:  c\nevent\nevent:ba\n\n",
	},
	{
		events: []sse.Event{
			sse.WithData(" a,b,c"),
			sse.WithComment(" Hello,  world!"),
		},
		want: "data:  a,b,c\n: Hello,  world!\n\n",
	},
	{
		events: []sse.Event{
			sse.WithRetry(1),
			sse.WithEventType("x\r\n\n\ry"),
			sse.WithId("  ab"),
			sse.WithData("hi"),
		},
		want: "retry:1\nevent:x\nevent\nevent\nevent:y\nid:   ab\ndata:hi\n\n",
	},
	{
		events: []sse.Event{sse.WithRetry(0), sse.WithData("a,b\r\n,c")},
		want:   "retry:0\ndata:a,b\ndata:,c\n\n",
	},
}

func TestSend(t *testing.T) {
	for i, st := range sendTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var out bytes.Buffer
			n, err := sse.Send(&out, st.events...)
			require.Nil(t, err)
			assert.Equal(t, len(st.want), n)
			assert.Equal(t, st.want, out.String())
		})
	}
}

var sendBenchmark = []sse.Event{
	sse.WithComment("keep-alive"),
	sse.WithId("1badb002"),
	sse.WithEventType("insert"),
	sse.WithData("Lorem\r\nipsum\ndolor\rsit\n\ramet."),
	sse.WithRetry(500),
}

func BenchmarkSend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = sse.Send(ioutil.Discard, sendBenchmark...)
	}
}

func ExampleSend() {
	var w bytes.Buffer
	_, err := sse.Send(&w, sse.WithComment("This is a\n comment."),
		sse.WithData("Hello\nthere,\r\n world\r!"),
		sse.WithId("123"))
	if err != nil {
		panic(err)
	}
	_, err = sse.Send(&w, sse.WithRetry(50), sse.WithRetry(100))
	if err != nil {
		panic(err)
	}
	fmt.Println(w.String())
	// Output:
	// :This is a
	// : comment.
	// data:Hello
	// data:there,
	// data:  world
	// data:!
	// id:123
	//
	// retry:50
	// retry:100
}
