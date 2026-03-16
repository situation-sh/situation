package saas

import (
	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&DiscordDetector{})
}

type DiscordDetector struct{}

func (d *DiscordDetector) Name() string {
	return "Discord"
}

func (d *DiscordDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, []string{
		"discord.gg",
		"discord.com",
	}) {
		return true, nil
	}
	return false, nil
}
