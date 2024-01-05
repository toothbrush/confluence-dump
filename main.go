/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package main

import (
	"fmt"

	"github.com/toothbrush/confluence-dump/cmd"
)

func main() {

	fmt.Printf("config: %s = %v\n", "debug", cmd.Debug)
	cmd.Execute()
	fmt.Printf("? working? config: %s = %v\n", "debug", cmd.Debug)
}
