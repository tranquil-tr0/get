package manager

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/tools"
)

func (pm *PackageManager) UpgradeAllPackages() error {
	pendingUpdates, err := pm.GetAllPendingUpdates()
	if err != nil {
		pm.Out.PrintInfo("No pending updates available.")
		return nil
	}

	pm.Out.PrintInfo("Found %d pending updates.", len(pendingUpdates))

	updateErrors := false
	for pkgID := range pendingUpdates {
		pm.Out.PrintAction("Upgrading %s...", pkgID)
		if updateErr := pm.UpgradeSpecificPackage(pkgID); updateErr != nil {
			pm.Out.PrintError("Error upgrading %s: %v", pkgID, updateErr)
			updateErrors = true
		} else {
			pm.Out.PrintSuccess("Successfully upgraded %s", pkgID)
		}
	}

	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return fmt.Errorf("failed to reload metadata: %v", err)
	}

	if len(metadata.PendingUpdates) > 0 && updateErrors {
		return fmt.Errorf("some packages could not be upgraded")
	}

	return nil
}

func (pm *PackageManager) UpgradeSpecificPackage(pkgID string) error {
	pendingReleaseVersion, err := pm.GetPendingUpdate(pkgID)
	if err != nil {
		return fmt.Errorf("failed checking for pending updates: %s", err)
	}

	if pendingReleaseVersion == "" {
		return fmt.Errorf("no pending update found for package: %s", pkgID)
	}

	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return err
	}
	pkgMetadata := metadata.Packages[pkgID]

	var release *github.Release
	if pkgMetadata.TagPrefix != "" {
		options := &github.ReleaseOptions{
			TagPrefix: pkgMetadata.TagPrefix,
		}
		release, err = pm.GithubClient.GetReleaseByTagWithOptions(pkgID, pendingReleaseVersion, options)
	} else {
		release, err = pm.GithubClient.GetReleaseByTag(pkgID, pendingReleaseVersion)
	}
	if err != nil {
		return err
	}

	savedAsset := pkgMetadata.ChosenAsset

	var chosenAsset *github.Asset
	if savedAsset != "" {
		for i, asset := range release.Assets {
			similar, err := tools.AreAssetNamesSimilar(savedAsset, asset.Name)
			if err != nil {
			}
			if similar {
				chosenAsset = &release.Assets[i]
				break
			}
		}
	}

	if chosenAsset != nil {
		if !pm.Yes {
			fmt.Printf("Select \"%s\" as install asset? [Y/n]: ", chosenAsset.Name)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			if strings.ToLower(input) == "n" {
				selectedAsset, _, err := pm.SelectAssetInteractively(release)
				if err != nil {
					return err
				}
				chosenAsset = selectedAsset
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

	pkgMetadata.ChosenAsset = chosenAsset.Name
	metadata.Packages[pkgID] = pkgMetadata
	if err := pm.WritePackageManagerMetadata(metadata); err != nil {
		return err
	}

	return pm.InstallVersion(pkgID, pendingReleaseVersion, chosenAsset)
}
