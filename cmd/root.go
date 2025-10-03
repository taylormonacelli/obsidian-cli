package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	logging "gopkg.in/op/go-logging.v1"
)

var rootCmd = &cobra.Command{
	Use:     "obsidian-cli",
	Short:   "obsidian-cli - CLI to open, search, move, create, delete and update notes",
	Version: "v0.1.9",
	Long:    "obsidian-cli - CLI to open, search, move, create, delete and update notes with frontmatter support",
}

func init() {
	// Configure go-logging to suppress yq warnings
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(`%{message}`)
	backendFormatted := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatted)
	backendLeveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backendLeveled)

	// Configure standard log package (used by yq)
	log.SetFlags(0)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
