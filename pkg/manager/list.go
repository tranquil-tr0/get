package manager

import (
	"github.com/tranquil-tr0/get/pkg/output"
)

func (pm *PackageManager) PrintInstalledPackages() error {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return loadErr
	}

	if len(metadata.Packages) == 0 {
		output.PrintNormal("No packages are currently installed.")
		return nil
	} else {
		output.PrintTitle("Installed packages:")
		output.PrintTitle("----------------------------------")
	}

	for _, pkg := range metadata.Packages {
		output.PrintNormal(" %s/%s (Version: %s, Installed: %s)", output.Bold(pkg.Owner), output.Bold(pkg.Repo), pkg.Version, pkg.InstalledAt)
		if pkg.AptName != "" {
			output.PrintGreen("   └APT Package Name: %s", pkg.AptName)
		}
		output.PrintNormal("")
	}

	return nil
}
