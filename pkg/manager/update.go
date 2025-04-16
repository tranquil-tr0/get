package manager

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

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
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %v", err)
	}

	// Track if we found any updates
	updatesFound := false

	// Loop through each installed package
	for pkgID, pkg := range metadata.Packages {
		// Call UpdatePackageOrReturnVersions for each package
		currentVersion, latestVersion, updateErr := pm.UpdatePackageOrReturnVersions(pkgID)
		// LOGGING Print current version, latest version, and any error
		output.PrintError("Current version: %d, Latest version: %d, Error: %v", currentVersion, latestVersion, updateErr)
		// Print error if any
		if updateErr != nil {
			output.PrintError("Error checking for updates for %s: %v", pkgID, updateErr)
			continue
		}

		// Print available update if current version is different from latest version
		if currentVersion != latestVersion {
			output.PrintYellow("Update available for %s: %s -> %s",
				pkgID, pkg.Version, latestVersion)
			updatesFound = true
		}
	}

	// Check metadata for duplicate pending updates and remove them
	metadata, err = pm.GetPackageManagerMetadata() // Reload to get latest state
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
func (pm *PackageManager) UpdatePackageOrReturnVersions(pkgID string) (currentVersion int, latestVersion int, err error) {
	//IMPLEMENTATION:
	/*
		1. From metadata, read the installed version of the package
			2. Call GetLatestVersionNumber() from client, then compare with installed version
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
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to load metadata: %v", err)
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return 0, 0, fmt.Errorf("package %s is not installed", pkgID)
	}

	// Get current version from metadata
	currentVersion, err = parseVersionToInt(pkg.Version)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse current version: %v", err)
	}

	// Call GetLatestVersionNumber from client
	latestVersionStr, err := pm.GithubClient.GetLatestVersionNumber(pkgID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get latest version: %v", err)
	}

	// Convert string version to integer
	latestVersion, err = parseVersionToInt(latestVersionStr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latest version: %v", err)
	}

	// Check if there are any pending updates
	// Compare versions
	if latestVersion > currentVersion {
		latestRelease, err := pm.GithubClient.GetLatestRelease(pkgID)
		if err != nil {
			return currentVersion, latestVersion, fmt.Errorf("failed to get latest release: %v", err)
		}
		// Check if the latest release has a .deb file
		if latestRelease.FindDebPackage() == nil {
			// No .deb file in the latest release
			return currentVersion, latestVersion, fmt.Errorf("latest release does not contain a .deb file")
		} else { //there is a deb package
			// Check if a pending update is already listed for this package
			updateVersion, err := pm.GetPendingUpdate(pkgID)
			if err != nil {
				return currentVersion, latestVersion, fmt.Errorf("error checking for existing updates: %s", err)
			}
			if updateVersion == "" { // if there is no pending update
				// Add the package and its latest version to pending updates
				if err := pm.SaveMetadata(metadata); err != nil {
					return 0, 0, fmt.Errorf("failed to save metadata: %v", err)
				}
			}

			return currentVersion, latestVersion, nil
		}
	}

	// No updates available
	return currentVersion, latestVersion, nil
}

// Helper function to parse version string to integer for comparison
func parseVersionToInt(version string) (int, error) {
	// Trim leading/trailing non-numeric characters
	version = strings.TrimFunc(version, func(r rune) bool {
		return !unicode.IsDigit(r)
	})

	// Split into major.minor.patch components
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	// Convert each component to integer and combine into a single comparable number
	// Weights: major*10000 + minor*100 + patch
	var result int
	for i, part := range parts {
		if i >= 3 { // Only consider major.minor.patch
			break
		}
		num, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("failed to parse version component: %v", err)
		}
		switch i {
		case 0: // major
			result += num * 10000
		case 1: // minor
			result += num * 100
		case 2: // patch
			result += num
		}
	}

	return result, nil
}
