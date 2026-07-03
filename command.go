package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

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
	f, rest := fold(opts)
	gloo.NewParameters[gloo.File, struct{}](rest...)
	if f.bytes > 0 {
		return byteCommand(f.bytes)
	}
	return patterns.Head[[]byte](resolveLines(f.lines))
}

// resolveLines applies the GNU default when no positive line count is given.
func resolveLines(n HeadLines) patterns.Count {
	if n <= 0 {
		return defaultLines
	}
	return patterns.Count(n)
}
