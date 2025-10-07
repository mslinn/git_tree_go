package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v2"
)

// Config represents the application's configuration.
type Config struct {
	GitTimeout   int      `yaml:"git_timeout"`
	Verbosity    int      `yaml:"verbosity"`
	DefaultRoots []string `yaml:"default_roots"`
}

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	configPath := filepath.Join(home, ".git-tree.yml")

	fmt.Printf("Welcome to git-tree configuration.\n")
	fmt.Printf("This utility will help you create a configuration file at %s\n", configPath)
	fmt.Printf("You can press Enter to accept default values presented within brackets.\n\n")

	existingConfig := &Config{}
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Failed to read existing config: %v", err)
		}
		err = yaml.Unmarshal(data, existingConfig)
		if err != nil {
			log.Fatalf("Failed to unmarshal existing config: %v", err)
		}
	}

	qs := []*survey.Question{
		{
			Name:     "GitTimeout",
			Prompt:   &survey.Input{Message: "Git command timeout in seconds?"},
			Validate: survey.Required,
		},
		{
			Name:     "Verbosity",
			Prompt:   &survey.Input{Message: "Default verbosity level (0=quiet, 1=normal, 2=verbose, 3=debug)?"},
			Validate: survey.Required,
		},
		{
			Name:     "DefaultRoots",
			Prompt:   &survey.Input{Message: "Default root directories (space-separated)?"},
			Validate: survey.Required,
		},
	}

	answers := struct {
		GitTimeout   int
		Verbosity    int
		DefaultRoots string
	}{}

	if existingConfig.GitTimeout != 0 {
		qs[0].Prompt.(*survey.Input).Default = fmt.Sprintf("%d", existingConfig.GitTimeout)
	}
	if existingConfig.Verbosity != 0 {
		qs[1].Prompt.(*survey.Input).Default = fmt.Sprintf("%d", existingConfig.Verbosity)
	}
	if len(existingConfig.DefaultRoots) > 0 {
		qs[2].Prompt.(*survey.Input).Default = strings.Join(existingConfig.DefaultRoots, " ")
	}

	err = survey.Ask(qs, &answers)
	if err != nil {
		log.Fatalf("Error during survey: %v", err)
	}

	newConfig := &Config{
		GitTimeout:   answers.GitTimeout,
		Verbosity:    answers.Verbosity,
		DefaultRoots: strings.Fields(answers.DefaultRoots),
	}

	data, err := yaml.Marshal(newConfig)
	if err != nil {
		log.Fatalf("Failed to marshal new config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", configPath)
}
