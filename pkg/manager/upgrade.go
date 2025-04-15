package manager

import (
	"fmt"
	"strings"
)

func (pm *PackageManager) UpgradeAllPackages() error {
	// IMPLEMENTATION:
	/*
		1. Load the get metadata using manager.go
		2. Check metadata for pending updates
		3. If there are pending updates, [for each package with pending update]
			1. Call UpdatePackage(PackageID)
		4. In addition to error logging already implemented, if there still exist pending updates, return an error
		5. If no errors, now return nil
	*/

	// Load the get metadata using manager.go
	metadata, err := pm.loadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %v", err)
	}

	// Check metadata for pending updates
	if len(metadata.PendingUpdates) == 0 {
		fmt.Println("No pending updates available.")
		return nil
	}

	// If there are pending updates, call UpdatePackage for each package with pending update
	updateErrors := false
	for pkgID := range metadata.PendingUpdates {
		fmt.Printf("Upgrading %s...\n", pkgID)
		if updateErr := pm.UpdatePackage(pkgID); updateErr != nil { // Changed variable name to updateErr
			fmt.Printf("Error upgrading %s: %v\n", pkgID, updateErr)
			updateErrors = true
		} else {
			fmt.Printf("Successfully upgraded %s\n", pkgID)
		}
	}

	// Reload metadata to check if there are still pending updates
	metadata, err = pm.loadMetadata()
	if err != nil {
		return fmt.Errorf("failed to reload metadata: %v", err)
	}

	// If there still exist pending updates, return an error
	if len(metadata.PendingUpdates) > 0 && updateErrors {
		return fmt.Errorf("some packages could not be upgraded")
	}

	return nil
}

func (pm *PackageManager) UpdatePackage(pkgID string) error {
	// IMPLEMENTATION:
	/*
		1. If there are pending updates for the package identified by pkgID,
			1. call Install() in install.go to install the latest version of the package
			2. remove the package from pending updates in metadata
	*/

	// Load metadata to check for pending updates
	metadata, err := pm.loadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %v", err)
	}

	// Check if there are pending updates for the package identified by pkgID
	pendingRelease, exists := metadata.PendingUpdates[pkgID]
	if !exists {
		return fmt.Errorf("no pending update for package %s", pkgID)
	}

	// Parse owner and repo from pkgID (format: owner/repo)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid package ID format: %s", pkgID)
	}
	owner, repo := parts[0], parts[1]

	// Call Install() in install.go to install the latest version of the package
	if err := pm.Install(owner, repo, pendingRelease.TagName); err != nil {
		return fmt.Errorf("failed to install update for %s: %v", pkgID, err)
	}

	// Remove the package from pending updates in metadata
	delete(metadata.PendingUpdates, pkgID)
	if err := pm.saveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to update metadata after upgrade: %v", err)
	}

	return nil
}
