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
	// IMPLEMENTATION:
	/*
		1. If there are pending updates for the package identified by pkgID,
			1. call Install() in install.go to install the latest version of the package
			2. remove the package from pending updates in metadata
	*/

	// get the pending update version
	output.PrintVerboseStart("Getting pending update version", pkgID)
	pendingRelease, err := pm.GetPendingUpdate(pkgID)
	if err != nil {
		output.PrintVerboseError("Get pending update version", err)
		return fmt.Errorf("failed checking for pending updates: %s", err)
	}

	if pendingRelease == "" {
		output.PrintVerboseError("Get pending update version", fmt.Errorf("no pending update found"))
		return fmt.Errorf("no pending update found for package: %s", pkgID)
	}
	output.PrintVerboseComplete("Get pending update version", pendingRelease)

	// Parse owner and repo from pkgID (format: owner/repo)
	output.PrintVerboseDebug("UPGRADE", "Validating package ID format: %s", pkgID)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		output.PrintVerboseError("Validate package ID", fmt.Errorf("invalid format: %s", pkgID))
		return fmt.Errorf("invalid package ID format: %s", pkgID)
	}

	// Call InstallVersion() to install the specified version of the package
	// Note: InstallVersion calls InstallRelease which already removes the package from pending updates
	output.PrintVerboseStart("Installing updated version", fmt.Sprintf("%s@%s", pkgID, pendingRelease))
	if err := pm.InstallVersion(pkgID, pendingRelease); err != nil {
		output.PrintVerboseError("Install updated version", err)
		return fmt.Errorf("failed to install update for %s: %v", pkgID, err)
	}
	output.PrintVerboseComplete("Install updated version", fmt.Sprintf("%s@%s", pkgID, pendingRelease))

	// No need to manually remove from pending updates since InstallRelease already handles this
	output.PrintVerboseDebug("UPGRADE", "Package upgrade completed successfully")

	return nil
}
