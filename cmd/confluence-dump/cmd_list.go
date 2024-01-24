/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"strings"

	"github.com/spf13/cobra"
)

var listUsage = strings.TrimSpace(`
Commands in this namespace are to help you explore the Confluence wiki.
`)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Commands to list items",
	Long:  listUsage,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
