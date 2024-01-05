/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	CfgFile string
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
	fmt.Println("called root.Execute")
	fmt.Printf("config: %s = %v\n", "debug", Debug)
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	fmt.Println("called root.initConfig")

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath("$HOME/.config/")
		viper.SetConfigType("yaml")
		viper.SetConfigName("confluence-dump")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatal("Couldn't determine config file location.  Use --config to specify, or place in ~/.config/confluence-dump.yaml")
		} else {
			// Config file was found but another error was produced
			log.Fatal(fmt.Errorf("Error loading config file: %w", err))
		}
	}

	// Config file found and successfully parsed
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	viper.AutomaticEnv()

	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) {
			//rootCmd.Flags().SetAnnotation(f.Name, cobra.BashCompOneRequiredFlag, []string{"false"})
			rootCmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})

}

func init() {
	fmt.Println("called root.init")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "output debugging information")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	fmt.Printf("config: %s = %v\n", "debug", Debug)
}
