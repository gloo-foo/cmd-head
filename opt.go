package command

// HeadLines sets the number of lines to output. Default: 10.
type HeadLines int

// HeadBytes sets the number of bytes to output (-c flag).
// When set, the command collects all input and emits the first N bytes.
type HeadBytes int

// flags is the folded option set for a head run. The zero value selects the
// default behavior: the first ten lines.
type flags struct {
	lines HeadLines
	bytes HeadBytes
}

// fold partitions opts: head's own option values are folded into the flag set,
// and every other argument is passed through unchanged for the framework's
// positional classifier.
func fold(opts []any) (flags, []any) {
	var f flags
	rest := make([]any, 0, len(opts))
	for _, o := range opts {
		switch v := o.(type) {
		case HeadLines:
			f.lines = v
		case HeadBytes:
			f.bytes = v
		default:
			rest = append(rest, o)
		}
	}
	return f, rest
}
