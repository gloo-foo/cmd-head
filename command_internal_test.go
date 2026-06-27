package command

import (
	"slices"
	"testing"
)

// TestSplitLines verifies the line-restoration contract: an empty buffer yields
// no lines, a terminating newline is dropped (not turned into a trailing empty
// line), and a buffer truncated mid-line keeps the partial final line.
func TestSplitLines(t *testing.T) {
	cases := []struct {
		name string
		buf  string
		want [][]byte
	}{
		{"empty", "", nil},
		{"trailing newline dropped", "short\n", [][]byte{[]byte("short")}},
		{"partial final line kept", "ab\nc", [][]byte{[]byte("ab"), []byte("c")}},
		{"interior empty line", "\ncd\n", [][]byte{[]byte(""), []byte("cd")}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := splitLines([]byte(c.buf)); !slices.EqualFunc(got, c.want, bytesEqual) {
				t.Fatalf("splitLines(%q) = %q, want %q", c.buf, got, c.want)
			}
		})
	}
}

// bytesEqual reports whether two byte slices hold the same bytes.
func bytesEqual(a, b []byte) bool { return slices.Equal(a, b) }

// TestSendAll_StopsWhenSendRejects verifies sendAll honors the send-false
// signal (the SIGPIPE analogue byte mode relies on): once the consumer stops
// accepting, no further lines are pushed.
func TestSendAll_StopsWhenSendRejects(t *testing.T) {
	lines := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	var got [][]byte
	send := func(line []byte) bool {
		got = append(got, line)
		return len(got) < 2 // reject once two lines have been accepted
	}

	sendAll(send, lines)

	if len(got) != 2 {
		t.Fatalf("sendAll pushed %d lines, want 2 (stop after rejection)", len(got))
	}
}
