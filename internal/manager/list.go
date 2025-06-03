package manager

import (
	"strings"

	"github.com/tranquil-tr0/get/internal/output"
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

	for pkgID, pkg := range metadata.Packages {
		parts := strings.Split(pkgID, "/")
		var owner, repo string
		if len(parts) >= 2 {
			owner, repo = parts[0], parts[1]
		} else {
			owner, repo = pkgID, ""
		}
		output.PrintNormal(" %s/%s (Version: %s, Installed: %s)", output.Bold(owner), output.Bold(repo), pkg.Version, pkg.InstalledAt)
		if pkg.AptName != "" {
			output.PrintGreen("   └APT Package Name: %s", pkg.AptName)
		}
		output.PrintNormal("")
	}

	return nil
}
