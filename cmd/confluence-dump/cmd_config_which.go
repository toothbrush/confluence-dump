/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// whichCmd represents the which command
var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Tell me the resolved config path",
	Long: `
Output the filename that's being used to store your config.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Config path: %s\n", ConfigActual)
	},
}

func init() {
	configCmd.AddCommand(whichCmd)
}
