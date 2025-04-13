package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tranquil-tr0/get/pkg/manager"
	"github.com/urfave/cli/v2"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	metadataPath := filepath.Join(homeDir, ".get-metadata.json")
	pm := manager.NewPackageManager(metadataPath)

	app := &cli.App{
		Name:    "get",
		Version: "v0.1.0",
		Usage:   "A package manager for GitHub releases",
		Before: func(c *cli.Context) error {
			return nil
		},
		Authors: []*cli.Author{
			{
				Name:  "tranquil-tr0",
				Email: "tranquil-tr0@github.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "install",
				Category:    "Package Management",
				Usage:       "Install a package from GitHub",
				Description: "Install a package from a GitHub repository. The package must contain a .deb file in its latest release.",
				ArgsUsage:   "<github-repo-url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "release",
						Aliases: []string{"r"},
						Usage:   "Specify a release version to install",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					owner, repo := parseRepoURL(repoURL)
					if owner == "" || repo == "" {
						return fmt.Errorf("Invalid GitHub repository URL")
					}

					pm.SetVerbose(c.Bool("verbose"))
					if err := pm.Install(owner, repo, c.String("release")); err != nil {
						return fmt.Errorf("Error installing package: %v", err)
					}
					fmt.Printf("Successfully installed %s/%s\n", owner, repo)
					return nil
				},
			},
			{
				Name:        "list",
				Category:    "Package Management",
				Usage:       "List installed packages",
				Description: "Display a list of all packages installed through get.",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					packages, err := pm.ListPackages()
					if err != nil {
						return fmt.Errorf("Error listing packages: %v", err)
					}

					if len(packages) == 0 {
						fmt.Println("No packages installed")
						return nil
					}

					fmt.Println("Installed packages:")
					for _, pkg := range packages {
						fmt.Printf("%s/%s (version: %s, installed: %s)\n", pkg.Owner, pkg.Repo, pkg.Version, pkg.InstalledAt)
					}
					return nil
				},
			},
			{
				Name:        "remove",
				Category:    "Package Management",
				Usage:       "Remove an installed package",
				Description: "Remove a previously installed package and clean up its metadata.",
				ArgsUsage:   "<github-repo-url>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					pm.SetVerbose(c.Bool("verbose"))
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					owner, repo := parseRepoURL(repoURL)
					if owner == "" || repo == "" {
						return fmt.Errorf("Invalid GitHub repository URL")
					}

					if err := pm.Remove(owner, repo); err != nil {
						return fmt.Errorf("Error removing package: %v", err)
					}
					fmt.Printf("Successfully removed %s/%s\n", owner, repo)
					return nil
				},
			},
			{
				Name:        "update",
				Category:    "Package Management",
				Usage:       "Update an installed package",
				Description: "Update a package to its latest version from GitHub releases.",
				ArgsUsage:   "<github-repo-url>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					pm.SetVerbose(c.Bool("verbose"))
					if c.NArg() == 0 {
						// TODO: Implement update all packages
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					owner, repo := parseRepoURL(repoURL)
					if owner == "" || repo == "" {
						return fmt.Errorf("Invalid GitHub repository URL")
					}

					if err := pm.Update(owner, repo); err != nil {
						return fmt.Errorf("Error updating package: %v", err)
					}
					fmt.Printf("Successfully updated %s/%s\n", owner, repo)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func parseRepoURL(url string) (owner, repo string) {
	// Remove protocol and domain if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	// Split remaining path into owner and repo
	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
