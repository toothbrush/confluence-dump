/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toothbrush/confluence-dump/confluence"
)

var listSpacesUsage = strings.TrimSpace(`
If you want to find out what spaces your Confluence wiki has, use this command.
`)

var listSpacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Print list of spaces",
	Long:  listSpacesUsage,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		tokenCmdOutput, err := exec.Command(AuthTokenCmd[0], AuthTokenCmd[1:]...).Output()
		if err != nil {
			return fmt.Errorf("download: couldn't execute auth-token-cmd '%v': %w", AuthTokenCmd, err)
		}

		token := strings.Split(string(tokenCmdOutput), "\n")[0]
		api, err := confluence.NewAPI(
			ConfluenceInstance,
			AuthUsername,
			token)
		if err != nil {
			return fmt.Errorf("download: couldn't instantiate Confluence API: %w", err)
		}

		// list all spaces:
		log.Printf("Listing Confluence spaces in %s...\n", ConfluenceInstance)
		spacesRemote, err := api.ListAllSpaces(ctx, ConfluenceInstance, IncludePersonal)
		if err != nil {
			return fmt.Errorf("download: couldn't list Confluence spaces: %w", err)
		}

		spacesRemote["blogposts"] = confluence.Space{
			ID:   "blogposts",
			Key:  "blogposts",
			Name: "Users' blogposts",
			Org:  ConfluenceInstance,
		}

		log.Printf("Found %d spaces on '%s'.\n", len(spacesRemote), ConfluenceInstance)

		spaceKeys := []string{}

		for _, space := range spacesRemote {
			spaceKeys = append(spaceKeys, space.Key)
		}

		sort.Strings(spaceKeys)

		fmt.Printf("spaces:\n")
		for _, spaceKey := range spaceKeys {
			if s, ok := spacesRemote[spaceKey]; ok {
				fmt.Printf("  - %s: %s\n", spaceKey, s.Name)
			}
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listSpacesCmd)

	listSpacesCmd.Flags().BoolVar(&IncludePersonal, "include-personal-spaces", false, "list individuals' personal spaces")
}
