/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var versionUsage = strings.TrimSpace(`
Show version information
`)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: versionUsage,
	Long:  versionUsage,
	RunE:  versionRun,
	Args:  cobra.ExactArgs(0),
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	// Version will be the version tag if the binary is built with "go install url/tool@version".
	// If the binary is built some other way, it will be "(devel)".  main.Version is automatically
	// set at build time.
	Version = "unknown"
	// Revision is taken from the vcs.revision tag in Go 1.18+.
	Revision = "unknown"
	// LastCommit is taken from the vcs.time tag in Go 1.18+.
	LastCommit time.Time
	// DirtyBuild is taken from the vcs.modified tag in Go 1.18+.
	DirtyBuild = true
)

func versionRun(cmd *cobra.Command, args []string) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("cmd_version: could not read build info")
	}
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			Revision = kv.Value
		case "vcs.time":
			LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			DirtyBuild = kv.Value == "true"
		}
	}

	parts := make([]string, 0, 3)
	if Version != "unknown" && Version != "(devel)" {
		parts = append(parts, Version)
	}
	if Revision != "unknown" && Revision != "" {
		parts = append(parts, "rev")
		parts = append(parts, Revision)
		if DirtyBuild {
			parts = append(parts, "dirty")
		}
	}
	shortVersion := "devel"
	if len(parts) > 0 {
		shortVersion = strings.Join(parts, "-")
	}

	fmt.Printf("confluence-dump version %s\n", shortVersion)
	return nil
}
