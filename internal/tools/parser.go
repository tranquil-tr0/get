package tools

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseRepoURL(url string) (pkgID string, err error) {
	// Remove protocol and domain if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	// TODO: consider removing dir beyond username/repo, but need to support specific releases if planning on impelementing that
	pkgID = url

	if len(pkgID) <= 2 {
		return "", fmt.Errorf("Invalid GitHub repository URL")
	}

	return pkgID, nil

}

func AreAssetNamesSimilar(name1, name2 string) (bool, error) {
	re, err := regexp.Compile(`v[0-9]+\.[0-9]+\.[0-9]+`)
	if err != nil {
		return false, err
	}
	name1 = re.ReplaceAllString(name1, "")
	name2 = re.ReplaceAllString(name2, "")
	return name1 == name2, nil
}
