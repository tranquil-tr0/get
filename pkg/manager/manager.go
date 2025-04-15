package manager

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tranquil-tr0/get/pkg/github"
)

type PackageManager struct {
	MetadataPath   string
	GithubClient   *github.Client
	Verbose        bool
	PendingUpdates map[string]github.Release // Tracks available updates
}

type PackageMetadata struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	AptName     string `json:"apt_name"`
}

type Metadata struct {
	Packages       map[string]PackageMetadata `json:"packages"`
	PendingUpdates map[string]github.Release  `json:"pending_updates"`
}

func NewPackageManager(metadataPath string) *PackageManager {
	return &PackageManager{
		MetadataPath:   metadataPath,
		GithubClient:   github.NewClient(),
		Verbose:        false,
		PendingUpdates: make(map[string]github.Release),
	}
}

func (pm *PackageManager) LoadMetadata() (*Metadata, error) {
	metadata := &Metadata{
		Packages:       make(map[string]PackageMetadata),
		PendingUpdates: make(map[string]github.Release),
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

func (pm *PackageManager) SaveMetadata(metadata *Metadata) error {
	data, marshalErr := json.MarshalIndent(metadata, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal metadata: %v", marshalErr)
	}

	if writeErr := os.WriteFile(pm.MetadataPath, data, 0644); writeErr != nil {
		return fmt.Errorf("failed to write metadata file: %v", writeErr)
	}

	return nil
}

func (pm *PackageManager) SetVerbose(verbose bool) {
	pm.Verbose = verbose
}

// GetPendingUpdates returns the pending updates from the metadata.
// Returns an error if there are no pending updates available.
func (pm *PackageManager) GetPendingUpdates() (map[string]github.Release, error) {
	// Load the metadata to get pending updates
	metadata, err := pm.LoadMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}

	// Check if there are pending updates
	if len(metadata.PendingUpdates) == 0 {
		return nil, fmt.Errorf("no pending updates available")
	}

	return metadata.PendingUpdates, nil
}
