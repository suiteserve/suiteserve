// Package sse implements functions to send events to a Server-Sent Events (SSE)
// stream. The format conforms to the ABNF specification here:
// https://www.w3.org/TR/eventsource/#parsing-an-event-stream.
package sse

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Event func(io.Writer) (int, error)

// WithComment returns an Event that writes comment events to the stream.
func WithComment(s string) Event {
	return func(w io.Writer) (int, error) {
		return forLines(s, func(line string) (int, error) {
			return w.Write([]byte(":" + line + "\n"))
		})
	}
}

// WithEventType returns an Event that writes event type events to the stream.
func WithEventType(s string) Event {
	return func(w io.Writer) (int, error) {
		return sendFields(w, "event", s)
	}
}

// WithData returns an Event that writes data events to the stream.
func WithData(s string) Event {
	return func(w io.Writer) (int, error) {
		return sendFields(w, "data", s)
	}
}

// WithId returns an Event that writes ID events to the stream.
func WithId(s string) Event {
	return func(w io.Writer) (int, error) {
		return sendFields(w, "id", s)
	}
}

// WithRetry returns an Event that writes retry events to the stream.
func WithRetry(i uint64) Event {
	return func(w io.Writer) (int, error) {
		return sendFields(w, "retry", strconv.FormatUint(i, 10))
	}
}

// SendWithBom sends a byte-order mark (BOM) followed by the given events, in
// order, to a writer, returning the number of bytes written and any error. A
// BOM must only appear at the start of an SSE stream, therefore this function
// should not be called after a call to Send on the same writer. It is rare to
// need to use this function at all.
//
// If the writer implements http.Flusher, the Flush function is called before
// SendWithBom returns.
func SendWithBom(w io.Writer, events ...Event) (n int, err error) {
	if n, err = w.Write([]byte("\ufeff")); err != nil {
		return n, err
	}
	n2, err := Send(w, events...)
	return n + n2, err
}

// Send sends the given events, in order, to a writer, returning the number of
// bytes written and any error.
//
// If the writer implements http.Flusher, the Flush function is called before
// Send returns.
func Send(w io.Writer, events ...Event) (n int, err error) {
	for _, event := range events {
		n2, err := event(w)
		n += n2
		if err != nil {
			return n, err
		}
	}
	n2, err := w.Write([]byte{'\n'})
	n += n2
	if err != nil {
		return n, err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return n, nil
}

func sendFields(w io.Writer, k, v string) (int, error) {
	if v == "" {
		return w.Write([]byte(k + "\n"))
	}
	return forLines(v, func(line string) (int, error) {
		if strings.HasPrefix(line, " ") {
			line = " " + line
		}
		return w.Write([]byte(k + ":" + line + "\n"))
	})
}

func forLines(s string, f func(line string) (int, error)) (int, error) {
	lines := strings.FieldsFunc(s, func(r rune) bool {
		return r == '\r' || r == '\n'
	})
	var n int
	for _, line := range lines {
		n2, err := f(line)
		n += n2
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
