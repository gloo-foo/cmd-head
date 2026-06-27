package head_test

import (
	"fmt"

	command "github.com/gloo-foo/cmd-head"

	"github.com/gloo-foo/testable"
)

// This example demonstrates head with default line count (10 lines).
func ExampleHead_fromFile_basic() {
	// head testdata/numbers.txt
	output, _ := testable.Test(
		command.Head(),
		"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12",
	)
	fmt.Print(output)
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 10
}
