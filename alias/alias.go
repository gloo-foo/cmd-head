// Package alias provides unprefixed names for the head command flags.
//
//	import "github.com/gloo-foo/cmd-head/alias"
//	head.Head(head.Lines(5))
//	head.Head(head.Bytes(20))
package alias

import command "github.com/gloo-foo/cmd-head"

// Head re-exports the constructor.
var Head = command.Head

// Lines (-n) sets the number of leading lines to emit.
type Lines = command.HeadLines

// Bytes (-c) sets the number of leading bytes to emit.
type Bytes = command.HeadBytes
