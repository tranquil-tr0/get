/* TODO:
- implement autocomplete
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tranquil-tr0/get/internal/manager"
	"github.com/tranquil-tr0/get/internal/output"
	"github.com/tranquil-tr0/get/internal/tools"

	"github.com/urfave/cli/v2"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		output.PrintError("Error getting home directory: %v", err)
		os.Exit(1)
	}

	metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
	pm := manager.NewPackageManager(metadataPath)

	app := &cli.App{
		Name:    "get",
		Version: "v0.1.0",
		Usage:   "A package manager for GitHub releases",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output",
			},
		},
		Before: func(c *cli.Context) error {
			// Set global verbose state based on global or command flags
			if c.Bool("verbose") {
				output.SetVerbose(true)
			}
			return nil
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
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					if c.NArg() != 1 {
						return fmt.Errorf("please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					output.PrintVerboseStart("Parsing repository URL", repoURL)
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						output.PrintVerboseError("Parse repository URL", err)
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}
					output.PrintVerboseComplete("Parse repository URL", pkgID)

					output.PrintVerboseStart("Installing package", pkgID)
					if err := pm.Install(pkgID, c.String("release")); err != nil {
						output.PrintVerboseError("Install package", err)
						return fmt.Errorf("error installing package: %v", err)
					}
					output.PrintVerboseComplete("Install package", pkgID)
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
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					output.PrintVerboseStart("Loading installed packages")
					err := pm.PrintInstalledPackages()
					if err != nil {
						output.PrintVerboseError("Load installed packages", err)
						return fmt.Errorf("error listing packages: %v", err)
					}
					output.PrintVerboseComplete("Load installed packages")
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
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					if c.NArg() != 1 {
						return fmt.Errorf("please provide a GitHub repository URL")
					}

					repoURL := c.Args().First()
					output.PrintVerboseStart("Parsing repository URL", repoURL)
					pkgID, err := tools.ParseRepoURL(repoURL)
					if err != nil {
						output.PrintVerboseError("Parse repository URL", err)
						return fmt.Errorf("failed to parse repository URL: %v", err)
					}
					output.PrintVerboseComplete("Parse repository URL", pkgID)

					output.PrintVerboseStart("Removing package", pkgID)
					if err := pm.Remove(pkgID); err != nil {
						output.PrintVerboseError("Remove package", err)
						return fmt.Errorf("error removing package: %v", err)
					}
					output.PrintVerboseComplete("Remove package", pkgID)
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
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					output.PrintAction("Checking for updates...")
					output.PrintVerboseStart("Checking for package updates")
					if err := pm.UpdateAllPackages(); err != nil {
						output.PrintVerboseError("Check for updates", err)
						return fmt.Errorf("error checking for updates: %v", err)
					}
					output.PrintVerboseComplete("Check for package updates")
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
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					output.PrintAction("Upgrading packages...")
					output.PrintVerboseStart("Upgrading packages")
					if err := pm.UpgradeAllPackages(); err != nil {
						output.PrintVerboseError("Upgrade packages", err)
						return fmt.Errorf("error upgrading packages: %v", err)
					}
					output.PrintVerboseComplete("Upgrade packages")
					output.PrintSuccess("Successfully applied all available updates")
					return nil
				},
			},
			{
				Name:        "update-upgrade",
				Category:    "Package Management",
				Aliases:     []string{"up"},
				Usage:       "Upgrade outdated packages",
				Description: "Check for updates then upgrade outdated packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: func(c *cli.Context) error {
					// Set verbose state from command flag
					if c.Bool("verbose") {
						output.SetVerbose(true)
						pm.Verbose = true
					}

					output.PrintAction("Checking for updates...")
					output.PrintVerboseStart("Checking for package updates")
					if err := pm.UpdateAllPackages(); err != nil {
						output.PrintVerboseError("Check for updates", err)
						return fmt.Errorf("error checking for updates: %v", err)
					}
					output.PrintVerboseComplete("Check for package updates")

					output.PrintAction("Applying updates...")
					output.PrintVerboseStart("Upgrading packages")
					if err := pm.UpgradeAllPackages(); err != nil {
						output.PrintVerboseError("Upgrade packages", err)
						return fmt.Errorf("error upgrading packages: %v", err)
					}
					output.PrintVerboseComplete("Upgrade packages")

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
