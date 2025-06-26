package manager

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tranquil-tr0/get/internal/output"
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
	output.PrintVerboseStart("Loading package metadata for update check")
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		output.PrintVerboseError("Load package metadata", err)
		return fmt.Errorf("failed to load metadata: %v", err)
	}
	output.PrintVerboseComplete("Load package metadata", fmt.Sprintf("%d packages", len(metadata.Packages)))

	// Track if we found any updates and error count
	updatesFound := false
	errorCount := 0
	var failedPackages []string

	output.PrintVerboseStart("Checking for updates on all packages")
	// Loop through each installed package
	for pkgID, pkg := range metadata.Packages {
		output.PrintVerboseDebug("UPDATE", "Checking package: %s (current: %s)", pkgID, pkg.Version)
		// Call UpdatePackageAndReturnVersions for each package
		hasNewUpdate, latestVersionString, updateErr := pm.UpdatePackageAndReturnNewVersion(pkgID)
		// Print error if any
		if updateErr != nil {
			output.PrintError("Error checking for updates for %s: %v", pkgID, updateErr)
			output.PrintVerboseError("Check update for package", updateErr)
			errorCount++
			failedPackages = append(failedPackages, pkgID)
			continue
		}

		// Print available update if current version is different from latest version
		if hasNewUpdate {
			output.PrintYellow("Update available for %s: %s -> %s", pkgID, pkg.Version, latestVersionString)
			output.PrintVerboseDebug("UPDATE", "Found update for %s: %s -> %s", pkgID, pkg.Version, latestVersionString)
			updatesFound = true
		} else {
			output.PrintVerboseDebug("UPDATE", "No update available for %s (latest: %s)", pkgID, latestVersionString)
		}
	}
	output.PrintVerboseComplete("Check for updates on all packages")

	if !updatesFound {
		output.PrintYellow("No updates available.")
		output.PrintVerboseDebug("UPDATE", "No updates found for any packages")
	}

	// Provide error summary if there were failures
	if errorCount > 0 {
		output.PrintVerboseError("Update check summary", fmt.Errorf("%d packages failed", errorCount))
		return fmt.Errorf("failed to check updates for %d package(s): %v", errorCount, failedPackages)
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
	output.PrintVerboseStart("Loading metadata for version check", pkgID)
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		output.PrintVerboseError("Load metadata for version check", err)
		return hasNewUpdate, "", fmt.Errorf("failed to load metadata: %v", err)
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		output.PrintVerboseError("Find package in metadata", fmt.Errorf("package not installed"))
		return hasNewUpdate, "", fmt.Errorf("package %s is not installed", pkgID)
	}
	output.PrintVerboseComplete("Load metadata for version check", fmt.Sprintf("current: %s", pkg.Version))

	// Get current version from metadata
	output.PrintVerboseDebug("VERSION", "Parsing current version: %s", pkg.Version)
	currentVersion, err := parseVersionToInt(pkg.Version)
	if err != nil {
		output.PrintVerboseError("Parse current version", err)
		return hasNewUpdate, "", fmt.Errorf("failed to parse current version: %v", err)
	}

	// Call GetLatestVersionNumber from client
	output.PrintVerboseStart("Fetching latest version from GitHub", pkgID)
	latestVersionString, err = pm.GithubClient.GetLatestVersionName(pkgID)
	if err != nil {
		output.PrintVerboseError("Fetch latest version", err)
		return hasNewUpdate, "", fmt.Errorf("failed to get latest version: %v", err)
	}
	output.PrintVerboseComplete("Fetch latest version from GitHub", latestVersionString)

	// Convert string version to integer
	output.PrintVerboseDebug("VERSION", "Parsing latest version: %s", latestVersionString)
	latestVersionInt, err := parseVersionToInt(latestVersionString)
	if err != nil {
		output.PrintVerboseError("Parse latest version", err)
		return hasNewUpdate, "", fmt.Errorf("failed to parse latest version: %v", err)
	}

	// Check if there are any pending updates
	// Compare versions
	output.PrintVerboseDebug("VERSION", "Comparing versions: current=%d, latest=%d", currentVersion, latestVersionInt)
	if latestVersionInt > currentVersion {
		output.PrintVerboseStart("Checking latest release for .deb package", pkgID)
		latestRelease, err := pm.GithubClient.GetLatestRelease(pkgID)
		if err != nil {
			output.PrintVerboseError("Fetch latest release", err)
			return hasNewUpdate, latestVersionString, fmt.Errorf("failed to get latest release: %v", err)
		}

		// Check if the latest release has a .deb file, only if the install type is "deb"
		if pkg.InstallType == "deb" {
			debPackage := latestRelease.FindDebPackage()
			if debPackage == nil {
				// No .deb file in the latest release
				output.PrintVerboseError("Find .deb package", fmt.Errorf("no .deb file in latest release"))
				return hasNewUpdate, latestVersionString, fmt.Errorf("latest release does not contain a .deb file")
			}
			output.PrintVerboseComplete("Check latest release for .deb package", debPackage.Name)
		} else {
			output.PrintVerboseDebug("UPDATE", "Skipping .deb check for non-deb package type: %s", pkg.InstallType)
		}

		// Check if a pending update is already listed for this package
		output.PrintVerboseStart("Checking for existing pending update", pkgID)
		updateVersion, err := pm.GetPendingUpdate(pkgID)
		if err != nil {
			output.PrintVerboseError("Check existing pending update", err)
			return hasNewUpdate, latestVersionString, fmt.Errorf("error checking for existing updates: %s", err)
		}

		if updateVersion == "" { // if there is no pending update, (if it is a new update)
			output.PrintVerboseDebug("UPDATE", "Adding new pending update: %s -> %s", pkg.Version, latestVersionString)
			hasNewUpdate = true
			// Add the pending update to metadata
			metadata.PendingUpdates[pkgID] = latestVersionString
			// Actually save the metadata
			output.PrintVerboseStart("Saving pending update to metadata")
			if err := pm.WritePackageManagerMetadata(metadata); err != nil {
				output.PrintVerboseError("Save pending update", err)
				return hasNewUpdate, "", fmt.Errorf("failed to save metadata: %v", err)
			}
			output.PrintVerboseComplete("Save pending update to metadata")
		} else {
			output.PrintVerboseDebug("UPDATE", "Pending update already exists: %s", updateVersion)
		}
		output.PrintVerboseComplete("Check for existing pending update")

		return hasNewUpdate, latestVersionString, nil
	} else {
		output.PrintVerboseDebug("VERSION", "No update needed - current version is up to date")
	}

	// No updates available
	return hasNewUpdate, latestVersionString, nil
}

// Helper function to parse version string to integer for comparison
func parseVersionToInt(version string) (int, error) {
	// Normalize version by removing all non-numeric characters from the beginning and end
	// This matches the normalization done in GetLatestVersionName()
	version = strings.TrimFunc(version, func(r rune) bool {
		return r != '.' && (r < '0' || r > '9')
	})

	// Remove pre-release and build metadata for comparison (everything after - or +)
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

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
		if part == "" {
			continue // Skip empty parts
		}
		num, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("failed to parse version component '%s': %v", part, err)
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
