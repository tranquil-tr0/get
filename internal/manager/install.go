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

	// Install with apt
	fmt.Println("Installing with apt...")
	output.PrintVerboseStart("Installing package with apt", packagePath)
	cmd := exec.Command("sudo", "-p", "[get] Password required for package installation: ", "apt", "install", "-y", packagePath)
	output.PrintVerboseDebug("APT", "Command: %v", cmd.Args)
	cmdReader, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		output.PrintVerboseError("Create apt output pipe", pipeErr)
		return fmt.Errorf("failed to create output pipe: %v", pipeErr)
	}
	cmd.Stderr = cmd.Stdout

	if startErr := cmd.Start(); startErr != nil {
		output.PrintVerboseError("Start apt installation", startErr)
		return fmt.Errorf("failed to start installation: %v", startErr)
	}

	// Capture output for package name
	output.PrintVerboseDebug("APT", "Capturing installation output")
	var outputBuilder strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := cmdReader.Read(buf)
		if n > 0 {
			outputBuilder.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			output.PrintVerboseError("Read apt output", err)
			return fmt.Errorf("error reading output: %v", err)
		}
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		output.PrintVerboseError("Wait for apt installation", waitErr)
		return fmt.Errorf("installation failed: %v", waitErr)
	}
	output.PrintVerboseComplete("Install package with apt")

	// Extract apt package name
	output.PrintVerboseStart("Extracting package name from apt output")
	outputStr := outputBuilder.String()
	output.PrintVerboseDebug("APT", "Parsing output (%d lines)", len(strings.Split(outputStr, "\n")))
	lines := strings.Split(outputStr, "\n")
	var aptPackageName string
	var prevLine string

	for _, line := range lines {
		if strings.TrimSpace(prevLine) == "Installing:" || strings.TrimSpace(prevLine) == "Upgrading:" {
			aptPackageName = strings.TrimSpace(line)
			break
		}
		prevLine = line
	}

	if aptPackageName == "" {
		output.PrintVerboseError("Extract package name", fmt.Errorf("package name not found in apt output"))
		return fmt.Errorf("failed to find package name in apt output")
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
		return err
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
