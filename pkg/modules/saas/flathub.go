package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&FlathubDetector{
		// dig dl.flathub.org +short
		nets: []*net.IPNet{
			// IPv4
			mustParseCIDR("151.101.129.91/32"),
			mustParseCIDR("151.101.65.91/32"),
			mustParseCIDR("151.101.1.91/32"),
			mustParseCIDR("151.101.193.91/32"),
		},
	})
}

type FlathubDetector struct {
	nets []*net.IPNet
}

func (d *FlathubDetector) Name() string {
	return "Flathub"
}

func (d *FlathubDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	return matchIPRange(endpoint, d.nets)
}
