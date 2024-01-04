/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Debug   bool
)

var rootCmd = &cobra.Command{Use: "confluence-dump",
	Short: "Download the entirety of a Confluence workspace",
	Long: `

Have you ever wanted to use local tools, like fuzzy-search, on a Confluence web workspace?  Wish no
more, this tool will scrape all of a given Confluence space to a set of local Markdown files.

`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.confluence-dump.yaml)")

	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Display debugging output in the console. (default: false)")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

}
