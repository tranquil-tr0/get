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

	"github.com/tranquil-tr0/get/pkg/github"
)

func (pm *PackageManager) InstallRelease(pkgID string, release *github.Release) error {
	debPackage := release.FindDebPackage()
	if debPackage == nil {
		return fmt.Errorf("no .deb package found in release")
	}

	// Download package
	resp, httpErr := http.Get(debPackage.BrowserDownloadURL)
	if httpErr != nil {
		return fmt.Errorf("failed to download package: %v", httpErr)
	}
	defer resp.Body.Close()

	// FIXME: use system tempdir, not weird workaround
	// Create temp directory
	tempDir, tempErr := os.MkdirTemp("", "get-*")
	if tempErr != nil {
		return fmt.Errorf("failed to create temp directory: %v", tempErr)
	}
	defer os.RemoveAll(tempDir)

	// Save package
	packagePath := filepath.Join(tempDir, debPackage.Name)
	file, createErr := os.Create(packagePath)
	if createErr != nil {
		return fmt.Errorf("failed to create package file: %v", createErr)
	}

	if _, copyErr := io.Copy(file, resp.Body); copyErr != nil {
		file.Close()
		return fmt.Errorf("failed to save package file: %v", copyErr)
	}
	file.Close()

	// Install with apt
	fmt.Println("Installing with apt...")
	// TODO: use -p in sudo for custom prompt
	cmd := exec.Command("sudo", "apt", "install", "-y", packagePath)
	cmdReader, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		return fmt.Errorf("failed to create output pipe: %v", pipeErr)
	}
	cmd.Stderr = cmd.Stdout

	if startErr := cmd.Start(); startErr != nil {
		return fmt.Errorf("failed to start installation: %v", startErr)
	}

	// Capture output for package name
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
			return fmt.Errorf("error reading output: %v", err)
		}
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		return fmt.Errorf("installation failed: %v", waitErr)
	}

	// FIXME: use system tempdir, not weird workaround
	os.RemoveAll(tempDir)
	// Extract apt package name
	outputStr := outputBuilder.String()
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
		return fmt.Errorf("failed to find package name in apt output")
	}

	// FIXME: all below code is bad. the metadata entirely overwrites any existing packages records.
	// Update metadata
	metadata, metaErr := pm.GetPackageManagerMetadata()
	if metaErr != nil {
		return metaErr
	}

	var owner, repo string
	parts := strings.Split(pkgID, "/")
	if len(parts) >= 2 {
		owner, repo = parts[0], parts[1]
		// handle the owner and repo here
	} else {
		return fmt.Errorf("failed to find owner and repo from pkgID") // Handle the case where the split does not produce two parts
	}

	metadata.Packages[pkgID] = PackageMetadata{
		Owner:       owner,
		Repo:        repo,
		Version:     release.TagName,
		InstalledAt: release.PublishedAt,
		AptName:     aptPackageName,
	}

	return pm.WritePackageManagerMetadata(metadata)
}

// Install does InstallRelease, but an additional version and already installed sanity check
func (pm *PackageManager) Install(pkgID string, version string) error {
	// Load metadata
	metadata, metaErr := pm.GetPackageManagerMetadata()
	if metaErr != nil {
		return metaErr
	}

	// Check if package is already installed
	if _, exists := metadata.Packages[pkgID]; exists {
		return fmt.Errorf("package %s is already installed", pkgID)
	}

	// check if version specified, and fetches latest by default or specified version
	var release *github.Release
	var err error
	if version == "" {
		release, err = pm.GithubClient.GetLatestRelease(pkgID)
	} else {
		release, err = pm.GithubClient.GetReleaseByTag(pkgID, version)
	}
	if err != nil {
		return fmt.Errorf("error fetching latest release: %s", err)
	}

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release)
}

// InstallVersion does InstallRelease after fetching the release based on version
func (pm *PackageManager) InstallVersion(pkgID string, version string) error {
	// get the Release
	release, err := pm.GithubClient.GetReleaseByTag(pkgID, version)
	if err != nil {
		return fmt.Errorf("error fetching latest release: %s", err)
	}
	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release)
}
