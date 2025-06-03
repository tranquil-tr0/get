package tools

import (
	"fmt"
	"strings"
)

func ParseRepoURL(url string) (pkgID string, err error) {
	// Remove protocol and domain if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	pkgID = url

	if len(pkgID) <= 2 {
		return "", fmt.Errorf("Invalid GitHub repository URL")
	}

	return pkgID, nil

}
