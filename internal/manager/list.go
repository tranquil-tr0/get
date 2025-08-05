package manager

import (
	"sort"
)

func (pm *PackageManager) ListInstalledPackages() (map[string]PackageMetadata, error) {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return nil, loadErr
	}

	// Extract and sort keys
	keys := make([]string, 0, len(metadata.Packages))
	for k := range metadata.Packages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// TODO: find out what kind of sort this is exactly

	// Collect packages in sorted order into a new map
	sortedMap := make(map[string]PackageMetadata, len(keys))
	for _, k := range keys {
		sortedMap[k] = metadata.Packages[k]
	}
	return sortedMap, nil
}

func (pm *PackageManager) ListInstalledPackagesAndPendingUpdates() (map[string]PackageMetadata, map[string]string, error) {
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return nil, nil, err
	}

	// Extract and sort keys
	keys := make([]string, 0, len(metadata.Packages))
	for k := range metadata.Packages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Collect packages in sorted order into a new map
	sortedMap := make(map[string]PackageMetadata, len(keys))
	for _, k := range keys {
		sortedMap[k] = metadata.Packages[k]
	}
	return sortedMap, metadata.PendingUpdates, nil
}
