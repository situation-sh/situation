//go:build windows
// +build windows

package cmd

import (
	"fmt"
	"strings"
)

func getBinaryURL(release *GitHubRelease) (string, error) {
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "windows") {
			return asset.URL, nil
		}
	}
	return "", fmt.Errorf("no linux binary found in release %s", release.TagName)
}
