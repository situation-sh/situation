package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&MicrosoftTeamsDetector{
		// Microsoft Teams IP ranges (IDs 11, 12)
		// see https://learn.microsoft.com/en-us/microsoft-365/enterprise/urls-and-ip-address-ranges
		nets: []*net.IPNet{
			// IPv4
			mustParseCIDR("52.112.0.0/14"),
			mustParseCIDR("52.122.0.0/15"),
			// IPv6
			mustParseCIDR("2603:1027::/48"),
			mustParseCIDR("2603:1037::/48"),
			mustParseCIDR("2603:1047::/48"),
			mustParseCIDR("2603:1057::/48"),
			mustParseCIDR("2603:1063::/38"),
			mustParseCIDR("2620:1ec:6::/48"),
			mustParseCIDR("2620:1ec:40::/42"),
		},
		// DNS suffixes from Teams endpoint rows (IDs 12, 127)
		dnsSuffixes: []string{
			".lync.com",
			".teams.cloud.microsoft",
			".teams.microsoft.com",
			"teams.cloud.microsoft",
			"teams.microsoft.com",
			".skype.com",
		},
	})
}

type MicrosoftTeamsDetector struct {
	nets        []*net.IPNet
	dnsSuffixes []string
}

func (d *MicrosoftTeamsDetector) Name() string {
	return "Microsoft Teams"
}

func (d *MicrosoftTeamsDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, d.dnsSuffixes) {
		return true, nil
	}
	return matchIPRange(endpoint, d.nets)
}
