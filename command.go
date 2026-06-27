package command

import (
	"bytes"
	"context"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// newline is the line separator used to reconstitute and re-split byte-mode output.
var newline = []byte{'\n'}

// defaultLines is the GNU head default: the first ten lines.
const defaultLines = 10

// Head returns a Command that outputs the leading prefix of its input.
//
// Flags:
//   - HeadLines (-n): emit the first N lines (default: 10).
//   - HeadBytes (-c): emit the first N bytes; collects the input, joins lines
//     with newlines, truncates to N bytes, and emits the result as a
//     line-oriented stream (one value per line, the newlines stripped).
//
// HeadBytes takes precedence over HeadLines when set to a positive count.
func Head(opts ...any) gloo.Command[[]byte, []byte] {
	f := gloo.NewParameters[gloo.File, flags](opts...).Flags
	if f.bytes > 0 {
		return byteCommand(int(f.bytes))
	}
	return patterns.Head[[]byte](resolveLines(f.lines))
}

// resolveLines applies the GNU default when no positive line count is given.
func resolveLines(n HeadLines) int {
	if n <= 0 {
		return defaultLines
	}
	return int(n)
}

// byteCommand builds a Command that emits the leading n bytes of the input.
//
// Line mode uses patterns.Head/Take so an unbounded upstream is stopped early;
// byte mode cannot, because the byte length of a line is unknown until it is
// read, so the whole input must be reconstituted before truncation.
func byteCommand(n int) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.GenerateFrom(ctx, in, emitBytes(in, n))
	})
}

// emitBytes returns the producer that drains in, joins it, truncates to the
// leading n bytes, and re-emits the result as a line-oriented stream.
func emitBytes(in gloo.Stream[[]byte], n int) func(context.Context, func([]byte) bool, func(error)) {
	return func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
		items, err := rill.ToSlice(in.Chan())
		if err != nil {
			sendErr(err)
			return
		}
		sendAll(send, splitLines(truncate(join(items), n)))
	}
}

// sendAll emits each line downstream, stopping early if send signals no more.
func sendAll(send func([]byte) bool, lines [][]byte) {
	for _, line := range lines {
		if !send(line) {
			return
		}
	}
}

// splitLines restores the line-oriented stream from the truncated buffer,
// dropping the empty trailer left by a terminating newline so each value is one
// line with its newline stripped. An empty buffer yields no lines.
func splitLines(buf []byte) [][]byte {
	if len(buf) == 0 {
		return nil
	}
	lines := bytes.Split(buf, newline)
	if last := len(lines) - 1; len(lines[last]) == 0 {
		return lines[:last]
	}
	return lines
}

// join concatenates the line-oriented stream back into a byte buffer, restoring
// the newline terminator the line splitter stripped from each item.
func join(items [][]byte) []byte {
	var buf []byte
	for _, item := range items {
		buf = append(buf, item...)
		buf = append(buf, '\n')
	}
	return buf
}

// truncate returns the leading n bytes of buf (all of buf when it is shorter).
func truncate(buf []byte, n int) []byte {
	if n < len(buf) {
		return buf[:n]
	}
	return buf
}
