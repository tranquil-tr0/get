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
)

// SelectAssetInteractively prompts the user to select which asset to install
func (pm *PackageManager) SelectAssetInteractively(release *github.Release) (*github.Asset, string, error) {
	debPackages := release.FindDebPackages()
	binaryAssets := release.FindBinaryAssets()
	// Other assets: not deb or binary
	var otherAssets []github.Asset
	assetMap := make(map[string]struct{})
	for _, a := range debPackages {
		assetMap[a.Name] = struct{}{}
	}
	for _, a := range binaryAssets {
		assetMap[a.Name] = struct{}{}
	}
	for _, asset := range release.Assets {
		if _, found := assetMap[asset.Name]; !found {
			otherAssets = append(otherAssets, asset)
		}
	}

	// Prepare name slices for output selection
	debNames := make([]string, len(debPackages))
	for i, a := range debPackages {
		debNames[i] = a.Name
	}
	binaryNames := make([]string, len(binaryAssets))
	for i, a := range binaryAssets {
		binaryNames[i] = a.Name
	}
	otherNames := make([]string, len(otherAssets))
	for i, a := range otherAssets {
		otherNames[i] = a.Name
	}

	idx, err := pm.Out.PromptAssetIndexSelection(debNames, binaryNames, otherNames)
	if err != nil || idx < 0 {
		return nil, "", fmt.Errorf("asset selection invalid: %v", err)
	}

	var selectedAsset *github.Asset
	var typ string
	if idx < len(debPackages) {
		selectedAsset = &debPackages[idx]
		typ = "deb"
	} else if idx < len(debPackages)+len(binaryAssets) {
		selectedAsset = &binaryAssets[idx-len(debPackages)]
		typ = "binary"
	} else if idx < len(debPackages)+len(binaryAssets)+len(otherAssets) {
		selectedAsset = &otherAssets[idx-len(debPackages)-len(binaryAssets)]
		typ = "other"
	} else {
		return nil, "", fmt.Errorf("invalid asset index: %d", idx)
	}
	return selectedAsset, typ, nil
}

func (pm *PackageManager) InstallRelease(pkgID string, release *github.Release, preSelectedAsset *github.Asset) error {
	return pm.InstallReleaseWithOptions(pkgID, release, preSelectedAsset, nil)
}

func (pm *PackageManager) InstallReleaseWithOptions(pkgID string, release *github.Release, preSelectedAsset *github.Asset, options *github.ReleaseOptions) error {
	var selectedAsset *github.Asset
	var installType string
	var err error

	if preSelectedAsset != nil {
		selectedAsset = preSelectedAsset
		// Determine installType based on asset name or other properties if needed
		if strings.HasSuffix(selectedAsset.Name, ".deb") {
			installType = "deb"
		} else {
			installType = "binary"
		}
	} else {
		// Interactive asset selection
		pm.Out.PrintStatus("Please choose the correct asset to install. Your selection will be saved for future installations.")
		selectedAsset, installType, err = pm.SelectAssetInteractively(release)
		if err != nil {
			return fmt.Errorf("failed to select asset: %v", err)
		}
	}

	if preSelectedAsset == nil {
		// Save chosen asset
		metadata, metaErr := pm.GetPackageManagerMetadata()
		if metaErr != nil {
			return metaErr
		}
		pkgMetadata := metadata.Packages[pkgID]
		pkgMetadata.ChosenAsset = selectedAsset.Name
		metadata.Packages[pkgID] = pkgMetadata
		if err := pm.WritePackageManagerMetadata(metadata); err != nil {
			return err
		}
	}

	// Route to appropriate installation method
	switch installType {
	case "deb":
		return pm.InstallDebPackage(pkgID, release, selectedAsset, options)
	case "binary":
		return pm.InstallBinary(pkgID, release, selectedAsset, options)
	default:
		return fmt.Errorf("unsupported installation type: %s", installType)
	}
}

