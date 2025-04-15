package manager

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tranquil-tr0/get/pkg/output"
)

func (pm *PackageManager) UpdateAllPackages() error {
	//IMPLEMENTATION:
	/*
		1. Form the loop by loading get metadata using manager.go to look for list of installed packages
		2. For each installed package (loop):
			1. Call UpdatePackageOrReturnVersions(), print error if any, and print available update with package name, current version, and latest version if (current version != latest version)
		3. Check metadata for duplicate pending updates, and remove if there are any
		In conlusion, all available updates will be listed as pending updates in the metadata, and listed for the user.
	*/

	// Load metadata to get list of installed packages
	metadata, err := pm.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %v", err)
	}

	// Track if we found any updates
	updatesFound := false

	// Loop through each installed package
	for pkgID, pkg := range metadata.Packages {
		// Call UpdatePackageOrReturnVersions for each package
		currentVersion, latestVersion, updateErr := pm.UpdatePackageOrReturnVersions(pkgID)

		// Print error if any
		if updateErr != nil {
			output.PrintError("Error checking for updates for %s: %v", pkgID, updateErr)
			continue
		}

		// Print available update if current version is different from latest version
		if currentVersion != latestVersion {
			output.PrintYellow("Update available for %s: %s -> %s",
				pkgID, pkg.Version, metadata.PendingUpdates[pkgID].TagName)
			updatesFound = true
		}
	}

	// Check metadata for duplicate pending updates and remove them
	metadata, err = pm.LoadMetadata() // Reload to get latest state
	if err != nil {
		return fmt.Errorf("failed to reload metadata: %v", err)
	}

	// Save metadata back
	if err := pm.SaveMetadata(metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %v", err)
	}

	if !updatesFound {
		output.PrintYellow("No updates available.")
	}

	return nil
}
func (pm *PackageManager) UpdatePackageOrReturnVersions(pkgID string) (currentVersion int, latestVersion int, error error) {
	//IMPLEMENTATION:
	/*
		1. From metadata, read the installed version of the package
			2. Call GetLatestRelease() from client, then compare with installed version
			3. If latest version is greater than installed version,
				2. Check if the latest release has a .deb file making use of client.go, if yes:
					1. Check if a pending update of the pkgID package is already listed for the latest version, if yes:
						1. return <installed version>, <latest version>, nil
					2. If no:
						1. add the package and its lastest version number to the metadata as a pending update
						2. return <installed version>, <latest version>, nil
				3. If no, return <installed version>, <latest version>, "latest release does not contain a .deb file"
	*/

	// From metadata, read the installed version of the package
	metadata, err := pm.LoadMetadata()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to load metadata: %v", err)
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return 0, 0, fmt.Errorf("package %s is not installed", pkgID)
	}

	// Parse owner and repo from pkgID (format: owner/repo)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid package ID format: %s", pkgID)
	}
	owner, repo := parts[0], parts[1]

	// Parse version numbers (removing 'v' prefix if present)
	currentVersionStr := strings.TrimPrefix(pkg.Version, "v")
	currentVersionInt, err := parseVersionToInt(currentVersionStr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse current version: %v", err)
	}

	// Call GetLatestRelease from client
	latestRelease, err := pm.GithubClient.GetLatestRelease(owner, repo)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get latest release: %v", err)
	}

	// Parse latest version
	latestVersionStr := strings.TrimPrefix(latestRelease.TagName, "v")
	latestVersionInt, err := parseVersionToInt(latestVersionStr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latest version: %v", err)
	}

	var updatesExist bool = false
	if latestVersionInt != currentVersionInt {
		updatesExist = true
	}
	// Check if there are any pending updates
	// Compare versions
	if latestVersionInt > currentVersionInt {
		// Check if the latest release has a .deb file
		if latestRelease.FindDebPackage() == nil {
			// No .deb file in the latest release
			return currentVersionInt, latestVersionInt, fmt.Errorf("latest release does not contain a .deb file")
		} else {

			// Check if a pending update is already listed for this package
			_, updateExists := metadata.PendingUpdates[pkgID]
			if !updateExists {
				// Add the package and its latest version to pending updates
				metadata.PendingUpdates[pkgID] = *latestRelease
				if err := pm.SaveMetadata(metadata); err != nil {
					return 0, 0, fmt.Errorf("failed to save metadata: %v", err)
				}
			}

			// Print package, current version in red, and latest version in green
			output.PrintNormal("Package: %s, Current Version: %s, Latest Version: %s", pkgID, output.Red(currentVersionStr), output.Green(latestVersionStr))

			return currentVersionInt, latestVersionInt, nil
		}
	}

	// Only return no update available if updatesExist is false
	if !updatesExist {
		return currentVersionInt, latestVersionInt, nil
	}
	// No updates available
	return currentVersionInt, latestVersionInt, nil
}

// Helper function to parse version string to integer for comparison
func parseVersionToInt(version string) (int, error) {
	// Remove any non-numeric characters and parse as integer
	// This is a simple implementation - for more complex versioning,
	// a proper semver parsing library would be better
	cleanVersion := strings.Split(version, ".")
	if len(cleanVersion) == 0 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	// Just use the first number for simple comparison
	majorVersion := cleanVersion[0]
	result, err := strconv.Atoi(majorVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version number: %v", err)
	}

	return result, nil
}
