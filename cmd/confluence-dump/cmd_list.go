/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Commands to list items",
	Long: `
Commands in this namespace are to help you explore the Confluence wiki.
`,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
