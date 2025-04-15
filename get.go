/* TODO:
- implement autocomplete
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tranquil-tr0/get/pkg/manager"
	"github.com/tranquil-tr0/get/pkg/output"
	"github.com/tranquil-tr0/get/pkg/tools"

	"github.com/urfave/cli/v2"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		output.PrintError("Error getting home directory: %v", err)
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
				Flags: []cli.Flag{ //TODO: add flag for force install, even if package is already installed
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
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}

					if err := pm.Install(pkgID, c.String("release")); err != nil {
						return fmt.Errorf("Error installing package: %v", err)
					}
					output.PrintSuccess("Successfully installed %s", pkgID)
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
					// Remove the unused variable declaration
					err := pm.PrintInstalledPackages()
					if err != nil {
						return fmt.Errorf("Error listing packages: %v", err)
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
					if c.NArg() != 1 {
						return fmt.Errorf("Please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}

					if err := pm.Remove(pkgID); err != nil {
						return fmt.Errorf("Error removing package: %v", err)
					}
					output.PrintSuccess("Successfully removed %s", pkgID)
					return nil
				},
			},
			{
				Name:        "update",
				Category:    "Package Management",
				Usage:       "Check for package updates",
				Description: "Check for available updates of installed packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Checking for updates...")
					if err := pm.UpdateAllPackages(); err != nil {
						return fmt.Errorf("Error checking for updates: %v", err)
					}
					return nil
				},
			},
			{
				Name:        "upgrade",
				Category:    "Package Management",
				Usage:       "Apply staged upgrades",
				Description: "Install available updates for packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Upgrading packages...")
					if err := pm.UpgradeAllPackages(); err != nil {
						return fmt.Errorf("Error upgrading packages: %v", err)
					}
					output.PrintSuccess("Successfully applied all available updates")
					return nil
				},
			},
			{
				Name:        "update-upgrade",
				Category:    "Package Management",
				Aliases:     []string{"uu", "up"},
				Usage:       "Upgrade outdated packages",
				Description: "Check for updates then upgrade outdated packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					output.PrintAction("Checking for updates...")
					if err := pm.UpdateAllPackages(); err != nil {
						return fmt.Errorf("Error checking for updates: %v", err)
					}

					output.PrintAction("Applying updates...")
					if err := pm.UpgradeAllPackages(); err != nil {
						return fmt.Errorf("Error upgrading packages: %v", err)
					}

					output.PrintSuccess("Successfully applied all available updates")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		output.PrintError("%v", err)
		os.Exit(1)
	}
}
