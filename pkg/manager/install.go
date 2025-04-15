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

func (pm *PackageManager) InstallRelease(owner, repo string, release *github.Release) error {
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

	// Update metadata
	metadata, metaErr := pm.LoadMetadata()
	if metaErr != nil {
		return metaErr
	}

	metadata.Packages[fmt.Sprintf("%s/%s", owner, repo)] = PackageMetadata{
		Owner:       owner,
		Repo:        repo,
		Version:     release.TagName,
		InstalledAt: release.PublishedAt,
		AptName:     aptPackageName,
	}

	return pm.SaveMetadata(metadata)
}

// Install does InstallRelease, but an extra version sanity check first
func (pm *PackageManager) Install(owner, repo string, version string) error {
	// Load metadata
	metadata, metaErr := pm.LoadMetadata()
	if metaErr != nil {
		return metaErr
	}

	// Check if package is already installed
	packageKey := fmt.Sprintf("%s/%s", owner, repo)
	if _, exists := metadata.Packages[packageKey]; exists {
		return fmt.Errorf("package %s is already installed", packageKey)
	}

	var release *github.Release
	var releaseErr error

	if version == "" {
		release, releaseErr = pm.GithubClient.GetLatestRelease(owner, repo)
	} else {
		release, releaseErr = pm.GithubClient.GetReleaseByTag(owner, repo, version)
	}
	if releaseErr != nil {
		return releaseErr
	}

	return pm.InstallRelease(owner, repo, release)
}