// InstallDebPackage handles .deb package installation
func (pm *PackageManager) InstallDebPackage(pkgID string, release *github.Release, debAsset *github.Asset, options *github.ReleaseOptions) error {
	// Download package
	pm.Out.PrintStatus("Downloading package: %s", debAsset.Name)
	resp, httpErr := http.Get(debAsset.BrowserDownloadURL)
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
	packagePath := filepath.Join(tempDir, debAsset.Name)
	file, createErr := os.Create(packagePath)
	if createErr != nil {
		return fmt.Errorf("failed to create package file: %v", createErr)
	}

	if _, copyErr := io.Copy(file, resp.Body); copyErr != nil {
		file.Close()
		return fmt.Errorf("failed to save package file: %v", copyErr)
	}
	file.Close()

	// Validate package before installation
	if err := pm.ValidateDebPackage(packagePath); err != nil {
		return fmt.Errorf("package validation failed: %v", err)
	}

	// Install with dpkg
	pm.Out.PrintStatus("Installing .deb package with dpkg...")
	cmd := exec.Command("sudo", "-p", "[get] Password required for package installation: ", "dpkg", "-i", packagePath)

	// Run dpkg installation
	cmdOutput, dpkgErr := cmd.CombinedOutput()

	if dpkgErr != nil {
		// If dpkg fails due to missing dependencies, try to fix with apt
		if strings.Contains(string(cmdOutput), "dependency problems") {
			fixCmd := exec.Command("sudo", "apt", "-f", "install", "-y")
			fixOutput, fixErr := fixCmd.CombinedOutput()
			if fixErr != nil {
				return fmt.Errorf("failed to fix dependencies: %v\nOutput: %s", fixErr, fixOutput)
			}
		} else {
			return fmt.Errorf("dpkg installation failed: %v\nOutput: %s", dpkgErr, cmdOutput)
		}
	}

	// Extract package name using dpkg-deb
	aptPackageName, nameErr := pm.GetPackageNameFromDeb(packagePath)
	if nameErr != nil {
		// Fallback: extract from .deb filename
		debFilename := filepath.Base(packagePath)
		if strings.HasSuffix(debFilename, ".deb") {
			nameWithoutExt := strings.TrimSuffix(debFilename, ".deb")
			parts := strings.Split(nameWithoutExt, "_")
			if len(parts) > 0 {
				aptPackageName = parts[0]
			}
		}

		if aptPackageName == "" {
			return fmt.Errorf("failed to extract package name: %v", nameErr)
		}
	}

	// Update metadata
	tagPrefix := ""
	if options != nil && options.TagPrefix != "" {
		tagPrefix = options.TagPrefix
	}
	return pm.UpdatePackageMetadata(pkgID, release, PackageMetadata{
		Version:      strings.TrimPrefix(release.TagName, "v"),
		InstalledAt:  release.PublishedAt,
		AptName:      aptPackageName,
		InstallType:  "deb",
		OriginalName: debAsset.Name,
		TagPrefix:    tagPrefix,
	})
}

