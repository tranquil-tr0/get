package manager

import (
	"fmt"

	"github.com/tranquil-tr0/get/pkg/output"
)

func (pm *PackageManager) PrintInstalledPackages() ([]PackageMetadata, error) {
	metadata, loadErr := pm.loadMetadata()
	if loadErr != nil {
		return nil, loadErr
	}

	packages := make([]PackageMetadata, 0, len(metadata.Packages))
	for _, pkg := range metadata.Packages {
		packages = append(packages, pkg)
		fmt.Printf("Package: %s/%s (Version: %s)\n", pkg.Owner, pkg.Repo, pkg.Version)
		if pkg.AptName != "" {
			output.PrintGreen("APT Package: %s", pkg.AptName)
		}
		fmt.Println()
	}

	return packages, nil
}

func (pm *PackageManager) GetPackage(pkgID string) (*PackageMetadata, error) {
	metadata, err := pm.loadMetadata()
	if err != nil {
		return nil, err
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return nil, fmt.Errorf("package %s not found", pkgID)
	}
	return &pkg, nil
}
