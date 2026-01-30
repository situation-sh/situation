package saas

import (
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&GithubDetector{})
}

type GithubDetector struct{}

func (d *GithubDetector) Name() string {
	return "GitHub"
}

func (d *GithubDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, []string{
		".github.com",
		".github.io",
		".githubusercontent.com",
	}) {
		return true, nil
	}
	return false, nil
}
