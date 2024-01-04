/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Output current config",
	Long: `

Is something not working for you?  Have a look whether your config is as you expect.

`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Dump current config state:\n")
		for key, value := range viper.GetViper().AllSettings() {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Printf(" %s: %s\n", "config", viper.GetString("config"))
	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
