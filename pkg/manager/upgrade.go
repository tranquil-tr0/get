package manager

func (pm *PackageManager) UpgradeAllPackages() error {
	panic("Upgrade not implemented")
	// IMPLEMENTATION:
	/*
		1. Load the get metadata using manager.go
		2. Check metadata for pending updates
		3. If there are pending updates, [for each package with pending update]
			1. Call UpdatePackage(PackageID)
		4. In addition to error logging already implemented, if there still exist pending updates, return an error
		5. If no errors, now return nil
	*/
}

func (pm *PackageManager) UpdatePackage(pkgID string) error {
	panic("UpdatePackage not implemented")
	// IMPLEMENTATION:
	/*
		1. If there are pending updates for the package identified by pkgID,
			1. call Install() in install.go to install the latest version of the package
			2. remove the package from pending updates in metadata
	*/
}
