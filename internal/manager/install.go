package manager

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tranquil-tr0/get/internal/github"
	"github.com/tranquil-tr0/get/internal/output"
)

// SelectAssetInteractively prompts the user to select which asset to install
func (pm *PackageManager) SelectAssetInteractively(release *github.Release) (*github.Asset, string, error) {
	debPackages := release.FindDebPackages()
	binaryAssets := release.FindBinaryAssets()
	
	fmt.Printf("\nAvailable assets in release %s:\n", release.TagName)
	
	var allAssets []github.Asset
	var assetTypes []string
	
	// Add .deb packages
	for _, asset := range debPackages {
		allAssets = append(allAssets, asset)
		assetTypes = append(assetTypes, "deb")
		fmt.Printf("  [%d] %s (.deb package)\n", len(allAssets), asset.Name)
	}
	
	// Add binary assets
	for _, asset := range binaryAssets {
		allAssets = append(allAssets, asset)
		assetTypes = append(assetTypes, "binary")
		fmt.Printf("  [%d] %s (binary executable)\n", len(allAssets), asset.Name)
	}
	
	// Add option to specify other file as executable
	fmt.Printf("  [%d] Other file (specify as executable)\n", len(allAssets)+1)
	
	if len(allAssets) == 0 {
		fmt.Println("\nNo .deb packages or likely binary executables found.")
		fmt.Println("Available assets:")
		for _, asset := range release.Assets {
			fmt.Printf("  - %s\n", asset.Name)
		}
	}
	
	fmt.Print("\nSelect an option (number): ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	
	choice, err := strconv.Atoi(input)
	if err != nil {
		return nil, "", fmt.Errorf("invalid selection: %s", input)
	}
	
	// Handle "Other file" option
	if choice == len(allAssets)+1 {
		fmt.Println("\nAvailable assets:")
		for i, asset := range release.Assets {
			fmt.Printf("  [%d] %s\n", i+1, asset.Name)
		}
		fmt.Print("Select asset number: ")
		
		scanner.Scan()
		otherInput := strings.TrimSpace(scanner.Text())
		otherChoice, err := strconv.Atoi(otherInput)
		if err != nil || otherChoice < 1 || otherChoice > len(release.Assets) {
			return nil, "", fmt.Errorf("invalid asset selection: %s", otherInput)
		}
		
		selectedAsset := release.Assets[otherChoice-1]
		return &selectedAsset, "binary", nil
	}
	
	// Validate choice
	if choice < 1 || choice > len(allAssets) {
		return nil, "", fmt.Errorf("invalid selection: %d (must be between 1 and %d)", choice, len(allAssets))
	}
	
	selectedAsset := allAssets[choice-1]
	selectedType := assetTypes[choice-1]
	
	return &selectedAsset, selectedType, nil
}

func (pm *PackageManager) InstallRelease(pkgID string, release *github.Release, preSelectedAsset *github.Asset) error {
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
		output.PrintVerboseComplete("Using pre-selected asset", fmt.Sprintf("%s (%s)", selectedAsset.Name, installType))
	} else {
		// Interactive asset selection
		output.PrintVerboseStart("Selecting asset for installation", release.TagName)
		selectedAsset, installType, err = pm.SelectAssetInteractively(release)
		if err != nil {
			output.PrintVerboseError("Select asset", err)
			return fmt.Errorf("failed to select asset: %v", err)
		}
		output.PrintVerboseComplete("Select asset", fmt.Sprintf("%s (%s)", selectedAsset.Name, installType))
	}

	if preSelectedAsset == nil {
		// Save chosen asset
		metadata, metaErr := pm.GetPackageManagerMetadata()
		if metaErr != nil {
			output.PrintVerboseError("Load package metadata", metaErr)
			return metaErr
		}
		pkgMetadata := metadata.Packages[pkgID]
		pkgMetadata.ChosenAsset = selectedAsset.Name
		metadata.Packages[pkgID] = pkgMetadata
		if err := pm.WritePackageManagerMetadata(metadata); err != nil {
			output.PrintVerboseError("Write package metadata", err)
			return err
		}
		output.PrintVerboseComplete("Saved chosen asset as ", selectedAsset.Name)
	}

	// Route to appropriate installation method
	switch installType {
	case "deb":
		return pm.InstallDebPackage(pkgID, release, selectedAsset)
	case "binary":
		return pm.InstallBinary(pkgID, release, selectedAsset)
	default:
		return fmt.Errorf("unsupported installation type: %s", installType)
	}
}

