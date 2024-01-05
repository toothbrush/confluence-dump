package cmd

import (
	"fmt"
	"os"
)

func d_log(format string, a ...any) {
	if Debug {
		string := fmt.Sprintf(format, a...)
		fmt.Fprintf(os.Stderr, "[confluence-dump] %s", string)
	}
}
