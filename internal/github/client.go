package github

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func NewClient() *Client {
	ctx := context.Background()

	// Create a new GitHub client
	client := github.NewClient(nil)

	// Check for GitHub token in environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = client.WithAuthToken(token)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func (c *Client) GetLatestRelease(pkgID string) (*Release, error) {
	return c.GetReleaseByTag(pkgID, "latest")
}

func (c *Client) GetReleaseByTag(pkgID, tag string) (*Release, error) {
	// Parse owner and repo from pkgID (format: owner/repo)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package ID format: %s (expected owner/repo)", pkgID)
	}
	owner, repo := parts[0], parts[1]

	var githubRelease *github.RepositoryRelease
	var err error

	if tag == "latest" {
		// Get the latest release
		githubRelease, _, err = c.client.Repositories.GetLatestRelease(c.ctx, owner, repo)
	} else {
		// Get release by tag
		githubRelease, _, err = c.client.Repositories.GetReleaseByTag(c.ctx, owner, repo, tag)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %v", err)
	}

	// Convert go-github release to our Release struct
	release := &Release{
		TagName:     githubRelease.GetTagName(),
		Name:        githubRelease.GetName(),
		PublishedAt: githubRelease.GetPublishedAt().Format(time.RFC3339),
		Assets:      make([]Asset, len(githubRelease.Assets)),
	}

	// Convert assets
	for i, asset := range githubRelease.Assets {
		release.Assets[i] = Asset{
			Name:               asset.GetName(),
			BrowserDownloadURL: asset.GetBrowserDownloadURL(),
		}
	}

	return release, nil
}

func (r *Release) FindDebPackage() *Asset {
	for _, asset := range r.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			return &asset
		}
	}
	return nil
}

// FindDebPackages returns all .deb packages in the release
func (r *Release) FindDebPackages() []Asset {
	var debPackages []Asset
	for _, asset := range r.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			debPackages = append(debPackages, asset)
		}
	}
	return debPackages
}

// FindBinaryAssets returns assets that are likely Linux executables based on common patterns
func (r *Release) FindBinaryAssets() []Asset {
	var binaries []Asset
	
	// Extensions that are NOT binaries
	nonBinaryExts := []string{
		".deb", ".rpm", ".tar.gz", ".tgz", ".zip", ".tar.bz2", ".tar.xz",
		".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".html",
		".sig", ".asc", ".sha256", ".sha512", ".checksum", ".exe", ".dll",
	}
	
	for _, asset := range r.Assets {
		name := strings.ToLower(asset.Name)
		
		// Skip if it has a non-binary extension
		isBinary := true
		for _, ext := range nonBinaryExts {
			if strings.HasSuffix(name, ext) {
				isBinary = false
				break
			}
		}
		
		if !isBinary {
			continue
		}
	}
	
	return binaries
}

// GetAllInstallableAssets returns both .deb packages and potential binary assets
func (r *Release) GetAllInstallableAssets() ([]Asset, []Asset) {
	return r.FindDebPackages(), r.FindBinaryAssets()
}

func (c *Client) GetLatestVersionName(pkgID string) (string, error) {
	release, err := c.GetLatestRelease(pkgID)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %v", err)
	}

	// Normalize version by removing all non-numeric characters from the beginning and end
	// This matches the normalization done in parseVersionToInt()
	version := strings.TrimFunc(release.TagName, func(r rune) bool {
		return r != '.' && (r < '0' || r > '9')
	})

	// More flexible validation for semantic versioning including pre-release and build metadata
	// Accepts formats like: 1.2.3, 1.2.3-alpha, 1.2.3-alpha.1, 1.2.3+build
	matched, err := regexp.MatchString(`^\d+(\.\d+)*(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?(\+[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$`, version)
	if !matched || err != nil {
		return "", fmt.Errorf("invalid version format in tag: %s read as %s", release.TagName, version)
	}

	return version, nil
}
