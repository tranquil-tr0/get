package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Client struct {
	HttpClient *http.Client
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
	return &Client{
		HttpClient: &http.Client{},
	}
}

func (c *Client) GetLatestRelease(pkgID string) (*Release, error) {
	return c.GetReleaseByTag(pkgID, "latest")
}

func (c *Client) GetReleaseByTag(pkgID, tag string) (*Release, error) {
	var url string
	if tag == "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", pkgID)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", pkgID, tag)
	}

	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %v", err)
	}

	return &release, nil
}

func (r *Release) FindDebPackage() *Asset {
	for _, asset := range r.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			return &asset
		}
	}
	return nil
}

func (c *Client) GetLatestVersionName(pkgID string) (string, error) {
	release, err := c.GetLatestRelease(pkgID)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %v", err)
	}

	// Extract version number from tag name (remove 'v' prefix if present)
	version := strings.TrimPrefix(release.TagName, "v")

	// Validate version format (semver-like)
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	if !matched || err != nil {
		return "", fmt.Errorf("invalid version format in tag: %s read as %s", release.TagName, version)
	}

	return version, nil
}