// InstallBinary handles binary executable installation
func (pm *PackageManager) InstallBinary(pkgID string, release *github.Release, binaryAsset *github.Asset, options *github.ReleaseOptions) error {
	// Download binary
	pm.Out.PrintStatus("Downloading binary: %s", binaryAsset.Name)
	resp, httpErr := http.Get(binaryAsset.BrowserDownloadURL)
	if httpErr != nil {
		return fmt.Errorf("failed to download binary: %v", httpErr)
	}
	defer resp.Body.Close()

	// Create temp directory
	tempDir, tempErr := os.MkdirTemp("", "get-*")
	if tempErr != nil {
		return fmt.Errorf("failed to create temp directory: %v", tempErr)
	}
	defer os.RemoveAll(tempDir)

	// Save binary
	tempBinaryPath := filepath.Join(tempDir, binaryAsset.Name)
	file, createErr := os.Create(tempBinaryPath)
	if createErr != nil {
		return fmt.Errorf("failed to create binary file: %v", createErr)
	}

	if _, copyErr := io.Copy(file, resp.Body); copyErr != nil {
		file.Close()
		return fmt.Errorf("failed to save binary file: %v", copyErr)
	}
	file.Close()

	// Make binary executable
	if err := os.Chmod(tempBinaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %v", err)
	}

	// Determine final binary name and path
	binaryName := pm.GetBinaryName(pkgID, binaryAsset.Name)
	finalBinaryPath := filepath.Join("/usr/local/bin", binaryName)

	// Install binary to /usr/local/bin
	pm.Out.PrintStatus("Installing binary to /usr/local/bin...")

	// Check if we're upgrading the 'get' binary itself (to avoid "Text file busy" error)
	isSelfUpgrade := pkgID == "tranquil-tr0/get" && binaryName == "get"

	// Check if the binary already exists and we need special handling for self-upgrade
	backupPath := finalBinaryPath + ".old"
	binaryExists := false
	if isSelfUpgrade {
		if _, err := os.Stat(finalBinaryPath); err == nil {
			binaryExists = true

			// First, move the existing binary to a backup location to avoid "Text file busy" error
			mvCmd := exec.Command("sudo", "-p", "[get] Password required for binary installation: ", "mv", finalBinaryPath, backupPath)

			mvOutput, mvErr := mvCmd.CombinedOutput()
			if mvErr != nil {
				return fmt.Errorf("failed to backup existing binary: %v\nOutput: %s", mvErr, mvOutput)
			}
		}
	}

	// Copy the new binary
	cmd := exec.Command("sudo", "-p", "[get] Password required for binary installation: ", "cp", tempBinaryPath, finalBinaryPath)

	cmdOutput, installErr := cmd.CombinedOutput()
	if installErr != nil {

		// If installation failed and we had a backup (self-upgrade), try to restore it
		if isSelfUpgrade && binaryExists {
			restoreCmd := exec.Command("sudo", "mv", backupPath, finalBinaryPath)
			if restoreErr := restoreCmd.Run(); restoreErr != nil {
				return fmt.Errorf("failed to install binary and failed to restore backup: install error: %v, restore error: %v", installErr, restoreErr)
			}
		}

		return fmt.Errorf("failed to install binary: %v\nOutput: %s", installErr, cmdOutput)
	}

	// Installation successful, clean up backup if it exists (self-upgrade only)
	if isSelfUpgrade && binaryExists {
		rmCmd := exec.Command("sudo", "rm", backupPath)
		if rmErr := rmCmd.Run(); rmErr != nil {
			// Don't fail the installation if cleanup fails, just warn
			fmt.Printf("Warning: Failed to clean up backup file %s: %v\n", backupPath, rmErr)
		} else {
		}
	}

	pm.Out.PrintSuccess("Binary installed as: %s", binaryName)

	// Update metadata
	tagPrefix := ""
	if options != nil && options.TagPrefix != "" {
		tagPrefix = options.TagPrefix
	}
	return pm.UpdatePackageMetadata(pkgID, release, PackageMetadata{
		Version:      strings.TrimPrefix(release.TagName, "v"),
		InstalledAt:  release.PublishedAt,
		BinaryPath:   finalBinaryPath,
		InstallType:  "binary",
		OriginalName: binaryAsset.Name,
		TagPrefix:    tagPrefix,
	})
}

func (pm *PackageManager) Install(pkgID string, version string) error {
	return pm.InstallWithOptions(pkgID, version, nil)
}

// InstallWithOptions installs a package with additional options like tag prefix filtering
func (pm *PackageManager) InstallWithOptions(pkgID string, version string, options *github.ReleaseOptions) error {
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
		pm.Out.PrintStatus("Fetching latest release from GitHub...")
		if options != nil {
			release, err = pm.GithubClient.GetLatestReleaseWithOptions(pkgID, options)
		} else {
			release, err = pm.GithubClient.GetLatestRelease(pkgID)
		}
	} else {
		pm.Out.PrintStatus("Fetching release %s from GitHub...", version)
		if options != nil {
			release, err = pm.GithubClient.GetReleaseByTagWithOptions(pkgID, version, options)
		} else {
			release, err = pm.GithubClient.GetReleaseByTag(pkgID, version)
		}
	}
	if err != nil {
		return fmt.Errorf("error fetching latest release: %s", err)
	}

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallReleaseWithOptions(pkgID, release, nil, options)
}

