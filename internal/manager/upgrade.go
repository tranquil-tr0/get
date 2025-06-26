package manager

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/output"
	"github.com/tranquil-tr0/get/internal/tools"
)

func (pm *PackageManager) UpgradeAllPackages() error {
	// IMPLEMENTATION:
	/*
		1. Get pending updates using pm.GetPendingUpdates()
		2. If there are pending updates, [for each package with pending update]
			1. Call UpdatePackage(PackageID)
		3. In addition to error logging already implemented, if there still exist pending updates, return an error
		4. If no errors, now return nil
	*/

	// Get pending updates using pm.GetPendingUpdates()
	output.PrintVerboseStart("Loading pending updates")
	pendingUpdates, err := pm.GetAllPendingUpdates()
	if err != nil {
		output.PrintVerboseDebug("UPGRADE", "No pending updates found")
		output.PrintYellow("No pending updates available.")
		return nil
	}
	output.PrintVerboseComplete("Load pending updates", fmt.Sprintf("%d packages", len(pendingUpdates)))

	// If there are pending updates, print the number of pending updates available
	output.PrintYellow("Found %d pending updates.", len(pendingUpdates))
	output.PrintVerboseDebug("UPGRADE", "Pending updates: %v", pendingUpdates)

	// If there are pending updates, call UpdatePackage for each package with pending update
	updateErrors := false
	output.PrintVerboseStart("Processing package upgrades")
	for pkgID := range pendingUpdates {
		output.PrintAction("Upgrading %s...", pkgID)
		output.PrintVerboseStart("Upgrading specific package", pkgID)
		if updateErr := pm.UpgradeSpecificPackage(pkgID); updateErr != nil { // Changed variable name to updateErr
			output.PrintError("Error upgrading %s: %v", pkgID, updateErr)
			output.PrintVerboseError("Upgrade specific package", updateErr)
			updateErrors = true
		} else {
			output.PrintSuccess("Successfully upgraded %s", pkgID)
			output.PrintVerboseComplete("Upgrade specific package", pkgID)
		}
	}
	output.PrintVerboseComplete("Process package upgrades")

	// Reload metadata to check if there are still pending updates
	output.PrintVerboseStart("Verifying upgrade completion")
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		output.PrintVerboseError("Reload metadata for verification", err)
		return fmt.Errorf("failed to reload metadata: %v", err)
	}

	// If there still exist pending updates, return an error
	if len(metadata.PendingUpdates) > 0 && updateErrors {
		output.PrintVerboseError("Upgrade verification", fmt.Errorf("%d packages still pending", len(metadata.PendingUpdates)))
		return fmt.Errorf("some packages could not be upgraded")
	}
	output.PrintVerboseComplete("Verify upgrade completion", "all upgrades successful")

	return nil
}

func (pm *PackageManager) UpgradeSpecificPackage(pkgID string) error {
	// get the pending update version
	output.PrintVerboseStart("Getting pending update version", pkgID)
	pendingReleaseVersion, err := pm.GetPendingUpdate(pkgID)
	if err != nil {
		output.PrintVerboseError("Get pending update version", err)
		return fmt.Errorf("failed checking for pending updates: %s", err)
	}

	if pendingReleaseVersion == "" {
		output.PrintVerboseError("Get pending update version", fmt.Errorf("no pending update found"))
		return fmt.Errorf("no pending update found for package: %s", pkgID)
	}
	output.PrintVerboseComplete("Get pending update version", pendingReleaseVersion)

	// Get new release
	release, err := pm.GithubClient.GetReleaseByTag(pkgID, pendingReleaseVersion)
	if err != nil {
		return err
	}

	// Get saved asset choice
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return err
	}
	pkgMetadata := metadata.Packages[pkgID]
	savedAsset := pkgMetadata.ChosenAsset

	// Check if saved asset is available in the new release
	var chosenAsset *github.Asset
	if savedAsset != "" {
		for i, asset := range release.Assets {
			similar, err := tools.AreAssetNamesSimilar(savedAsset, asset.Name)
			if err != nil {
				output.PrintVerboseError("Asset similarity check", err)
			}
			if similar {
				chosenAsset = &release.Assets[i]
				break
			}
		}
	}

	// Prompt user if necessary
	if chosenAsset != nil {
		if !pm.Yes {
			fmt.Printf("Select \"%s\" as install asset? [Y/n]: ", chosenAsset.Name)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if strings.ToLower(input) == "n" {
				chosenAsset = nil
			}
		}
	} else {
		fmt.Println("Saved asset not found in new release. Please select a new asset.")
		selectedAsset, _, err := pm.SelectAssetInteractively(release)
		if err != nil {
			return err
		}
		chosenAsset = selectedAsset
	}

	if chosenAsset == nil {
		return fmt.Errorf("no asset selected for installation")
	}

	// Save chosen asset
	pkgMetadata.ChosenAsset = chosenAsset.Name
	metadata.Packages[pkgID] = pkgMetadata
	if err := pm.WritePackageManagerMetadata(metadata); err != nil {
		return err
	}

	// Install the chosen asset
	return pm.InstallVersion(pkgID, pendingReleaseVersion, chosenAsset)
}