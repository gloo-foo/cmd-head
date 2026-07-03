package head_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-head"
)

func ExampleHead_basic() {
	// echo "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12" | head
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
