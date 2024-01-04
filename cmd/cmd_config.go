/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Commands to work with the app config",
	Long: `

Commands in this namespace are to help you configure the app.  Find out what the current config is,
or learn where it's being read from.

`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
