package head_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-head"
)

func ExampleHead_lines() {
	// echo "1\n2\n3\n4\n5" | head -n 3
	output, _ := testable.Test(
		command.Head(command.HeadLines(3)),
		"1\n2\n3\n4\n5",
	)
	fmt.Print(output)
	// Output:
	// 1
	// 2
	// 3
}
