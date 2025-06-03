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
	output.PrintVerboseComplete("Check package installation", pkg.AptName)

	if pkg.AptName == "" {
		output.PrintVerboseError("Validate package metadata", fmt.Errorf("missing apt package name"))
		return fmt.Errorf("package %s was installed without capturing the apt package name", pkgID)
	}

	// Remove the package using apt
	output.PrintVerboseStart("Removing package with apt", pkg.AptName)
	cmd := exec.Command("sudo", "apt", "remove", "-y", pkg.AptName)
	output.PrintVerboseDebug("APT", "Command: %v", cmd.Args)
	cmdOutput, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		output.PrintVerboseError("Remove package with apt", cmdErr)
		output.PrintVerboseDebug("APT", "Output: %s", string(cmdOutput))
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", cmdErr, cmdOutput)
	}
	output.PrintVerboseComplete("Remove package with apt")
	output.PrintVerboseDebug("APT", "Removal output: %s", string(cmdOutput))

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
