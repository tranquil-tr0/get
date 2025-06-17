package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tranquil-tr0/get/internal/github"
)

type PackageManager struct {
	MetadataPath string
	GithubClient *github.Client
	Verbose      bool
}

type PackageMetadata struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	AptName     string `json:"apt_name"`
}

type PackageManagerMetadata struct {
	Packages       map[string]PackageMetadata `json:"packages"`        // map[pkgID] = PackageMetadata of the package
	PendingUpdates map[string]string          `json:"pending_updates"` // map[pkgID] = latest_version
}

// NewPackageManager returns a new PackageManager struct that stores metadata at the specified path
func NewPackageManager(metadataPath string) *PackageManager {
	return &PackageManager{
		MetadataPath: metadataPath,
		GithubClient: github.NewClient(),
	}
}

func (pm *PackageManager) GetPackageManagerMetadata() (*PackageManagerMetadata, error) {
	metadata := &PackageManagerMetadata{
		Packages:       make(map[string]PackageMetadata),
		PendingUpdates: make(map[string]string),
	}

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

// WritePackageManagerMetadata overwrites PackageManagerMetadata
func (pm *PackageManager) WritePackageManagerMetadata(metadata *PackageManagerMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	// Ensure the directory exists before writing the file
	if err := os.MkdirAll(filepath.Dir(pm.MetadataPath), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %v", err)
	}

	if err := os.WriteFile(pm.MetadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	return nil
}

// GetPendingUpdates returns the pending updates from the metadata.
// Returns an error if there are no pending updates available.
func (pm *PackageManager) GetAllPendingUpdates() (map[string]string, error) {
	// Load the metadata to get pending updates
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}

	// Check if there are pending updates
	if len(metadata.PendingUpdates) == 0 {
		return nil, fmt.Errorf("no pending updates available")
	}

	return metadata.PendingUpdates, nil
}

// GetPendingUpdate returns the version of the pending update for the package.
// If version="", the pkgID does NOT have a version or does not exist.
func (pm *PackageManager) GetPendingUpdate(pkgID string) (version string, err error) {
	// Check if pkg has a pending update
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return "", fmt.Errorf("error loading metadata: %v", err)
	}
	version, exists := metadata.PendingUpdates[pkgID]
	if !exists {
		// Package does NOT have a pending update, but no error
		return "", nil
	}
	// Return version
	return version, nil
}

// GetPackage retrieves a package by its ID from the metadata.
func (pm *PackageManager) GetPackage(pkgID string) (*PackageMetadata, error) {
	metadata, err := pm.GetPackageManagerMetadata()
	if err != nil {
		return nil, err
	}

	pkg, exists := metadata.Packages[pkgID]
	if !exists {
		return nil, fmt.Errorf("package %s not found", pkgID)
	}
	return &pkg, nil
}
