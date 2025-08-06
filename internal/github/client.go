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

type ReleaseOptions struct {
	TagPrefix string // Filter releases by tag prefix (e.g., "auth-", "photos-")
}

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
	// NOTE: ContentType is not a reliable detection method between binaries and even sometimes zips give "application/octet-stream"
	ContentType string `json:"content_type"`
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

func (c *Client) GetLatestReleaseWithOptions(pkgID string, options *ReleaseOptions) (*Release, error) {
	return c.GetReleaseByTagWithOptions(pkgID, "latest", options)
}

func (c *Client) GetReleaseByTag(pkgID, tag string) (*Release, error) {
	return c.GetReleaseByTagWithOptions(pkgID, tag, nil)
}

func (c *Client) GetReleaseByTagWithOptions(pkgID, tag string, options *ReleaseOptions) (*Release, error) {
	// Parse owner and repo from pkgID (format: owner/repo)
	parts := strings.Split(pkgID, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package ID format: %s (expected owner/repo)", pkgID)
	}
	owner, repo := parts[0], parts[1]

	var githubRelease *github.RepositoryRelease
	var err error

	if tag == "latest" {
		if options != nil && options.TagPrefix != "" {
			// Get latest release with tag prefix filtering
			githubRelease, err = c.getLatestReleaseWithPrefix(owner, repo, options.TagPrefix)
		} else {
			// Get the latest release
			githubRelease, _, err = c.client.Repositories.GetLatestRelease(c.ctx, owner, repo)
		}
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

// returns the first package in the release with a .deb extension
func (r *Release) FindFirstDebPackage() *Asset {
	for _, asset := range r.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			return &asset
		}
	}
	return nil
}

// FindDebPackages returns .deb packages in the release without requiring download of the assets
func (r *Release) FindDebPackages() []Asset {
	var debPackages []Asset
	for _, asset := range r.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			debPackages = append(debPackages, asset)
		}
	}
	return debPackages
}

// FindBinaryAssets returns assets that are likely Linux executables without requiring download of the assets
func (r *Release) FindBinaryAssets() []Asset {
	// TODO: consider better detection logic
	var binaries []Asset
	for _, asset := range r.Assets {
		name := asset.Name
		if asset.ContentType == "application/octet-stream" &&
			(!strings.Contains(name, ".") ||
				strings.HasSuffix(name, ".run") ||
				strings.HasSuffix(name, ".bin")) {
			binaries = append(binaries, asset)
		}
		// NOTE: ContentType is used to filter out text files etc
	}
	return binaries
}

// GetAllInstallableAssets returns both .deb packages and potential binary assets
func (r *Release) GetAllInstallableAssets() ([]Asset, []Asset) {
	return r.FindDebPackages(), r.FindBinaryAssets()
}

// getLatestReleaseWithPrefix finds the latest release with a specific tag prefix
func (c *Client) getLatestReleaseWithPrefix(owner, repo, tagPrefix string) (*github.RepositoryRelease, error) {
	// List all releases to find the latest one with the specified prefix
	releases, _, err := c.client.Repositories.ListReleases(c.ctx, owner, repo, &github.ListOptions{
		PerPage: 100, // Get more releases to ensure we find the latest with prefix
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %v", err)
	}

	// Find the latest release with the specified tag prefix
	for _, release := range releases {
		tagName := release.GetTagName()
		if strings.HasPrefix(tagName, tagPrefix) {
			return release, nil
		}
	}

	return nil, fmt.Errorf("no releases found with tag prefix \"%s\"", tagPrefix)
}

func (c *Client) GetLatestVersionName(pkgID string) (string, error) {
	return c.GetLatestVersionNameWithOptions(pkgID, nil)
}

func (c *Client) GetLatestVersionNameWithOptions(pkgID string, options *ReleaseOptions) (string, error) {
	release, err := c.GetLatestReleaseWithOptions(pkgID, options)
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
