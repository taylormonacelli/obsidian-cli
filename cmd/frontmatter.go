package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/frontmatter"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// isReadOnlyExpression checks if a yq expression is read-only (query) or modifying (mutation)
func isReadOnlyExpression(expr string) bool {
	expr = strings.TrimSpace(expr)
	// Expressions with assignment operators are mutations
	if strings.Contains(expr, "=") && !strings.Contains(expr, "==") && !strings.Contains(expr, "!=") && !strings.Contains(expr, "<=") && !strings.Contains(expr, ">=") {
		return false
	}
	return true
}

var frontmatterCmd = &cobra.Command{
	Use:     "frontmatter [expression] <note-name>",
	Aliases: []string{"fm"},
	Short:   "Query or modify note frontmatter",
	Long:    "Query or modify YAML frontmatter in markdown files using yq expressions. The expression syntax determines the operation: queries like '.title' read values, mutations like '.title = \"value\"' modify the file.",
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
		isModifying := !isReadOnlyExpression(expression)
		result, err := frontmatter.Run(extractor, expression, mdContent, isModifying)
		if err != nil {
			handleFrontmatterError(err, noteName)
		}

		// Guard: if read-only expression, just print and return early
		if !isModifying {
			fmt.Print(result)
			return
		}

		// Guard: if content unchanged, return early
		if result == mdContent {
			return
		}

		// Validate the output before writing
		validationResult, err := extractor.Extract(result)
		if err != nil {
			log.Fatalf("failed to validate output: %v", err)
		}
		if !validationResult.IsValid {
			if errors.Is(validationResult.ValidationError, frontmatter.ErrScalarFrontmatter) {
				log.Fatalf("Error: Expression would create invalid frontmatter (scalar value instead of key-value pairs).\nYour expression '%s' is a query, not a mutation.\nDid you mean to set a value? Example: .title = \"value\"", expression)
			}
			log.Fatalf("Error: Expression would create invalid frontmatter: %v", validationResult.ValidationError)
		}

		if err := os.WriteFile(filename, []byte(result), 0o644); err != nil {
			log.Fatalf("failed to write file %s: %v", filename, err)
		}
	},
}

func handleFrontmatterError(err error, noteName string) {
	if errors.Is(err, frontmatter.ErrScalarFrontmatter) {
		log.Fatalf("Error: The existing frontmatter in '%s' is invalid.\nIt contains a scalar value instead of key-value pairs.\nFix the file's frontmatter or use: obsidian-cli frontmatter '{}' '%s'", noteName, noteName)
	}
	if errors.Is(err, frontmatter.ErrInvalidFrontmatter) {
		log.Fatalf("Error: The existing frontmatter in '%s' is invalid.\nIt's missing the closing '---' delimiter.", noteName)
	}
	log.Fatal(err)
}

func init() {
	rootCmd.AddCommand(frontmatterCmd)
	frontmatterCmd.PersistentFlags().String("vault", "", "vault name")
}
