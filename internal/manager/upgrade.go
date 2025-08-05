package manager

import (
	"fmt"

	"context"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/tools"
)

func (pm *PackageManager) UpgradeAllPackages(ctx context.Context) error {
	pendingUpdates, err := pm.GetAllPendingUpdates()
	if err != nil {
		pm.Out.PrintInfo("No pending updates available.")
		return nil
	}

	pm.Out.PrintInfo("Found %d pending updates.", len(pendingUpdates))

	updateErrors := false
	for pkgID := range pendingUpdates {
		pm.Out.PrintStatus("Upgrading %s...", pkgID)
		if updateErr := pm.UpgradeSpecificPackage(ctx, pkgID); updateErr != nil {
			if updateErr == context.Canceled {
				pm.Out.PrintInfo("Upgrade cancelled for %s", pkgID)
				continue
			}
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

func (pm *PackageManager) UpgradeSpecificPackage(ctx context.Context, pkgID string) error {
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
			yes, err := pm.Out.PromptYesNo(fmt.Sprintf("Select \"%s\" as install asset?", chosenAsset.Name))
			if err != nil {
				return err
			}
			if !yes {
				selectedAsset, _, err := pm.SelectAssetInteractively(ctx, release)
				if err != nil {
					if err == context.Canceled {
						return nil
					}
					return err
				}
				chosenAsset = selectedAsset
			}
		}
	} else {
		pm.Out.PrintInfo("Saved asset was not found in the new release. Please select a new asset.")
		selectedAsset, _, err := pm.SelectAssetInteractively(ctx, release)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
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

	return pm.InstallVersion(ctx, pkgID, pendingReleaseVersion, chosenAsset)
}
