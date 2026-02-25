//go:build linux

package cmd

import (
	"fmt"
	"strings"
)

func getBinaryURL(release *GitHubRelease) (string, error) {
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "linux") {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no linux binary found in release %s", release.TagName)
}
