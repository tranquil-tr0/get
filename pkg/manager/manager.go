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
	MetadataPath string
	GithubClient *github.Client
	Verbose      bool
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

	if _, err := os.Stat(pm.MetadataPath); os.IsNotExist(err) {
		return metadata, nil
	}

	data, err := os.ReadFile(pm.MetadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	if err := json.Unmarshal(data, metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	return metadata, nil
}

func (pm *PackageManager) saveMetadata(metadata *Metadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(pm.MetadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	return nil
}

func (pm *PackageManager) ListPackages() ([]PackageMetadata, error) {
	metadata, err := pm.loadMetadata()
	if err != nil {
		return nil, err
	}

	packages := make([]PackageMetadata, 0, len(metadata.Packages))
	for _, pkg := range metadata.Packages {
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (pm *PackageManager) Install(owner, repo string, version string) error {
	var release *github.Release
	var err error

	if version == "" {
		release, err = pm.GithubClient.GetLatestRelease(owner, repo)
	} else {
		release, err = pm.GithubClient.GetReleaseByTag(owner, repo, version)
	}
	if err != nil {
		return err
	}

	debPackage := release.FindDebPackage()
	if debPackage == nil {
		return fmt.Errorf("no .deb package found in the latest release")
	}

	// Download the .deb package
	resp, err := http.Get(debPackage.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download package: %v", err)
	}
	defer resp.Body.Close()

	tempDir, err := os.MkdirTemp("", "get-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	packagePath := filepath.Join(tempDir, debPackage.Name)
	file, err := os.Create(packagePath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %v", err)
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		return fmt.Errorf("failed to save package file: %v", err)
	}
	file.Close()

	// Install the package using apt
	cmd := exec.Command("sudo", "apt", "install", "-y", packagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install package: %v\nOutput: %s", err, output)
	}

	// Extract package name from apt output
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	var aptPackageName string
	for _, line := range lines {
		if strings.HasPrefix(line, "Note, selecting '") {
			parts := strings.Split(line, "' instead of")
			if len(parts) >= 1 {
				nameParts := strings.Split(parts[0], "Note, selecting '")
				if len(nameParts) >= 2 {
					aptPackageName = nameParts[1]
					break
				}
			}
		}
	}

	if aptPackageName == "" {
		return fmt.Errorf("failed to extract package name from apt output")
	}

	// Update metadata
	metadata, err := pm.loadMetadata()
	if err != nil {
		return err
	}

	metadata.Packages[fmt.Sprintf("%s/%s", owner, repo)] = PackageMetadata{
		Owner:       owner,
		Repo:        repo,
		Version:     release.TagName,
		InstalledAt: release.PublishedAt,
		AptName:     aptPackageName,
	}

	return pm.saveMetadata(metadata)
}

func (pm *PackageManager) Remove(owner, repo string) error {
	metadata, err := pm.loadMetadata()
	if err != nil {
		return err
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove package: %v\nOutput: %s", err, output)
	}

	delete(metadata.Packages, packageKey)
	return pm.saveMetadata(metadata)
}

func (pm *PackageManager) Update(owner, repo string) error {
	metadata, err := pm.loadMetadata()
	if err != nil {
		return err
	}

	packageKey := fmt.Sprintf("%s/%s", owner, repo)
	pkg, exists := metadata.Packages[packageKey]
	if !exists {
		return fmt.Errorf("package %s is not installed", packageKey)
	}

	release, err := pm.GithubClient.GetLatestRelease(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to check updates for %s: %v", packageKey, err)
	}

	if release.TagName == pkg.Version {
		fmt.Printf("Package %s is already up to date\n", packageKey)
		return nil
	}

	// Remove old version and install new version
	if err := pm.Remove(owner, repo); err != nil {
		return err
	}

	return pm.Install(owner, repo, "")
}

func (pm *PackageManager) UpdateAll() error {
	metadata, err := pm.loadMetadata()
	if err != nil {
		return err
	}

	updateErrors := make([]string, 0)
	for packageKey, pkg := range metadata.Packages {
		release, err := pm.GithubClient.GetLatestRelease(pkg.Owner, pkg.Repo)
		if err != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to check updates for %s: %v", packageKey, err))
			continue
		}

		if release.TagName == pkg.Version {
			fmt.Printf("Package %s is already up to date\n", packageKey)
			continue
		}

		// Remove old version and install new version
		if err := pm.Remove(pkg.Owner, pkg.Repo); err != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to remove old version of %s: %v", packageKey, err))
			continue
		}

		if err := pm.Install(pkg.Owner, pkg.Repo, ""); err != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("failed to install new version of %s: %v", packageKey, err))
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
