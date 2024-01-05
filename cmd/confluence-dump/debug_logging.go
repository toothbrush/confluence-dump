package main

import (
	"fmt"
	"os"
)

func debugLog(format string, a ...any) {
	if Debug {
		string := fmt.Sprintf(format, a...)
		fmt.Fprintf(os.Stderr, "[confluence-dump] %s", string)
	}
}
