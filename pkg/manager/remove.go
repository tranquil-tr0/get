package manager

import (
	"fmt"
	"os/exec"
)

func (pm *PackageManager) Remove(owner, repo string) error {
	metadata, loadErr := pm.loadMetadata()
	if loadErr != nil {
		return loadErr
	}

	packageKey := fmt.Sprintf("%s/%s", owner, repo)
	pkg, exists := metadata.Packages[packageKey]
	if !exists {
		return fmt.Errorf("package %s is not installed", packageKey)
	}

	if pkg.AptName == "" {
		return fmt.Errorf("package %s was installed without capturing the apt package name", packageKey)
	}

	// Remove the package using apt
	cmd := exec.Command("sudo", "apt", "remove", "-y", pkg.AptName)
	cmdOutput, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", cmdErr, cmdOutput)
	}

	delete(metadata.Packages, packageKey)
	return pm.saveMetadata(metadata)
}