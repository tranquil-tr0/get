package manager

import (
	"fmt"
	"os/exec"

	"github.com/tranquil-tr0/get/internal/output"
)

func (pm *PackageManager) Remove(pkgID string) error {
	output.PrintVerboseStart("Loading package metadata for removal", pkgID)
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		output.PrintVerboseError("Load package metadata", loadErr)
		return loadErr
	}
	output.PrintVerboseComplete("Load package metadata")

	output.PrintVerboseStart("Checking if package is installed", pkgID)
	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		output.PrintVerboseError("Check package installation", fmt.Errorf("package not installed"))
		return fmt.Errorf("package %s is not installed", pkgID)
	}
	
	// Handle different installation types
	var displayName string
	if pkg.InstallType == "binary" {
		displayName = pkg.BinaryPath
	} else {
		displayName = pkg.AptName
	}
	output.PrintVerboseComplete("Check package installation", displayName)

	// Remove based on installation type
	switch pkg.InstallType {
	case "deb":
		if err := pm.RemoveDebPackage(pkg); err != nil {
			return fmt.Errorf("failed to remove .deb package: %v", err)
		}
	case "binary":
		if err := pm.RemoveBinaryPackage(pkg); err != nil {
			return fmt.Errorf("failed to remove binary package: %v", err)
		}
	default:
		// Legacy packages without InstallType - assume .deb
		if pkg.AptName == "" {
			output.PrintVerboseError("Validate package metadata", fmt.Errorf("missing apt package name"))
			return fmt.Errorf("package %s was installed without capturing the apt package name", pkgID)
		}
		if err := pm.RemoveDebPackage(pkg); err != nil {
			return fmt.Errorf("failed to remove legacy .deb package: %v", err)
		}
	}

	// Remove from both packages and pending updates
	output.PrintVerboseStart("Updating package metadata after removal")
	delete(metadata.Packages, pkgID)
	if _, hadUpdate := metadata.PendingUpdates[pkgID]; hadUpdate {
		output.PrintVerboseDebug("METADATA", "Removing pending update for %s", pkgID)
		delete(metadata.PendingUpdates, pkgID)
	}

	err := pm.WritePackageManagerMetadata(metadata)
	if err != nil {
		output.PrintVerboseError("Write package metadata", err)
		return err
	}
	output.PrintVerboseComplete("Update package metadata after removal")
	return nil
}

// RemoveDebPackage removes a .deb package using apt
func (pm *PackageManager) RemoveDebPackage(pkg PackageMetadata) error {
	if pkg.AptName == "" {
		return fmt.Errorf("missing apt package name for .deb removal")
	}

	output.PrintVerboseStart("Removing .deb package with apt", pkg.AptName)
	cmd := exec.Command("sudo", "apt", "remove", "-y", pkg.AptName)
	output.PrintVerboseDebug("APT", "Command: %v", cmd.Args)
	cmdOutput, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		output.PrintVerboseError("Remove package with apt", cmdErr)
		output.PrintVerboseDebug("APT", "Output: %s", string(cmdOutput))
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", cmdErr, cmdOutput)
	}
	output.PrintVerboseComplete("Remove .deb package with apt")
	output.PrintVerboseDebug("APT", "Removal output: %s", string(cmdOutput))
	return nil
}

// RemoveBinaryPackage removes a binary executable
func (pm *PackageManager) RemoveBinaryPackage(pkg PackageMetadata) error {
	if pkg.BinaryPath == "" {
		return fmt.Errorf("missing binary path for binary removal")
	}

	output.PrintVerboseStart("Removing binary executable", pkg.BinaryPath)
	cmd := exec.Command("sudo", "rm", "-f", pkg.BinaryPath)
	output.PrintVerboseDebug("REMOVE", "Command: %v", cmd.Args)
	cmdOutput, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		output.PrintVerboseError("Remove binary executable", cmdErr)
		output.PrintVerboseDebug("REMOVE", "Output: %s", string(cmdOutput))
		return fmt.Errorf("failed to remove binary: %v\nOutput: %s", cmdErr, cmdOutput)
	}
	output.PrintVerboseComplete("Remove binary executable", pkg.BinaryPath)
	output.PrintVerboseDebug("REMOVE", "Removal output: %s", string(cmdOutput))
	return nil
}