// InstallDebPackage handles .deb package installation
func (pm *PackageManager) InstallDebPackage(pkgID string, release *github.Release, debAsset *github.Asset) error {
	// Download package
	output.PrintVerboseStart("Downloading .deb package", debAsset.BrowserDownloadURL)
	resp, httpErr := http.Get(debAsset.BrowserDownloadURL)
	if httpErr != nil {
		output.PrintVerboseError("Download package", httpErr)
		return fmt.Errorf("failed to download package: %v", httpErr)
	}
	defer resp.Body.Close()
	output.PrintVerboseDebug("HTTP", "Download response status: %s", resp.Status)

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
	packagePath := filepath.Join(tempDir, debAsset.Name)
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

	// Install with dpkg
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

	// Extract package name using dpkg-deb
	output.PrintVerboseStart("Extracting package name")
	aptPackageName, nameErr := pm.GetPackageNameFromDeb(packagePath)
	if nameErr != nil {
		// Fallback: extract from .deb filename
		output.PrintVerboseDebug("DPKG", "Falling back to extracting from .deb filename")
		debFilename := filepath.Base(packagePath)
		if strings.HasSuffix(debFilename, ".deb") {
			nameWithoutExt := strings.TrimSuffix(debFilename, ".deb")
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

	// Update metadata
	return pm.UpdatePackageMetadata(pkgID, release, PackageMetadata{
		Version:      strings.TrimPrefix(release.TagName, "v"),
		InstalledAt:  release.PublishedAt,
		AptName:      aptPackageName,
		InstallType:  "deb",
		OriginalName: debAsset.Name,
	})
}

// InstallBinary handles binary executable installation
func (pm *PackageManager) InstallBinary(pkgID string, release *github.Release, binaryAsset *github.Asset) error {
	// Download binary
	output.PrintVerboseStart("Downloading binary", binaryAsset.BrowserDownloadURL)
	resp, httpErr := http.Get(binaryAsset.BrowserDownloadURL)
	if httpErr != nil {
		output.PrintVerboseError("Download binary", httpErr)
		return fmt.Errorf("failed to download binary: %v", httpErr)
	}
	defer resp.Body.Close()
	output.PrintVerboseDebug("HTTP", "Download response status: %s", resp.Status)

	// Create temp directory
	output.PrintVerboseStart("Creating temporary directory")
	tempDir, tempErr := os.MkdirTemp("", "get-*")
	if tempErr != nil {
		output.PrintVerboseError("Create temporary directory", tempErr)
		return fmt.Errorf("failed to create temp directory: %v", tempErr)
	}
	defer os.RemoveAll(tempDir)
	output.PrintVerboseComplete("Create temporary directory", tempDir)

	// Save binary
	tempBinaryPath := filepath.Join(tempDir, binaryAsset.Name)
	output.PrintVerboseStart("Saving binary file", tempBinaryPath)
	file, createErr := os.Create(tempBinaryPath)
	if createErr != nil {
		output.PrintVerboseError("Create binary file", createErr)
		return fmt.Errorf("failed to create binary file: %v", createErr)
	}

	if _, copyErr := io.Copy(file, resp.Body); copyErr != nil {
		file.Close()
		output.PrintVerboseError("Save binary file", copyErr)
		return fmt.Errorf("failed to save binary file: %v", copyErr)
	}
	file.Close()
	output.PrintVerboseComplete("Save binary file", tempBinaryPath)

	// Make binary executable
	output.PrintVerboseStart("Making binary executable")
	if err := os.Chmod(tempBinaryPath, 0755); err != nil {
		output.PrintVerboseError("Make binary executable", err)
		return fmt.Errorf("failed to make binary executable: %v", err)
	}
	output.PrintVerboseComplete("Make binary executable")

	// Determine final binary name and path
	binaryName := pm.GetBinaryName(pkgID, binaryAsset.Name)
	finalBinaryPath := filepath.Join("/usr/local/bin", binaryName)

	// Install binary to /usr/local/bin
	output.PrintVerboseStart("Installing binary to /usr/local/bin", finalBinaryPath)
	cmd := exec.Command("sudo", "-p", "[get] Password required for binary installation: ", "cp", tempBinaryPath, finalBinaryPath)
	output.PrintVerboseDebug("INSTALL", "Command: %v", cmd.Args)

	cmdOutput, installErr := cmd.CombinedOutput()
	if installErr != nil {
		output.PrintVerboseError("Install binary", installErr)
		output.PrintVerboseDebug("INSTALL", "Installation output: %s", string(cmdOutput))
		return fmt.Errorf("failed to install binary: %v\nOutput: %s", installErr, cmdOutput)
	}
	output.PrintVerboseComplete("Install binary to /usr/local/bin", finalBinaryPath)

	fmt.Printf("Binary installed as: %s\n", binaryName)

	// Update metadata
	return pm.UpdatePackageMetadata(pkgID, release, PackageMetadata{
		Version:      strings.TrimPrefix(release.TagName, "v"),
		InstalledAt:  release.PublishedAt,
		BinaryPath:   finalBinaryPath,
		InstallType:  "binary",
		OriginalName: binaryAsset.Name,
	})
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
	return pm.InstallRelease(pkgID, release, nil)
}

// InstallVersion does InstallRelease after fetching the release based on version
func (pm *PackageManager) InstallVersion(pkgID string, version string, chosenAsset *github.Asset) error {
	// get the Release
	output.PrintVerboseStart("Fetching specific release for installation", fmt.Sprintf("%s@%s", pkgID, version))
	release, err := pm.GithubClient.GetReleaseByTag(pkgID, version)
	if err != nil {
		output.PrintVerboseError("Fetch GitHub release", err)
		return fmt.Errorf("error fetching latest release: %s", err)
	}
	output.PrintVerboseComplete("Fetch GitHub release", release.TagName)

	// install the package with a call to InstallRelease, and returns error
	return pm.InstallRelease(pkgID, release, chosenAsset)
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

// UpdatePackageMetadata updates the package metadata and handles rollback on failure
func (pm *PackageManager) UpdatePackageMetadata(pkgID string, release *github.Release, pkgMetadata PackageMetadata) error {
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

	output.PrintVerboseDebug("METADATA", "Adding package: %s version %s", pkgID, pkgMetadata.Version)
	metadata.Packages[pkgID] = pkgMetadata

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
		var rollbackErr error
		switch pkgMetadata.InstallType {
case "deb":
			rollbackErr = pm.RollbackInstallation(pkgMetadata.AptName)
		case "binary":
			rollbackErr = pm.RollbackBinaryInstallation(pkgMetadata.BinaryPath)
		}
		
		if rollbackErr != nil {
			output.PrintVerboseError("Rollback installation", rollbackErr)
			return fmt.Errorf("installation succeeded but metadata write failed, and rollback also failed: %v (rollback error: %v)", err, rollbackErr)
		}
		return fmt.Errorf("installation succeeded but metadata write failed (package was rolled back): %v", err)
	}
	output.PrintVerboseComplete("Update package metadata")
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
	output.PrintVerboseStart("Rolling back binary installation", binaryPath)
	
	cmd := exec.Command("sudo", "rm", "-f", binaryPath)
	output.PrintVerboseDebug("ROLLBACK", "Binary rollback command: %v", cmd.Args)
	
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		output.PrintVerboseError("Rollback binary installation", err)
		output.PrintVerboseDebug("ROLLBACK", "Rollback output: %s", string(cmdOutput))
		return fmt.Errorf("binary rollback failed: %v", err)
	}
	
	output.PrintVerboseComplete("Rollback binary installation", binaryPath)
	return nil
}
