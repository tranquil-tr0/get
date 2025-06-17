/*
 * TODO: if more than one deb package do interactive prompt, best implemented in Install()
 */
package manager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/output"
)

func (pm *PackageManager) InstallRelease(pkgID string, release *github.Release) error {
	output.PrintVerboseStart("Finding .deb package in release", release.TagName)
	debPackage := release.FindDebPackage()
	if debPackage == nil {
		output.PrintVerboseError("Find .deb package", fmt.Errorf("no .deb package found"))
		return fmt.Errorf("no .deb package found in release")
	}
	output.PrintVerboseComplete("Find .deb package", debPackage.Name)

	// Download package
	output.PrintVerboseStart("Downloading package", debPackage.BrowserDownloadURL)
	resp, httpErr := http.Get(debPackage.BrowserDownloadURL)
	if httpErr != nil {
		output.PrintVerboseError("Download package", httpErr)
		return fmt.Errorf("failed to download package: %v", httpErr)
	}
	defer resp.Body.Close()
	output.PrintVerboseDebug("HTTP", "Download response status: %s", resp.Status)

	// FIXME: use system tempdir, not weird workaround
	// Create temp directory
	output.PrintVerboseStart("Creating temporary directory")
	tempDir, tempErr := os.MkdirTemp("", "get-*")
	if tempErr != nil {
		output.PrintVerboseError("Create temporary directory", tempErr)
		return fmt.Errorf("failed to create temp directory: %v", tempErr)
	}
	defer os.RemoveAll(tempDir)
	output.PrintVerboseComplete("Create temporary directory", tempDir)

	// Save package
	packagePath := filepath.Join(tempDir, debPackage.Name)
	output.PrintVerboseStart("Saving package file", packagePath)
	file, createErr := os.Create(packagePath)
	if createErr != nil {
		output.PrintVerboseError("Create package file", createErr)
		return fmt.Errorf("failed to create package file: %v", createErr)
	}

	if _, copyErr := io.Copy(file, resp.Body); copyErr != nil {
		file.Close()
		output.PrintVerboseError("Save package file", copyErr)
		return fmt.Errorf("failed to save package file: %v", copyErr)
	}
	file.Close()
	output.PrintVerboseComplete("Save package file", packagePath)

	// Validate package before installation
	if err := pm.ValidateDebPackage(packagePath); err != nil {
		return fmt.Errorf("package validation failed: %v", err)
	}

	// Install with dpkg (more direct than apt)
	fmt.Println("Installing with dpkg...")
	output.PrintVerboseStart("Installing package with dpkg", packagePath)
	cmd := exec.Command("sudo", "-p", "[get] Password required for package installation: ", "dpkg", "-i", packagePath)
	output.PrintVerboseDebug("DPKG", "Command: %v", cmd.Args)

	// Run dpkg installation
	cmdOutput, dpkgErr := cmd.CombinedOutput()
	output.PrintVerboseDebug("DPKG", "Installation output: %s", string(cmdOutput))

	if dpkgErr != nil {
		// If dpkg fails due to missing dependencies, try to fix with apt
		if strings.Contains(string(cmdOutput), "dependency problems") {
			output.PrintVerboseStart("Fixing dependency issues with apt")
			fixCmd := exec.Command("sudo", "apt", "-f", "install", "-y")
			fixOutput, fixErr := fixCmd.CombinedOutput()
			output.PrintVerboseDebug("APT", "Dependency fix output: %s", string(fixOutput))
			if fixErr != nil {
				output.PrintVerboseError("Fix dependencies", fixErr)
				return fmt.Errorf("failed to fix dependencies: %v\nOutput: %s", fixErr, fixOutput)
			}
			output.PrintVerboseComplete("Fix dependency issues")
		} else {
			output.PrintVerboseError("Install package with dpkg", dpkgErr)
			return fmt.Errorf("dpkg installation failed: %v\nOutput: %s", dpkgErr, cmdOutput)
		}
	}
	output.PrintVerboseComplete("Install package with dpkg")

	// Extract package name using dpkg-deb (most reliable method)
	output.PrintVerboseStart("Extracting package name")
	aptPackageName, nameErr := pm.GetPackageNameFromDeb(packagePath)
	if nameErr != nil {
		// Fallback: extract from .deb filename
		output.PrintVerboseDebug("DPKG", "Falling back to extracting from .deb filename")
		debFilename := filepath.Base(packagePath)
		if strings.HasSuffix(debFilename, ".deb") {
			nameWithoutExt := strings.TrimSuffix(debFilename, ".deb")
			// Common pattern: package-name_version_arch.deb
			parts := strings.Split(nameWithoutExt, "_")
			if len(parts) > 0 {
				aptPackageName = parts[0]
			}
		}
		
		if aptPackageName == "" {
			output.PrintVerboseError("Extract package name", nameErr)
			return fmt.Errorf("failed to extract package name: %v", nameErr)
		}
	}
	output.PrintVerboseComplete("Extract package name", aptPackageName)

	// Update metadata - reload to avoid overwriting other changes
	output.PrintVerboseStart("Updating package metadata")
	metadata, metaErr := pm.GetPackageManagerMetadata()
	if metaErr != nil {
		output.PrintVerboseError("Load package metadata", metaErr)
		return metaErr
	}

	parts := strings.Split(pkgID, "/")
	if len(parts) < 2 {
		output.PrintVerboseError("Parse package ID", fmt.Errorf("invalid pkgID format: %s", pkgID))
		return fmt.Errorf("failed to find owner and repo from pkgID: %s", pkgID)
	}

	normalizedVersion := strings.TrimPrefix(release.TagName, "v")
	output.PrintVerboseDebug("METADATA", "Adding package: %s version %s", pkgID, normalizedVersion)
	metadata.Packages[pkgID] = PackageMetadata{
		Version:     normalizedVersion, // Normalize version
		InstalledAt: release.PublishedAt,
		AptName:     aptPackageName,
	}

	// Remove from pending updates if it exists
	if _, hadUpdate := metadata.PendingUpdates[pkgID]; hadUpdate {
		output.PrintVerboseDebug("METADATA", "Removing pending update for %s", pkgID)
		delete(metadata.PendingUpdates, pkgID)
	}

	output.PrintVerboseStart("Writing package metadata to disk")
	err := pm.WritePackageManagerMetadata(metadata)
	if err != nil {
		output.PrintVerboseError("Write package metadata", err)
		// Attempt rollback if metadata write fails
		output.PrintVerboseStart("Attempting rollback due to metadata write failure")
		if rollbackErr := pm.RollbackInstallation(aptPackageName); rollbackErr != nil {
			output.PrintVerboseError("Rollback installation", rollbackErr)
			return fmt.Errorf("installation succeeded but metadata write failed, and rollback also failed: %v (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("installation succeeded but metadata write failed (package was rolled back): %v", err)
	}
	output.PrintVerboseComplete("Update package metadata")
	return nil
}

// Install does InstallRelease, but an additional version and already installed sanity check
func (pm *PackageManager) Install(pkgID string, version string) error {
	// Load metadata
	output.PrintVerboseStart("Loading package metadata for installation check")
	metadata, metaErr := pm.GetPackageManagerMetadata()
	if metaErr != nil {
		output.PrintVerboseError("Load package metadata", metaErr)
		return metaErr
	}
	output.PrintVerboseComplete("Load package metadata")

	// Check if package is already installed
	output.PrintVerboseStart("Checking if package is already installed", pkgID)
	if _, exists := metadata.Packages[pkgID]; exists {
		output.PrintVerboseError("Check package installation", fmt.Errorf("package already installed"))
		return fmt.Errorf("package %s is already installed", pkgID)
	}
	output.PrintVerboseComplete("Check package installation", "not installed")

	// check if version specified, and fetches latest by default or specified version
	var release *github.Release
	var err error
	if version == "" {
		output.PrintVerboseStart("Fetching latest release from GitHub", pkgID)
		release, err = pm.GithubClient.GetLatestRelease(pkgID)
	} else {
		output.PrintVerboseStart("Fetching specific release from GitHub", fmt.Sprintf("%s@%s", pkgID, version))
		release, err = pm.GithubClient.GetReleaseByTag(pkgID, version)
	}
	if err != nil {
		output.PrintVerboseError("Fetch GitHub release", err)
		return fmt.Errorf("error fetching latest release: %s", err)
	}
	output.PrintVerboseComplete("Fetch GitHub release", release.TagName)

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release)
}

// InstallVersion does InstallRelease after fetching the release based on version
func (pm *PackageManager) InstallVersion(pkgID string, version string) error {
	// get the Release
	output.PrintVerboseStart("Fetching specific release for installation", fmt.Sprintf("%s@%s", pkgID, version))
	release, err := pm.GithubClient.GetReleaseByTag(pkgID, version)
	if err != nil {
		output.PrintVerboseError("Fetch GitHub release", err)
		return fmt.Errorf("error fetching latest release: %s", err)
	}
	output.PrintVerboseComplete("Fetch GitHub release", release.TagName)

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release)
}

