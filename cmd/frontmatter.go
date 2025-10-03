package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Yakitrak/obsidian-cli/pkg/frontmatter"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var frontmatterCmd = &cobra.Command{
	Use:     "frontmatter [expression] <note-name>",
	Aliases: []string{"fm"},
	Short:   "Work with YAML frontmatter",
	Long:    "Query and modify YAML frontmatter in markdown files. Defaults to print mode if no subcommand specified.",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var expression, noteName string

		if len(args) == 1 {
			expression = "."
			noteName = args[0]
		} else {
			expression = args[0]
			noteName = args[1]
		}

		vaultName, _ := cmd.Flags().GetString("vault")
		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		filename := filepath.Join(vaultPath, obsidian.AddMdSuffix(noteName))

		mdBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("failed to read file %s: %v", filename, err)
		}
		mdContent := string(mdBytes)

		extractor := frontmatter.NewYAMLFrontmatterExtractor()
		result, err := frontmatter.Run(extractor, expression, mdContent, false)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(result)
	},
}

var printFrontmatterCmd = &cobra.Command{
	Use:     "print [expression] <note-name>",
	Aliases: []string{"p"},
	Short:   "Print frontmatter (read-only)",
	Long:    "Query and display YAML frontmatter without modifying the file. If no expression is provided, defaults to '.' (show all).",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var expression, noteName string

		if len(args) == 1 {
			expression = "."
			noteName = args[0]
		} else {
			expression = args[0]
			noteName = args[1]
		}

		vaultName, _ := cmd.Flags().GetString("vault")
		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		filename := filepath.Join(vaultPath, obsidian.AddMdSuffix(noteName))

		mdBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("failed to read file %s: %v", filename, err)
		}
		mdContent := string(mdBytes)

		extractor := frontmatter.NewYAMLFrontmatterExtractor()
		result, err := frontmatter.Run(extractor, expression, mdContent, false)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(result)
	},
}

var editFrontmatterCmd = &cobra.Command{
	Use:     "edit <expression> <note-name>",
	Aliases: []string{"e"},
	Short:   "Edit frontmatter (in-place)",
	Long:    "Modify YAML frontmatter and write changes back to the file.",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		expression := args[0]
		noteName := args[1]

		vaultName, _ := cmd.Flags().GetString("vault")
		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		filename := filepath.Join(vaultPath, obsidian.AddMdSuffix(noteName))

		mdBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("failed to read file %s: %v", filename, err)
		}
		mdContent := string(mdBytes)

		extractor := frontmatter.NewYAMLFrontmatterExtractor()
		result, err := frontmatter.Run(extractor, expression, mdContent, true)
		if err != nil {
			log.Fatal(err)
		}

		if result == mdContent {
			return
		}

		if err := os.WriteFile(filename, []byte(result), 0o644); err != nil {
			log.Fatalf("failed to write file %s: %v", filename, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(frontmatterCmd)
	frontmatterCmd.PersistentFlags().String("vault", "", "vault name")
	frontmatterCmd.AddCommand(printFrontmatterCmd)
	frontmatterCmd.AddCommand(editFrontmatterCmd)
}
