/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	CfgFile string
	Debug   bool
)

// I'm declaring as vars so I can test easier, I recommend declaring these as constants
var (
	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to STING_NUMBER.
	envPrefix = "STING"

	// Replace hyphenated flag names with camelCase in the config file
	replaceHyphenWithCamelCase = false
	// Store the result of binding cobra flags and viper config. In a
	// real application these would be data structures, most likely
	// custom structs per command. This is simplified for the demo app and is
	// not recommended that you use one-off variables. The point is that we
	// aren't retrieving the values directly from viper or flags, we read the values
	// from standard Go data structures.
	color  = ""
	number = 0
)

// Build the cobra command that handles our command line tool.
var rootCmd = &cobra.Command{

	Use:   "confluence-dump",
	Short: "Download the entirety of a Confluence workspace",
	Long: `

Have you ever wanted to use local tools, like fuzzy-search, on a Confluence web workspace?  Wish no
more, this tool will scrape all of a given Confluence space to a set of local Markdown files.

`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
		return initializeConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Working with OutOrStdout/OutOrStderr allows us to unit test our command easier
		out := cmd.OutOrStdout()

		// Print the final resolved value from binding cobra flags and viper config
		fmt.Fprintln(out, "Config:", CfgFile)
		fmt.Fprintln(out, "Your favorite color is:", color)
		fmt.Fprintln(out, "The magic number is:", number)
	},
}

func init() {
	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.Flags().IntVarP(&number, "number", "n", 7, "What is the magic number?")
	rootCmd.Flags().StringVarP(&color, "favorite-color", "c", "red", "Should come from flag first, then env var STING_FAVORITE_COLOR then the config file, then the default last")
	rootCmd.Flags().StringVar(&CfgFile, "config", "", "Should come from flag first, then env var STING_CONFIG, then the default last")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	if CfgFile != "" {
		fmt.Printf("CfgFile has value! %s\n", CfgFile)
		// Use config file from the flag.
		v.SetConfigFile(CfgFile)
		// v.SetConfigFile("/Users/pauldavid/src/toothbrush/confluence-dump/confluence-dump.yaml")
	} else {
		fmt.Printf("CfgFile is empty - searching\n")
		// Search config in home directory with name ".cobra" (without extension).
		v.AddConfigPath("$HOME/.config/")
		v.SetConfigType("yaml")
		v.SetConfigName("confluence-dump")
	}

	fmt.Println("Trying config file:", v.ConfigFileUsed())

	// Attempt to read the config file

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Printf("%v\n", err)
			log.Fatal("Couldn't read config file.  Use --config to specify, or place in ~/.config/confluence-dump.yaml")
		} else {
			// Config file was found but another error was produced
			log.Fatal(fmt.Errorf("Error loading config file: %w", err))
		}
	}
	fmt.Println("Using config file:", v.ConfigFileUsed())

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				// hmm
				panic(err)
			}
		}
	})

	return nil
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
