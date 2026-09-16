package command_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
	"github.com/gloo-foo/testable/assertion"
	"github.com/gloo-foo/testable/run"

	command "github.com/gloo-foo/cmd-head"
)

// errUpstream is a sentinel emitted by a deliberately failing source so byte
// mode's error-propagation path is observable via errors.Is.
var errUpstream = errors.New("upstream failed")

// countingSource adapts a Stream factory to gloo.Source[[]byte] for tests.
type countingSource func(context.Context) gloo.Stream[[]byte]

func (f countingSource) Stream(ctx context.Context) gloo.Stream[[]byte] { return f(ctx) }

// TestHead_StopsUnboundedUpstream is the regression that justifies using
// Head/Take over a StatefulFilter counter: line-mode head must stop the
// upstream after N lines (the SIGPIPE analogue), not drain the whole source. A
// counter-based filter would pull every one of these million lines; Take stops
// the producer, so only a small prefix is ever generated.
func TestHead_StopsUnboundedUpstream(t *testing.T) {
	const huge = 1_000_000
	var produced atomic.Int64
	src := countingSource(func(ctx context.Context) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, _ func(error)) {
			for i := range huge {
				produced.Add(1)
				if !send(fmt.Appendf(nil, "line %d", i)) {
					return
				}
			}
		})
	})

	out, err := gloo.From(context.Background(), src, command.Head(command.HeadLines(3))).Collect()
	assertion.NoError(t, err)
	if len(out) != 3 {
		t.Fatalf("got %d lines, want 3", len(out))
	}
	if string(out[0]) != "line 0" || string(out[2]) != "line 2" {
		t.Errorf("got %q…%q, want line 0…line 2", out[0], out[2])
	}
	// With early-stop, the producer emits only ~3 + one stream buffer of items
	// before it sees the stop. Without it, all `huge` would be produced.
	if n := produced.Load(); n > 10_000 {
		t.Errorf("head did not stop the upstream: producer emitted %d items (want a small prefix)", n)
	}
}

// ==============================================================================
// Default Behavior (10 lines)
// ==============================================================================

func TestHead_DefaultTenLines(t *testing.T) {
	input := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n"
	lines, err := testable.TestLines(command.Head(), run.Input(input))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
}

func TestHead_LessThanDefault(t *testing.T) {
	lines, err := testable.TestLines(command.Head(), "1\n2\n3\n4\n5\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3", "4", "5"})
}

func TestHead_ExactlyTenLines(t *testing.T) {
	input := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"
	lines, err := testable.TestLines(command.Head(), run.Input(input))
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
}

func TestHead_EmptyInput(t *testing.T) {
	lines, err := testable.TestLines(command.Head(), "")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

// ==============================================================================
// Custom Line Counts
// ==============================================================================

func TestHead_CustomThreeLines(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadLines(3)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c"})
}

func TestHead_CustomOneLine(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadLines(1)), "first\nsecond\nthird\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"first"})
}

func TestHead_FewerLinesThanN(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadLines(100)), "1\n2\n3\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1", "2", "3"})
}

func TestHead_ExactlyN(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadLines(5)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c", "d", "e"})
}

// ==============================================================================
// Table-Driven
// ==============================================================================

func TestHead_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		n        command.HeadLines
	}{
		{"three from five", "a\nb\nc\nd\ne\n", []string{"a", "b", "c"}, 3},
		{"one line", "first\nsecond\nthird\n", []string{"first"}, 1},
		{"all lines", "a\nb\n", []string{"a", "b"}, 5},
		{"with empty lines", "a\n\nb\nc\n", []string{"a", "", "b"}, 3},
		{"unicode", "hello\nworld\nend\n", []string{"hello", "world"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := testable.TestLines(command.Head(tt.n), run.Input(tt.input))
			assertion.NoError(t, err)
			assertion.Lines(t, lines, tt.expected)
		})
	}
}

// ==============================================================================
// Byte Count (-c)
// ==============================================================================

func TestHead_Bytes(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadBytes(5)), "hello world\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"hello"})
}

func TestHead_Bytes_MoreThanInput(t *testing.T) {
	// Input "short\n" → stream ["short"] → reconstituted "short\n" (6 bytes)
	// N=100 exceeds length, so all bytes emitted: "short\n"
	// TestLines sees "short\n" + "\n" → trims → splits → ["short"]
	lines, err := testable.TestLines(command.Head(command.HeadBytes(100)), "short\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"short"})
}

func TestHead_Bytes_EmptyInput(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadBytes(5)), "")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

func TestHead_Bytes_MultipleLines(t *testing.T) {
	// Input "ab\ncd\n" → stream ["ab", "cd"] → reconstituted "ab\ncd\n" (6 bytes)
	// First 4 bytes: "ab\nc" → TestLines splits on \n → ["ab", "c"]
	lines, err := testable.TestLines(command.Head(command.HeadBytes(4)), "ab\ncd\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"ab", "c"})
}

// TestHead_Bytes_PropagatesUpstreamError covers byte mode's error path: when
// the upstream stream fails, draining it surfaces the error, which Head must
// forward downstream rather than swallow.
func TestHead_Bytes_PropagatesUpstreamError(t *testing.T) {
	src := countingSource(func(ctx context.Context) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, sendErr func(error)) {
			if send([]byte("partial")) {
				sendErr(errUpstream)
			}
		})
	})

	_, err := gloo.From(context.Background(), src, command.Head(command.HeadBytes(3))).Collect()
	if !errors.Is(err, errUpstream) {
		t.Fatalf("got err %v, want %v", err, errUpstream)
	}
}

// ==============================================================================
// Edge Cases
// ==============================================================================

func TestHead_ManyLines(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 1000; i++ {
		// strings.Builder.Write never returns an error (documented contract); the
		// blank assignment acknowledges the error return the linter sees.
		_, _ = fmt.Fprintf(&b, "line %d\n", i)
	}
	lines, err := testable.TestLines(command.Head(command.HeadLines(10)), run.Input(b.String()))
	assertion.NoError(t, err)
	assertion.Count(t, lines, 10)
	assertion.Equal(t, lines[0], "line 1", "first line")
	assertion.Equal(t, lines[9], "line 10", "tenth line")
}

func TestHead_EmptyLines(t *testing.T) {
	lines, err := testable.TestLines(command.Head(command.HeadLines(3)), "a\nb\nc\nd\ne\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c"})
}

func TestHead_UnknownOptionPassesThrough(t *testing.T) {
	// An argument that is not one of head's option types is handed to the
	// framework's positional classifier and must not disturb line selection.
	lines, err := testable.TestLines(command.Head(command.HeadLines(1), 42), "a\nb\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 || lines[0] != "a" {
		t.Fatalf("expected [a], got %v", lines)
	}
}
