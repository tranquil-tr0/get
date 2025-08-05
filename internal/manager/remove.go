package manager

import (
	"fmt"
	"os"
)

func (pm *PackageManager) Remove(pkgID string) error {
	metadata, loadErr := pm.GetPackageManagerMetadata()
	if loadErr != nil {
		return loadErr
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return fmt.Errorf("package %s is not installed", pkgID)
	}

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
			return fmt.Errorf("package %s was installed without capturing the apt package name", pkgID)
		}
		if err := pm.RemoveDebPackage(pkg); err != nil {
			return fmt.Errorf("failed to remove legacy .deb package: %v", err)
		}
	}

	// Remove from both packages and pending updates
	delete(metadata.Packages, pkgID)
	delete(metadata.PendingUpdates, pkgID)

	err := pm.WritePackageManagerMetadata(metadata)
	if err != nil {
		return err
	}
	return nil
}

// RemoveDebPackage removes a .deb package using apt
func (pm *PackageManager) RemoveDebPackage(pkg PackageMetadata) error {
	if pkg.AptName == "" {
		return fmt.Errorf("missing apt package name for .deb removal")
	}

	cmdOutput, cmdErr := pm.Out.PromptElevatedCommand("Password required for package removal: ", "apt", "remove", "-y", pkg.AptName)
	if cmdErr != nil {
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", cmdErr, cmdOutput)
	}
	return nil
}

// RemoveBinaryPackage removes a binary executable
func (pm *PackageManager) RemoveBinaryPackage(pkg PackageMetadata) error {
	if pkg.BinaryPath == "" {
		return fmt.Errorf("missing binary path for binary removal")
	}

	if err := os.Remove(pkg.BinaryPath); err != nil {
		return fmt.Errorf("failed to remove binary %s: %v", pkg.BinaryPath, err)
	}
	return nil
}
