package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&AnthoropicDetector{
		// Anthropic IP ranges - see https://ipinfo.io/AS399358
		nets: []*net.IPNet{
			mustParseCIDR("160.79.104.0/23"),
			mustParseCIDR("209.249.57.0/24"),
			mustParseCIDR("2607:6bc0::/48"),
			mustParseCIDR("2607:6bc0:11::/48"),
		},
	})
}

type AnthoropicDetector struct {
	nets []*net.IPNet
}

func (d *AnthoropicDetector) Name() string {
	return "Anthropic"
}

func (d *AnthoropicDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	return matchIPRange(endpoint, d.nets)
}
