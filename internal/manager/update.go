package manager

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tranquil-tr0/get/internal/github"
)

// UpdateAllPackages returns new updates for all installed packages and updates that could not be checked and updates Metadata
func (pm *PackageManager) UpdateAllPackages() (updates map[string]string, err error) {

	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}

	updates = make(map[string]string)
	var failedPackages []string

	for pkgID := range metadata.Packages {
		hasNewUpdate, latestVersionString, updateErr := pm.updatePackageAndReturnNewVersion(pkgID)
		if updateErr != nil {
			failedPackages = append(failedPackages, pkgID)
			continue
		}

		if hasNewUpdate {
			updates[pkgID] = latestVersionString
		}
	}

	if len(failedPackages) > 0 {
		return updates, fmt.Errorf("failed to check updates for %d package(s): %v", len(failedPackages), failedPackages)
	}

	return updates, nil
}

// updatePackageAndReturnNewVersion marks a package update if it hasn't already been marked
// and returns the new version that it has been marked to update to
func (pm *PackageManager) updatePackageAndReturnNewVersion(pkgID string) (hasNewUpdate bool, latestVersionString string, err error) {
	hasNewUpdate = false

	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to load metadata: %v", err)
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return hasNewUpdate, "", fmt.Errorf("package %s is not installed", pkgID)
	}

	currentVersion, err := parseVersionToInt(pkg.Version)
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to parse current version: %v", err)
	}

	var options *github.ReleaseOptions
	if pkg.TagPrefix != "" {
		options = &github.ReleaseOptions{
			TagPrefix: pkg.TagPrefix,
		}
	}

	if options != nil {
		latestVersionString, err = pm.GithubClient.GetLatestVersionNameWithOptions(pkgID, options)
	} else {
		latestVersionString, err = pm.GithubClient.GetLatestVersionName(pkgID)
	}
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to get latest version: %v", err)
	}

	latestVersionInt, err := parseVersionToInt(latestVersionString)
	if err != nil {
		return hasNewUpdate, "", fmt.Errorf("failed to parse latest version: %v", err)
	}

	if latestVersionInt > currentVersion {
		var latestRelease *github.Release
		if options != nil {
			latestRelease, err = pm.GithubClient.GetLatestReleaseWithOptions(pkgID, options)
		} else {
			latestRelease, err = pm.GithubClient.GetLatestRelease(pkgID)
		}
		if err != nil {
			return hasNewUpdate, latestVersionString, fmt.Errorf("failed to get latest release: %v", err)
		}

		if pkg.InstallType == "deb" {
			debPackage := latestRelease.FindDebPackage()
			if debPackage == nil {
				return hasNewUpdate, latestVersionString, fmt.Errorf("latest release does not contain a .deb file")
			}
		}

		updateVersion, err := pm.GetPendingUpdate(pkgID)
		if err != nil {
			return hasNewUpdate, latestVersionString, fmt.Errorf("error checking for existing updates: %s", err)
		}

		if updateVersion == "" {
			hasNewUpdate = true
			metadata.PendingUpdates[pkgID] = latestRelease.TagName
			if err := pm.WritePackageManagerMetadata(metadata); err != nil {
				return hasNewUpdate, "", fmt.Errorf("failed to save metadata: %v", err)
			}
		}

		return hasNewUpdate, latestVersionString, nil
	}

	return hasNewUpdate, latestVersionString, nil
}

// Helper function to parse version string to integer for comparison
func parseVersionToInt(version string) (int, error) {
	version = strings.TrimFunc(version, func(r rune) bool {
		return r != '.' && (r < '0' || r > '9')
	})

	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}

	var result int
	for i, part := range parts {
		if i >= 3 {
			break
		}
		if part == "" {
			continue
		}
		num, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("failed to parse version component '%s': %v", part, err)
		}
		switch i {
		case 0:
			result += num * 10000
		case 1:
			result += num * 100
		case 2:
			result += num
		}
	}

	return result, nil
}
