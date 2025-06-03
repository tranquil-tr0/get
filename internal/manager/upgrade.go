package manager

import (
	"fmt"
	"strings"

	"github.com/tranquil-tr0/get/internal/output"
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
	pendingUpdates, err := pm.GetAllPendingUpdates()
	if err != nil {
		output.PrintYellow("No pending updates available.")
		return nil
	}

	// If there are pending updates, print the number of pending updates available
	output.PrintYellow("Found %d pending updates.", len(pendingUpdates))

	// If there are pending updates, call UpdatePackage for each package with pending update
	updateErrors := false
	for pkgID := range pendingUpdates {
		output.PrintAction("Upgrading %s...", pkgID)
		if updateErr := pm.UpgradeSpecificPackage(pkgID); updateErr != nil { // Changed variable name to updateErr
			output.PrintError("Error upgrading %s: %v", pkgID, updateErr)
			updateErrors = true
		} else {
			output.PrintSuccess("Successfully upgraded %s", pkgID)
		}
	}

	// Reload metadata to check if there are still pending updates
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return fmt.Errorf("failed to reload metadata: %v", err)
	}

	// If there still exist pending updates, return an error
	if len(metadata.PendingUpdates) > 0 && updateErrors {
		return fmt.Errorf("some packages could not be upgraded")
	}

	return nil
}

func (pm *PackageManager) UpgradeSpecificPackage(pkgID string) error {
	// IMPLEMENTATION:
	/*
		1. If there are pending updates for the package identified by pkgID,
			1. call Install() in install.go to install the latest version of the package
			2. remove the package from pending updates in metadata
	*/

	// get the pending update version
	pendingRelease, err := pm.GetPendingUpdate(pkgID)
	if err != nil {
		return fmt.Errorf("failed checking for pending updates: %s", err)
	}

	// Parse owner and repo from pkgID (format: owner/repo)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid package ID format: %s", pkgID)
	}

	// Call InstallVersion() to install the specified version of the package
	// Note: InstallVersion calls InstallRelease which already removes the package from pending updates
	if err := pm.InstallVersion(pkgID, pendingRelease); err != nil {
		return fmt.Errorf("failed to install update for %s: %v", pkgID, err)
	}

	// No need to manually remove from pending updates since InstallRelease already handles this

	return nil
}
