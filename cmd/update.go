package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/config"
	"github.com/urfave/cli/v3"
	"golang.org/x/mod/semver"
)

type updateUnavailableError struct {
	currentVersion string
	releases       []string
}

func (e *updateUnavailableError) Error() string {
	return fmt.Sprintf("cannot find a non-breaking update (current: %s, releases: %v), but you can pass --force",
		e.currentVersion, e.releases)
}

var forceUpdate bool = false
var releaseURL string = "https://api.github.com/repos/situation-sh/situation/releases"
var releaseToken string = ""

var updateCmd = cli.Command{
	Name:   "update",
	Usage:  "Update the agent",
	Action: updateAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "force",
			Usage:       "Force update (breaking changes are possible)",
			Value:       false,
			Destination: &forceUpdate,
			Aliases:     []string{"f"},
		},
		&cli.StringFlag{
			Name:        "release-url",
			Usage:       "Define the endpoint that lists situation releases",
			Value:       releaseURL,
			Destination: &releaseURL,
			Aliases:     []string{"-r"},
		},
		&cli.StringFlag{
			Name:        "release-token",
			Usage:       "Optional token to fetch releases",
			Value:       releaseToken,
			Destination: &releaseToken,
			Aliases:     []string{"-t"},
		},
	},
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		logrus.SetLevel(logrus.InfoLevel)
		return ctx, nil
	},
}

func listReleases() ([]GitHubRelease, error) {
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if releaseToken != "" {
		req.Header.Set("Authorization", "Bearer "+releaseToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	releases := make([]GitHubRelease, 0)
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func filterReleases(releases []GitHubRelease, currentVersion string) []GitHubRelease {
	out := make([]GitHubRelease, 0)

	// compare with tags
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		// return +1 if release.TagName > config.Version
		if semver.Compare(release.TagName, currentVersion) > 0 {
			out = append(out, release)
		}
	}
	return out
}

func selectRelease() (*GitHubRelease, error) {
	releases, err := listReleases()
	if err != nil {
		return nil, err
	}
	// clean the version
	currentVersion := config.Version
	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}
	upperReleases := filterReleases(releases, currentVersion)
	if len(upperReleases) == 0 {
		return nil, fmt.Errorf("no release found")
	}

	for _, release := range upperReleases {
		c := semver.Compare(semver.MajorMinor(release.TagName), semver.MajorMinor(currentVersion))
		if c > 0 && forceUpdate {
			return &release, nil
		}
		if c == 0 {
			return &release, nil
		}
	}

	available := make([]string, 0)
	for _, r := range upperReleases {
		available = append(available, r.TagName)
	}
	return nil, &updateUnavailableError{
		currentVersion: currentVersion,
		releases:       available,
	}
}

func downloadBinary(u string) ([]byte, error) {
	// Send HTTP GET request
	parsedURL, err := url.Parse(u)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid URL: %s", u)
	}
	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP response code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status returned: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func updateAction(ctx context.Context, cmd *cli.Command) error {
	release, err := selectRelease()
	if err != nil {
		return err
	}
	logrus.Infof("A new version is available: %s (current: v%s)\n", release.TagName, config.Version)

	url, err := getBinaryURL(release)
	if err != nil {
		return err
	}

	binary, err := downloadBinary(url)
	if err != nil {
		return err
	}
	logrus.Infof("Download successful")
	// inject our current ID into the downloaded binary
	toWrite := bytes.Replace(binary, config.GetDefaultID(), config.ID[:16], 1)
	if err := selfupdate.Apply(bytes.NewReader(toWrite), selfupdate.Options{}); err != nil {
		return err
	}
	logrus.Infof("Situation has been updated to %s\n", release.TagName)
	return nil
}

type GitHubRelease struct {
	URL             string        `json:"url"`
	ID              int64         `json:"id"`
	NodeID          string        `json:"node_id"`
	TagName         string        `json:"tag_name"`
	TargetCommitish string        `json:"target_commitish"`
	Name            *string       `json:"name"`
	Draft           bool          `json:"draft"`
	Prerelease      bool          `json:"prerelease"`
	CreatedAt       time.Time     `json:"created_at"`
	PublishedAt     *time.Time    `json:"published_at"`
	Assets          []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	URL                string    `json:"url"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	ID                 int64     `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              *string   `json:"label"`
	State              string    `json:"state"`
	ContentType        string    `json:"content_type"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
