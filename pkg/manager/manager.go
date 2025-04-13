package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tranquil-tr0/get/pkg/github"
)

type PackageManager struct {
	MetadataPath   string
	GithubClient   *github.Client
	Verbose        bool
	PendingUpdates map[string]github.Release // Tracks available updates
}

func (pm *PackageManager) Upgrade() any {
	panic("unimplemented")
}

func (pm *PackageManager) CheckForUpdates() any {
	panic("unimplemented")
}

type PackageMetadata struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	AptName     string `json:"apt_name"`
}

type Metadata struct {
	Packages map[string]PackageMetadata `json:"packages"`
}

func NewPackageManager(metadataPath string) *PackageManager {
	return &PackageManager{
		MetadataPath: metadataPath,
		GithubClient: github.NewClient(),
		Verbose:      false,
	}
}

func (pm *PackageManager) loadMetadata() (*Metadata, error) {
	metadata := &Metadata{Packages: make(map[string]PackageMetadata)}

	if _, statErr := os.Stat(pm.MetadataPath); os.IsNotExist(statErr) {
		return metadata, nil
	}

	data, readErr := os.ReadFile(pm.MetadataPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", readErr)
	}

	if jsonErr := json.Unmarshal(data, metadata); jsonErr != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", jsonErr)
	}

	return metadata, nil
}

func (pm *PackageManager) saveMetadata(metadata *Metadata) error {
	data, marshalErr := json.MarshalIndent(metadata, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal metadata: %v", marshalErr)
	}

	if writeErr := os.WriteFile(pm.MetadataPath, data, 0644); writeErr != nil {
		return fmt.Errorf("failed to write metadata file: %v", writeErr)
	}

	return nil
}

func (pm *PackageManager) ListPackages() ([]PackageMetadata, error) {
	metadata, loadErr := pm.loadMetadata()
	if loadErr != nil {
		return nil, loadErr
	}

	packages := make([]PackageMetadata, 0, len(metadata.Packages))
	for _, pkg := range metadata.Packages {
		packages = append(packages, pkg)
		fmt.Printf("Package: %s/%s (Version: %s)\n", pkg.Owner, pkg.Repo, pkg.Version)
		if pkg.AptName != "" {
			fmt.Printf("\033[32mAPT Package: %s\033[0m\n", pkg.AptName)
		}
		fmt.Println()
	}

	return packages, nil
}

func (pm *PackageManager) installRelease(owner, repo string, release *github.Release) error {
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
		if strings.TrimSpace(prevLine) == "Installing:" {
			aptPackageName = strings.TrimSpace(line)
			break
		}
		prevLine = line
	}

	if aptPackageName == "" {
		return fmt.Errorf("failed to find package name in apt output")
	}

	// Update metadata
	metadata, metaErr := pm.loadMetadata()
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

	return pm.saveMetadata(metadata)

	// Existing download and installation logic...
}

func (pm *PackageManager) Install(owner, repo string, version string) error {
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

	return pm.installRelease(owner, repo, release)
}

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

func (pm *PackageManager) Update(owner, repo string) error {
	metadata, statErr := pm.loadMetadata()
	if statErr != nil {
		return statErr
	}

	packageKey := fmt.Sprintf("%s/%s", owner, repo)
	pkg, exists := metadata.Packages[packageKey]
	if !exists {
		return fmt.Errorf("package %s is not installed", packageKey)
	}

	release, releaseErr := pm.GithubClient.GetLatestRelease(owner, repo)
	if releaseErr != nil {
		return fmt.Errorf("failed to check updates for %s: %v", packageKey, releaseErr)
	}

	if release.TagName == pkg.Version {
		fmt.Printf("Package %s is already up to date\n", packageKey)
		return nil
	}

	// Remove old version and install new version
	if removeErr := pm.Remove(owner, repo); removeErr != nil {
		return removeErr
	}

	return pm.Install(owner, repo, "")
}

func (pm *PackageManager) UpdateAll() error {
	metadata, loadErr := pm.loadMetadata()
	if loadErr != nil {
		return loadErr
	}

	updateErrors := make([]string, 0)
	for packageKey, pkg := range metadata.Packages {
		release, releaseErr := pm.GithubClient.GetLatestRelease(pkg.Owner, pkg.Repo)
		if releaseErr != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to check updates for %s: %v", packageKey, releaseErr))
			continue
		}

		if release.TagName == pkg.Version {
			fmt.Printf("Package %s is already up to date\n", packageKey)
			continue
		}

		// Remove old version and install new version
		if removeErr := pm.Remove(pkg.Owner, pkg.Repo); removeErr != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to remove old version of %s: %v", packageKey, removeErr))
			continue
		}

		if installErr := pm.Install(pkg.Owner, pkg.Repo, ""); installErr != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to install new version of %s: %v", packageKey, installErr))
			continue
		}

		fmt.Printf("Updated %s from %s to %s\n", packageKey, pkg.Version, release.TagName)
	}

	if len(updateErrors) > 0 {
		return fmt.Errorf("some packages failed to update:\n%s", strings.Join(updateErrors, "\n"))
	}

	return nil
}

func (pm *PackageManager) SetVerbose(verbose bool) {
	pm.Verbose = verbose
}
