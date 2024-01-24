/*
Copyright © 2024 paul <paul@denknerd.org>
*/

package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/fatih/structs"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	// Store the result of binding cobra flags
	Config string
	Debug  bool

	// Command to run to retrieve API Personal Access Token
	AuthTokenCmd []string

	AuthUsername       string
	LocalStore         string
	ConfluenceInstance string

	ParsedConfig YamlConfig
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
			return fmt.Errorf("confluence-dump: failed to initialise config: %w", err)
		}

		if len(AuthTokenCmd) < 1 {
			return fmt.Errorf("confluence-dump: please provide --auth-token-cmd")
		}
		return nil
	},
}

func init() {
	// Define cobra flags, the default value has the lowest (least significant) precedence
	rootCmd.PersistentFlags().StringVar(&Config, "config", "", "config file location (default: ~/.config/confluence-dump.yaml, respects CONFLUENCE_DUMP_CONFIG)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "display debug output")
	rootCmd.PersistentFlags().StringSliceVar(&AuthTokenCmd, "auth-token-cmd", []string{}, "shell command to retrieve Atlassian auth token")
	rootCmd.PersistentFlags().StringVar(&LocalStore, "store", "", "location to save Confluence pages")
	rootCmd.PersistentFlags().StringVar(&AuthUsername, "auth-username", "", "your Atlassian username")
	rootCmd.PersistentFlags().StringVar(&ConfluenceInstance, "confluence-instance", "", "your Atlassian ORG name, e.g. ORG in ORG.atlassian.net")
}

func initializeConfig(cmd *cobra.Command) error {
	if Config == "" {
		// Did the user provide an ENV?
		envConfig := os.Getenv("CONFLUENCE_DUMP_CONFIG")
		if envConfig != "" {
			Config = envConfig
		} else {
			// As fallback, search for config in home XDG-ish directory
			Config = "~/.config/confluence-dump.yaml"
		}
	}
	config, err := homedir.Expand(Config)
	if err != nil {
		return fmt.Errorf("confluence-dump: unable to expand homedir: %w", err)
	}
	Config = config

	// Use config file from the flag.
	if _, err := os.Stat(Config); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Couldn't read config file %s, does it exist?  Override with --config.\n", Config)
		return fmt.Errorf("confluence-dump: specified config file does not exist: %w", err)
	}

	yamlFile, err := os.ReadFile(Config)
	if err != nil {
		return fmt.Errorf("confluence-dump: error reading config file: %w", err)
	}

	// I'd like to bark if a user sets a flag we don't recognise:
	if err := yaml.UnmarshalStrict(yamlFile, &ParsedConfig); err != nil {
		return fmt.Errorf("confluence-dump: issue parsing config file: %w", err)
	}

	// Bind the current command's flags to viper
	if err := bindFlags(cmd, ParsedConfig); err != nil {
		return fmt.Errorf("confluence-dump: failed to bind flags: %w", err)
	}

	return nil
}

type YamlConfig struct {
	AlwaysDownload   *bool `yaml:"always-download"`
	WithVCR          *bool `yaml:"with-vcr"`
	AllSpaces        *bool `yaml:"all-spaces"`
	IncludeArchived  *bool `yaml:"include-archived"`
	IncludeBlogposts *bool `yaml:"include-blogposts"`
	WriteMarkdown    *bool `yaml:"write-markdown"`
	Prune            *bool `yaml:"prune"`

	StorePath          string   `yaml:"store"`
	ConfluenceInstance string   `yaml:"confluence-instance"`
	AuthUsername       string   `yaml:"auth-username"`
	AuthTokenCmd       []string `yaml:"auth-token-cmd"`
	Spaces             []string `yaml:"spaces"`

	PostDownloadCmd []string `yaml:"post-download-cmd"`
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v YamlConfig) error {
	for _, field := range structs.Fields(v) {
		key := field.Tag("yaml")
		if key == "" {
			return fmt.Errorf("confluence-dump: could not retrieve struct tag 'yaml'")
		}
		if flag := cmd.Flag(key); flag == nil {
			// hmm... the flag is unknown.  but that can legitimately happen if you're running
			// e.g. `list spaces` which has no `include-archived` flag but your YAML file does
			// define that flag...
			continue
		}
		if !cmd.Flags().Changed(key) {
			switch field.Kind() {
			case reflect.Ptr:
				// err, this is crappy, but i know YamlConfig only uses pointers for bools.....
				b, ok := field.Value().(*bool)
				if !ok {
					return fmt.Errorf("confluence-dump: found unrecognised field: %+v", field)
				}
				if b != nil {
					cmd.Flags().Set(key, fmt.Sprintf("%v", *b))
				}

			case reflect.String:
				s, ok := field.Value().(string)
				if !ok {
					return fmt.Errorf("confluence-dump: found unrecognised field: %+v", field)
				}
				if s != "" {
					cmd.Flags().Set(key, s)
				}

			case reflect.Slice:
				ss, ok := field.Value().([]string)
				if !ok {
					return fmt.Errorf("confluence-dump: found unrecognised field: %+v", field)
				}
				for _, s := range ss {
					// yes, repeatedly calling Set() appends to the slice...
					cmd.Flags().Set(key, s)
				}

			default:
				return fmt.Errorf("confluence-dump: found unrecognised field: %+v", field)
			}
		}
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Flags are only available after (or inside, presumably) the .Execute() thing.
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("confluence-dump: execution error: %w", err)
	}

	return nil
}
