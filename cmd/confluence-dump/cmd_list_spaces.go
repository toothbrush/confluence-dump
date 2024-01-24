/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toothbrush/confluence-dump/confluence"
)

var listSpacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Output list of spaces",
	Long: `
If you want to find out what spaces your Confluence wiki has, use this command.
`,
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
		spacesRemote, err := api.ListAllSpaces(ctx, ConfluenceInstance)
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

		fmt.Printf("spaces:\n")
		for _, space := range spacesRemote {
			fmt.Printf("  - %s: %s\n", space.Key, space.Name)
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listSpacesCmd)
}