// InstallVersion does InstallRelease after fetching the release based on version
func (pm *PackageManager) InstallVersion(pkgID string, version string, chosenAsset *github.Asset) error {
	// get the Release
	release, err := pm.GithubClient.GetReleaseByTag(pkgID, version)
	if err != nil {
		return fmt.Errorf("error fetching latest release: %s", err)
	}

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release, chosenAsset)
}

// ValidateDebPackage validates a .deb package before installation
func (pm *PackageManager) ValidateDebPackage(packagePath string) error {

	// Check if file exists and is readable
	if _, err := os.Stat(packagePath); err != nil {
		return fmt.Errorf("package file not accessible: %v", err)
	}

	// Use dpkg --info to validate the package
	cmd := exec.Command("dpkg", "--info", packagePath)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid .deb package: %v", err)
	}

	return nil
}

// GetPackageNameFromDeb extracts package name from .deb file using dpkg-deb
func (pm *PackageManager) GetPackageNameFromDeb(packagePath string) (string, error) {

	cmd := exec.Command("dpkg-deb", "--field", packagePath, "Package")

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to extract package name: %v", err)
	}

	packageName := strings.TrimSpace(string(cmdOutput))
	if packageName == "" {
		return "", fmt.Errorf("empty package name extracted")
	}

	return packageName, nil
}

// RollbackInstallation removes a package if installation metadata update fails
func (pm *PackageManager) RollbackInstallation(packageName string) error {

	cmd := exec.Command("sudo", "dpkg", "--remove", packageName)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rollback failed: %v", err)
	}

	return nil
}

// UpdatePackageMetadata updates the package metadata and handles rollback on failure
func (pm *PackageManager) UpdatePackageMetadata(pkgID string, release *github.Release, pkgMetadata PackageMetadata) error {
	metadata, metaErr := pm.GetPackageManagerMetadata()
	if metaErr != nil {
		return metaErr
	}

	parts := strings.Split(pkgID, "/")
	if len(parts) < 2 {
		return fmt.Errorf("failed to find owner and repo from pkgID: %s", pkgID)
	}

	metadata.Packages[pkgID] = pkgMetadata

	delete(metadata.PendingUpdates, pkgID)

	err := pm.WritePackageManagerMetadata(metadata)
	if err != nil {
		// Attempt rollback if metadata write fails
		var rollbackErr error
		switch pkgMetadata.InstallType {
		case "deb":
			rollbackErr = pm.RollbackInstallation(pkgMetadata.AptName)
		case "binary":
			rollbackErr = pm.RollbackBinaryInstallation(pkgMetadata.BinaryPath)
		}

		if rollbackErr != nil {
			return fmt.Errorf("installation succeeded but metadata write failed, and rollback also failed: %v (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("installation succeeded but metadata write failed (package was rolled back): %v", err)
	}
	return nil
}

// GetBinaryName determines the final binary name for installation
func (pm *PackageManager) GetBinaryName(pkgID, originalName string) string {
	parts := strings.Split(pkgID, "/")
	if len(parts) >= 2 {
		repoName := parts[1]

		// If the original name is just the repo name or similar, use it
		baseName := filepath.Base(originalName)

		// Remove common suffixes that might indicate architecture/platform
		suffixes := []string{"-linux", "-x86_64", "-amd64", "-gnu", ".exe"}
		for _, suffix := range suffixes {
			baseName = strings.TrimSuffix(baseName, suffix)
		}

		// If the cleaned name is similar to repo name, prefer the original
		if strings.Contains(strings.ToLower(baseName), strings.ToLower(repoName)) {
			return baseName
		}

		// Otherwise, use repo name
		return repoName
	}

	// Fallback to cleaned original name
	baseName := filepath.Base(originalName)
	suffixes := []string{"-linux", "-x86_64", "-amd64", "-gnu", ".exe"}
	for _, suffix := range suffixes {
		baseName = strings.TrimSuffix(baseName, suffix)
	}
	return baseName
}

// RollbackBinaryInstallation removes a binary installation
func (pm *PackageManager) RollbackBinaryInstallation(binaryPath string) error {

	cmd := exec.Command("sudo", "rm", "-f", binaryPath)

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary rollback failed: %v", err)
	}

	return nil
}
