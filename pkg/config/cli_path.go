package config

import (
	"errors"
	"os"
	"path/filepath"
)

var UserConfigDirectory = getUserConfigDir

func getUserConfigDir() (string, error) {
	// Check for test override first
	if testConfig := os.Getenv("OBSIDIAN_CLI_CONFIG_HOME"); testConfig != "" {
		return testConfig, nil
	}
	return os.UserConfigDir()
}

func CliPath() (cliConfigDir string, cliConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}
	cliConfigDir = filepath.Join(userConfigDir, ObsidianCLIConfigDirectory)
	cliConfigFile = filepath.Join(cliConfigDir, ObsidianCLIConfigFile)
	return cliConfigDir, cliConfigFile, nil
}
