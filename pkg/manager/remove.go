package manager

import (
	"fmt"
	"os/exec"
)

func (pm *PackageManager) Remove(pkgID string) error {
	metadata, loadErr := pm.LoadMetadata()
	if loadErr != nil {
		return loadErr
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return fmt.Errorf("package %s is not installed", pkgID)
	}

	if pkg.AptName == "" {
		return fmt.Errorf("package %s was installed without capturing the apt package name", pkgID)
	}

	// Remove the package using apt
	cmd := exec.Command("sudo", "apt", "remove", "-y", pkg.AptName)
	cmdOutput, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", cmdErr, cmdOutput)
	}

	delete(metadata.Packages, pkgID)
	return pm.SaveMetadata(metadata)
}