// ValidateDebPackage validates a .deb package before installation
func (pm *PackageManager) ValidateDebPackage(packagePath string) error {
	output.PrintVerboseStart("Validating .deb package", packagePath)

	// Check if file exists and is readable
	if _, err := os.Stat(packagePath); err != nil {
		output.PrintVerboseError("Check package file", err)
		return fmt.Errorf("package file not accessible: %v", err)
	}

	// Use dpkg --info to validate the package
	cmd := exec.Command("dpkg", "--info", packagePath)
	output.PrintVerboseDebug("DPKG", "Validation command: %v", cmd.Args)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		output.PrintVerboseError("Validate package", err)
		output.PrintVerboseDebug("DPKG", "Validation output: %s", string(cmdOutput))
		return fmt.Errorf("invalid .deb package: %v", err)
	}

	output.PrintVerboseComplete("Validate .deb package")
	output.PrintVerboseDebug("DPKG", "Package info: %s", string(cmdOutput))
	return nil
}

// GetPackageNameFromDeb extracts package name from .deb file using dpkg-deb
func (pm *PackageManager) GetPackageNameFromDeb(packagePath string) (string, error) {
	output.PrintVerboseStart("Extracting package name using dpkg-deb", packagePath)

	// Use dpkg-deb to get package name reliably
	cmd := exec.Command("dpkg-deb", "--field", packagePath, "Package")
	output.PrintVerboseDebug("DPKG", "Package name extraction command: %v", cmd.Args)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		output.PrintVerboseError("Extract package name with dpkg-deb", err)
		return "", fmt.Errorf("failed to extract package name: %v", err)
	}

	packageName := strings.TrimSpace(string(cmdOutput))
	if packageName == "" {
		return "", fmt.Errorf("empty package name extracted")
	}

	output.PrintVerboseComplete("Extract package name using dpkg-deb", packageName)
	return packageName, nil
}

// RollbackInstallation removes a package if installation metadata update fails
func (pm *PackageManager) RollbackInstallation(packageName string) error {
	output.PrintVerboseStart("Rolling back installation", packageName)
	
	cmd := exec.Command("sudo", "dpkg", "--remove", packageName)
	output.PrintVerboseDebug("DPKG", "Rollback command: %v", cmd.Args)
	
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		output.PrintVerboseError("Rollback installation", err)
		output.PrintVerboseDebug("DPKG", "Rollback output: %s", string(cmdOutput))
		return fmt.Errorf("rollback failed: %v", err)
	}
	
	output.PrintVerboseComplete("Rollback installation", packageName)
	return nil
}
