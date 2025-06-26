package manager

import (
	"fmt"
	"strings"

	"github.com/tranquil-tr0/get/internal/output"
)

func (pm *PackageManager) PrintInstalledPackages() error {
	output.PrintVerboseStart("Loading package metadata for listing")
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		output.PrintVerboseError("Load package metadata", loadErr)
		return loadErr
	}
	output.PrintVerboseComplete("Load package metadata", fmt.Sprintf("%d packages", len(metadata.Packages)))

	if len(metadata.Packages) == 0 {
		output.PrintVerboseDebug("LIST", "No packages found in metadata")
		output.PrintNormal("No packages are currently installed.")
		return nil
	} else {
		output.PrintVerboseDebug("LIST", "Displaying %d installed packages", len(metadata.Packages))
		output.PrintTitle("Installed packages:")
		output.PrintTitle("----------------------------------")
	}

	output.PrintVerboseStart("Formatting package list display")
	for pkgID, pkg := range metadata.Packages {
		output.PrintVerboseDebug("LIST", "Processing package: %s", pkgID)
		parts := strings.Split(pkgID, "/")
		var owner, repo string
		if len(parts) >= 2 {
			owner, repo = parts[0], parts[1]
		} else {
			owner, repo = pkgID, ""
			output.PrintVerboseDebug("LIST", "Warning: package ID format unusual: %s", pkgID)
		}
		
		output.PrintNormal(" %s/%s (Version: %s, Installed: %s)", output.Bold(owner), output.Bold(repo), pkg.Version, pkg.InstalledAt)
		
		// Display installation type and details
		switch pkg.InstallType {
		case "deb":
			output.PrintGreen("   └Type: .deb package")
			if pkg.AptName != "" {
				output.PrintGreen("   └APT Package Name: %s", pkg.AptName)
				output.PrintVerboseDebug("LIST", "APT package name: %s", pkg.AptName)
			}
		case "binary":
			output.PrintYellow("   └Type: Binary executable")
			if pkg.BinaryPath != "" {
				output.PrintYellow("   └Binary Path: %s", pkg.BinaryPath)
				output.PrintVerboseDebug("LIST", "Binary path: %s", pkg.BinaryPath)
			}
		default:
			// Legacy packages without InstallType - assume .deb
			output.PrintGreen("   └Type: .deb package (legacy)")
			if pkg.AptName != "" {
				output.PrintGreen("   └APT Package Name: %s", pkg.AptName)
				output.PrintVerboseDebug("LIST", "APT package name: %s", pkg.AptName)
			} else {
				output.PrintVerboseDebug("LIST", "Warning: no APT package name for legacy package %s", pkgID)
			}
		}
		
		// Show original asset name if available
		if pkg.OriginalName != "" {
			output.PrintNormal("   └Original Asset: %s", pkg.OriginalName)
		}
		
		output.PrintNormal("")
	}
	output.PrintVerboseComplete("Format package list display")

	return nil
}
