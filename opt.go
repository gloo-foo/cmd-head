package command

// HeadLines sets the number of lines to output. Default: 10.
type HeadLines int

// HeadBytes sets the number of bytes to output (-c flag).
// When set, the command collects all input and emits the first N bytes.
type HeadBytes int

type flags struct {
	lines HeadLines
	bytes HeadBytes
}

func (l HeadLines) Configure(f *flags) { f.lines = l }
func (b HeadBytes) Configure(f *flags) { f.bytes = b }
