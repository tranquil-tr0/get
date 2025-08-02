package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/manager"
	"github.com/tranquil-tr0/get/internal/output"
	"github.com/tranquil-tr0/get/internal/tools"

	"github.com/spf13/cobra"
)

var pm *manager.PackageManager

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		output.PrintError("Error getting home directory: %v", err)
		os.Exit(1)
	}

	metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
	pm = manager.NewPackageManager(metadataPath)

	rootCmd := &cobra.Command{
		Use:     "get",
		Version: "v0.1.0",
		Short:   "A package manager for GitHub releases",
		Long:    "A package manager for GitHub releases that helps you install and manage packages from GitHub without worrying about leaving unupdated packages on your system.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				output.SetVerbose(true)
				pm.Verbose = true
			}
		},
	}

	// Add persistent verbose flag that works for all commands
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Install command
	installCmd := &cobra.Command{
		Use:   "install <github-repo-url> (or <user>/<repo>)",
		Short: "Install a package from GitHub",
		Long:  "Install a package from a GitHub repository. Supports both .deb packages and binary executables. You will be prompted to select which asset to install.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoURL := args[0]
			release, _ := cmd.Flags().GetString("release")
			packageType, _ := cmd.Flags().GetString("tag-prefix")

			output.PrintAction("Parsing repository URL...")
			output.PrintVerboseStart("Parsing repository URL", repoURL)
			pkgID, err := tools.ParseRepoURL(repoURL)
			if err != nil {
				output.PrintVerboseError("Parse repository URL", err)
				return fmt.Errorf("failed to parse repository URL: %v", err)
			}
			output.PrintVerboseComplete("Parse repository URL", pkgID)

			// Prepare options if tag prefix is specified
			var options *github.ReleaseOptions
			if packageType != "" {
				options = &github.ReleaseOptions{
					TagPrefix: packageType,
				}
			}

			output.PrintAction("Installing package...")
			output.PrintVerboseStart("Installing package", pkgID)
			if err := pm.InstallWithOptions(pkgID, release, options); err != nil {
				output.PrintVerboseError("Install package", err)
				return fmt.Errorf("error installing package: %v", err)
			}
			output.PrintVerboseComplete("Install package", pkgID)
			output.PrintSuccess("Successfully installed %s", pkgID)
			return nil
		},
	}
	installCmd.Flags().StringP("release", "r", "", "Specify a release version to install")
	installCmd.Flags().StringP("tag-prefix", "t", "", "Specify tag prefix for package variants (e.g., \"auth-\" for auth-v1.0.0 tags)")
	rootCmd.AddCommand(installCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Long:  "Display a list of all packages installed through get.",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.PrintVerboseStart("Loading installed packages")
			err := pm.PrintInstalledPackages()
			if err != nil {
				output.PrintVerboseError("Load installed packages", err)
				return fmt.Errorf("error listing packages: %v", err)
			}
			output.PrintVerboseComplete("Load installed packages")
			return nil
		},
	}
	rootCmd.AddCommand(listCmd)

	// Remove command
	removeCmd := &cobra.Command{
		Use:   "remove <github-repo-url>",
		Short: "Remove an installed package",
		Long:  "Remove a previously installed package and clean up its metadata.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoURL := args[0]

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
	}
	rootCmd.AddCommand(removeCmd)

	// Update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Check for package updates",
		Long:  "Check for available updates of installed packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.PrintAction("Checking for updates...")
			output.PrintVerboseStart("Checking for package updates")
			if err := pm.UpdateAllPackages(); err != nil {
				output.PrintVerboseError("Check for updates", err)
				return fmt.Errorf("error checking for updates: %v", err)
			}
			output.PrintVerboseComplete("Check for package updates")
			return nil
		},
	}
	rootCmd.AddCommand(updateCmd)

	// Upgrade command
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Apply staged upgrades",
		Long:  "Install available updates for packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			if yes {
				pm.Yes = true
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
	}
	upgradeCmd.Flags().BoolP("yes", "y", false, "Skip interactive prompts and use saved asset choices")
	rootCmd.AddCommand(upgradeCmd)

	// Update-upgrade command (with alias)
	updateUpgradeCmd := &cobra.Command{
		Use:     "update-upgrade",
		Aliases: []string{"up"},
		Short:   "Upgrade outdated packages",
		Long:    "Check for updates then upgrade outdated packages",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	}
	rootCmd.AddCommand(updateUpgradeCmd)

	if err := rootCmd.Execute(); err != nil {
		output.PrintError("%v", err)
		os.Exit(1)
	}
}
