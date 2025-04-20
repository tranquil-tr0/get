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
		// Call UpdatePackageAndReturnVersions for each package
		hasNewUpdate, latestVersionString, updateErr := pm.UpdatePackageAndReturnNewVersion(pkgID)
		// Print error if any
		if updateErr != nil {
			output.PrintError("Error checking for updates for %s: %v", pkgID, updateErr)
			continue
		}

		// Print available update if current version is different from latest version
		if hasNewUpdate {
			output.PrintYellow("Update available for %s: %s -> %s", pkgID, pkg.Version, latestVersionString)
			updatesFound = true
		}
	}

	if !updatesFound {
		output.PrintYellow("No updates available.")
	}

	return nil
}

// UpdatePackageAndReturnNewVersion marks a package update if it hasn't already been marked
// and returns the new version that it has been marked to update to
func (pm *PackageManager) UpdatePackageAndReturnNewVersion(pkgID string) (hasNewUpdate bool, latestVersionString string, err error) {
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

	// hasUpdate default value
	hasNewUpdate = false

	// From metadata, read the installed version of the package
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to load metadata: %v", err)
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return hasNewUpdate, "", fmt.Errorf("package %s is not installed", pkgID)
	}

	// Get current version from metadata
	currentVersion, err := parseVersionToInt(pkg.Version)
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to parse current version: %v", err)
	}

	// Call GetLatestVersionNumber from client
	latestVersionString, err = pm.GithubClient.GetLatestVersionName(pkgID)
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to get latest version: %v", err)
	}

	// Convert string version to integer
	latestVersionInt, err := parseVersionToInt(latestVersionString)
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to parse latest version: %v", err)
	}

	// Check if there are any pending updates
	// Compare versions
	if latestVersionInt > currentVersion {
		latestRelease, err := pm.GithubClient.GetLatestRelease(pkgID)
		if err != nil {
			return hasNewUpdate, latestVersionString, fmt.Errorf("failed to get latest release: %v", err)
		}
		// Check if the latest release has a .deb file
		if latestRelease.FindDebPackage() == nil {
			// No .deb file in the latest release
			return hasNewUpdate, latestVersionString, fmt.Errorf("latest release does not contain a .deb file")
		} else { //there is a deb package
			// Check if a pending update is already listed for this package
			updateVersion, err := pm.GetPendingUpdate(pkgID)
			if err != nil {
				return hasNewUpdate, latestVersionString, fmt.Errorf("error checking for existing updates: %s", err)
			}
			if updateVersion == "" { // if there is no pending update, (if it is a new update)
				// We should add the package and its latest version to pending updates
				hasNewUpdate = true
				if err := pm.SaveMetadata(metadata); err != nil {
					return hasNewUpdate, "", fmt.Errorf("failed to save metadata: %v", err)
				}
			}

			return hasNewUpdate, latestVersionString, nil
		}
	}

	// No updates available
	return hasNewUpdate, latestVersionString, nil
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
