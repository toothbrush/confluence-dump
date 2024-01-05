/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Output current config",
	Long: `
Is something not working for you?  Have a look whether your config is as you expect.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Note, you can only talk about persistent flags here.  Command-specific ones won't be
		// visible.
		fmt.Printf("Dump current config state:\n\n")

		fmt.Printf("  Config: %s\n", Config)
		fmt.Printf("  ConfigActual: %s\n", ConfigActual)
		fmt.Printf("  Debug: %v\n", Debug)
		fmt.Println()
		fmt.Printf("  AuthTokenCmd: %v\n", AuthTokenCmd)
		fmt.Printf("  LocalStore: %v\n", LocalStore)
	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
