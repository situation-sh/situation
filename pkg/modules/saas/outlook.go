package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&MicrosoftOutlookDetector{
		// Exchange Online IP ranges (ID 1)
		// see https://learn.microsoft.com/en-us/microsoft-365/enterprise/urls-and-ip-address-ranges
		nets: []*net.IPNet{
			// IPv4
			mustParseCIDR("13.107.6.152/31"),
			mustParseCIDR("13.107.18.10/31"),
			mustParseCIDR("13.107.128.0/22"),
			mustParseCIDR("23.103.160.0/20"),
			mustParseCIDR("40.96.0.0/13"),
			mustParseCIDR("40.104.0.0/15"),
			mustParseCIDR("52.96.0.0/14"),
			mustParseCIDR("131.253.33.215/32"),
			mustParseCIDR("132.245.0.0/16"),
			mustParseCIDR("150.171.32.0/22"),
			mustParseCIDR("204.79.197.215/32"),
			// IPv6
			mustParseCIDR("2603:1006::/40"),
			mustParseCIDR("2603:1016::/36"),
			mustParseCIDR("2603:1026::/36"),
			mustParseCIDR("2603:1036::/36"),
			mustParseCIDR("2603:1046::/36"),
			mustParseCIDR("2603:1056::/36"),
			mustParseCIDR("2620:1ec:4::152/128"),
			mustParseCIDR("2620:1ec:4::153/128"),
			mustParseCIDR("2620:1ec:c::10/128"),
			mustParseCIDR("2620:1ec:c::11/128"),
			mustParseCIDR("2620:1ec:d::10/128"),
			mustParseCIDR("2620:1ec:d::11/128"),
			mustParseCIDR("2620:1ec:8f0::/46"),
			mustParseCIDR("2620:1ec:900::/46"),
			mustParseCIDR("2620:1ec:a92::152/128"),
			mustParseCIDR("2620:1ec:a92::153/128"),
		},
		// DNS suffixes from Exchange Online endpoint rows (IDs 1, 8, 9, 10)
		dnsSuffixes: []string{
			".outlook.com",
			"outlook.cloud.microsoft",
			"outlook.office.com",
			"outlook.office365.com",
			".mx.microsoft",
		},
	})
}

type MicrosoftOutlookDetector struct {
	nets        []*net.IPNet
	dnsSuffixes []string
}

func (d *MicrosoftOutlookDetector) Name() string {
	return "Microsoft Outlook"
}

func (d *MicrosoftOutlookDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, d.dnsSuffixes) {
		return true, nil
	}
	return matchIPRange(endpoint, d.nets)
}
