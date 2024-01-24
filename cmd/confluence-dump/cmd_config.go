/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"strings"

	"github.com/spf13/cobra"
)

var configUsage = strings.TrimSpace(`
Commands in this namespace are to help you configure the app.  Find out what the current config is,
or learn where it's being read from.
`)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Commands to work with the app config",
	Long:  configUsage,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
