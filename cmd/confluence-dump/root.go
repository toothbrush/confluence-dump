/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// The environment variable prefix of all environment variables bound to our command line flags.
	// For example, --number is bound to CONFLUENCE_DUMP_NUMBER.
	envPrefix = "CONFLUENCE_DUMP"

	// Replace hyphenated flag names with camelCase in the config file
	replaceHyphenWithCamelCase = false

	// Store the result of binding cobra flags and viper config.
	Config       string // this is what the user provides
	ConfigActual string // this is what Viper ended up loading
	Debug        bool

	// Command to run to retrieve API Personal Access Token
	AuthTokenCmd []string

	AuthUsername       string
	LocalStore         string
	ConfluenceInstance string
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
		// You can bind cobra and viper in a few locations, but PersistentPreRunE on the root command works well
		if err := initializeConfig(cmd); err != nil {
			return fmt.Errorf("cmd: failed to initialise config: %w", err)
		}
		return nil
	},
}

func init() {
	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.PersistentFlags().StringVar(&Config, "config", "", "config file location (default: ~/.config/confluence-dump.yaml)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "display debug output")
	rootCmd.PersistentFlags().StringSliceVar(&AuthTokenCmd, "auth-token-cmd", []string{}, "shell command to retrieve Atlassian auth token")
	rootCmd.PersistentFlags().StringVar(&LocalStore, "store", "", "location to save Confluence pages")
	rootCmd.PersistentFlags().StringVar(&AuthUsername, "auth-username", "", "your Atlassian username")
	rootCmd.PersistentFlags().StringVar(&ConfluenceInstance, "confluence-instance", "", "your Atlassian ORG name, e.g. ORG in ORG.atlassian.net")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.GetViper()

	if Config != "" {
		// Use config file from the flag.
		v.SetConfigFile(Config)
		if _, err := os.Stat(Config); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Specified config file does not exist: %s", Config)
		}
	} else {
		// Search config in home XDG-ish directory
		v.AddConfigPath("$HOME/.config/")
		v.SetConfigType("yaml")
		v.SetConfigName("confluence-dump")
	}

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

	ConfigActual = v.ConfigFileUsed()

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable CONFLUENCE_DUMP_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to CONFLUENCE_DUMP_FAVORITE_COLOR
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
			rt := reflect.TypeOf(val)

			if rt.Kind() == reflect.Slice {
				// let's handle slices specially, because i don't want nested stuff...
				//
				// cobra expects a slice to be provided with commas, i believe.
				if valslice, ok := val.([]interface{}); ok {
					for _, vs := range valslice {
						// yes, repeatedly calling Set() appends to the slice...
						err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", vs))
						if err != nil {
							// hmm
							panic(err)
						}
					}
				}
			} else {
				err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
				if err != nil {
					// hmm
					panic(err)
				}
			}
		}
	})

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Flags are only available after (or inside, presumably) the .Execute() thing.
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
