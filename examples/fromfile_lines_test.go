package head_test

import (
	"fmt"

	command "github.com/gloo-foo/cmd-head"

	"github.com/gloo-foo/testable"
)

func ExampleHead_fromFile_lines() {
	// head -n 3 testdata/numbers.txt
	output, _ := testable.Test(
		command.Head(command.HeadLines(3)),
		"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12",
	)
	fmt.Print(output)
	// Output:
	// 1
	// 2
	// 3
}
