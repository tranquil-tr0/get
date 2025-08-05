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
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	metadataPath := filepath.Join(homeDir, ".local/share/get/get.json")
	out := output.NewCLIOutput()
	pm = manager.NewPackageManager(metadataPath, out)

	rootCmd := &cobra.Command{
		Use:     "get",
		Version: "v0.1.0",
		Short:   "A package manager for GitHub releases",
		Long:    "A package manager for GitHub releases that helps you install and manage packages from GitHub without worrying about leaving unupdated packages on your system.",
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

			pkgID, err := tools.ParseRepoURL(repoURL)
			if err != nil {
				return fmt.Errorf("failed to parse repository URL: %v", err)
			}

			var options *github.ReleaseOptions
			if packageType != "" {
				options = &github.ReleaseOptions{
					TagPrefix: packageType,
				}
			}

			if err := pm.InstallWithOptions(cmd.Context(), pkgID, release, options); err != nil {
				return fmt.Errorf("error installing package: %v", err)
			}
			pm.Out.PrintSuccess("Successfully installed %s", pkgID)
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
			sortedKeys, packages, err := pm.ListInstalledPackages()
			if err != nil {
				return err
			}

			if len(sortedKeys) == 0 {
				pm.Out.PrintInfo("No packages installed.")
			} else {
				pm.Out.PrintInfo("Installed packages:")
				for _, pkgID := range sortedKeys {
					pkg := packages[pkgID]
					pm.Out.PrintInfo(" %s (Version: %s, Installed: %s)", pkgID, pkg.Version, pkg.InstalledAt)
				}
			}
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

			pkgID, err := tools.ParseRepoURL(repoURL)
			if err != nil {
				return fmt.Errorf("failed to parse repository URL: %v", err)
			}

			if err := pm.Remove(pkgID); err != nil {
				return fmt.Errorf("error removing package: %v", err)
			}
			pm.Out.PrintSuccess("Successfully removed %s", pkgID)
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
			updates, err := pm.UpdateAllPackages()
			if err != nil {
				return err
			}

			if len(updates) == 0 {
				pm.Out.PrintInfo("No updates available.")
				return nil
			}

			pm.Out.PrintInfo("Available updates:")
			for pkgID, version := range updates {
				pm.Out.PrintInfo("  %s: %s", pkgID, version)
				// TODO: add update available from version to version
			}
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

			if err := pm.UpgradeAllPackages(cmd.Context()); err != nil {
				return fmt.Errorf("error upgrading packages: %v", err)
			}
			pm.Out.PrintSuccess("Successfully applied all available updates")
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
			pm.Out.PrintStatus("Checking for updates...")
			if _, err := pm.UpdateAllPackages(); err != nil {
				return fmt.Errorf("error checking for updates: %v", err)
			}

			pm.Out.PrintStatus("Applying updates...")
			if err := pm.UpgradeAllPackages(cmd.Context()); err != nil {
				return fmt.Errorf("error upgrading packages: %v", err)
			}

			pm.Out.PrintSuccess("Successfully applied all available updates")
			return nil
		},
	}
	rootCmd.AddCommand(updateUpgradeCmd)

	if err := rootCmd.Execute(); err != nil {
		pm.Out.PrintError("%v", err)
		os.Exit(1)
	}
}
