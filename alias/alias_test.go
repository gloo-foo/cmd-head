package alias_test

import (
	"slices"
	"testing"

	"github.com/gloo-foo/testable"

	head "github.com/gloo-foo/cmd-head/alias"
)

// The alias package re-exports the constructor and flag types under unprefixed
// names. A mis-wired re-export (say, Bytes aliased to HeadLines, or Head bound
// to the wrong function) compiles cleanly, so only behavior can prove the
// wiring. Each test exercises one re-export and asserts the GNU head output it
// must produce.

const input = "alpha\nbeta\ngamma\n"

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAlias_DefaultEmitsFirstTenLines(t *testing.T) {
	// With no flag, Head emits the leading ten lines; this input has three.
	lines, err := testable.TestLines(head.Head(), input)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"alpha", "beta", "gamma"})
}

func TestAlias_LinesLimitsToCount(t *testing.T) {
	// Lines(2) must emit exactly the first two lines.
	lines, err := testable.TestLines(head.Head(head.Lines(2)), input)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"alpha", "beta"})
}

func TestAlias_BytesLimitsToByteCount(t *testing.T) {
	// Bytes(7) joins lines with newlines ("alpha\nbeta\n…") and keeps the first
	// seven bytes: "alpha\nb". TestLines splits that on the newline.
	lines, err := testable.TestLines(head.Head(head.Bytes(7)), input)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"alpha", "b"})
}
