package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yakitrak/obsidian-cli/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

var update = flag.Bool("u", false, "update testscript output files")

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"obsidian-cli": func() int {
			cmd.Execute()
			return 0
		},
	}))
}

func TestFrontmatterScript(t *testing.T) {
	t.Parallel()
	
	testscript.Run(t, testscript.Params{
		Dir:                 filepath.Join("pkg", "frontmatter", "testdata", "script"),
		UpdateScripts:       *update,
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			// CRITICAL: Set environment variable so spawned obsidian-cli processes
			// use our test config directory instead of real user config
			testConfigHome := filepath.Join(env.WorkDir, ".config")
			env.Setenv("OBSIDIAN_CLI_CONFIG_HOME", testConfigHome)
			
			// Create isolated test vault
			vaultPath := filepath.Join(env.WorkDir, "test-vault")
			if err := os.MkdirAll(vaultPath, 0o755); err != nil {
				return err
			}

			// Create isolated Obsidian config
			obsidianConfigDir := filepath.Join(testConfigHome, "obsidian")
			if err := os.MkdirAll(obsidianConfigDir, 0o755); err != nil {
				return err
			}

			obsidianConfig := `{"vaults":{"test-vault":{"path":"` + vaultPath + `"}}}`
			if err := os.WriteFile(filepath.Join(obsidianConfigDir, "obsidian.json"), []byte(obsidianConfig), 0o644); err != nil {
				return err
			}

			// Create isolated CLI config
			cliConfigDir := filepath.Join(testConfigHome, "obsidian-cli")
			if err := os.MkdirAll(cliConfigDir, 0o755); err != nil {
				return err
			}

			cliConfig := `{"default_vault_name":"test-vault"}`
			if err := os.WriteFile(filepath.Join(cliConfigDir, "preferences.json"), []byte(cliConfig), 0o644); err != nil {
				return err
			}

			return nil
		},
	})
}
