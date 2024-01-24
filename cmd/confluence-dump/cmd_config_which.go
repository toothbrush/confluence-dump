/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var whichUsage = strings.TrimSpace(`
Output your resolved config filename.

You can influence this with the --config flag, or by exporting the CONFLUENCE_DUMP_CONFIG variable.
`)

// whichCmd represents the which command
var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Tell me the resolved config path",
	Long:  whichUsage,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Config path: %s\n", Config)
	},
}

func init() {
	configCmd.AddCommand(whichCmd)
}
