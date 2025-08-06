package manager

import (
	"sort"
)

func (pm *PackageManager) ListInstalledPackages() ([]string, map[string]PackageMetadata, error) {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return nil, nil, loadErr
	}

	// Extract and sort sortedKeys
	sortedKeys := make([]string, 0, len(metadata.Packages))
	for k := range metadata.Packages {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	return sortedKeys, metadata.Packages, nil
}

func (pm *PackageManager) ListInstalledPackagesAndPendingUpdates() ([]string, map[string]PackageMetadata, map[string]string, error) {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return nil, nil, nil, loadErr
	}

	// Extract and sort keys
	sortedKeys := make([]string, 0, len(metadata.Packages))
	for k := range metadata.Packages {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	return sortedKeys, metadata.Packages, metadata.PendingUpdates, nil
}
