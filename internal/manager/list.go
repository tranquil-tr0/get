package manager

func (pm *PackageManager) ListInstalledPackages() (map[string]PackageMetadata, error) {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return nil, loadErr
	}

	// TODO: return lists of packages based on/separated by install type

	return metadata.Packages, nil
}
