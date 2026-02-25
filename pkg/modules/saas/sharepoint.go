package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&SharePointDetector{
		// SharePoint Online IP ranges (ID 31)
		// see https://learn.microsoft.com/en-us/microsoft-365/enterprise/urls-and-ip-address-ranges
		nets: []*net.IPNet{
			// IPv4
			mustParseCIDR("13.107.136.0/22"),
			mustParseCIDR("40.108.128.0/17"),
			mustParseCIDR("52.104.0.0/14"),
			mustParseCIDR("104.146.128.0/17"),
			mustParseCIDR("150.171.40.0/22"),
			// IPv6
			mustParseCIDR("2603:1061:1300::/40"),
			mustParseCIDR("2603:1063:6000::/35"),
			mustParseCIDR("2620:1ec:8f8::/46"),
			mustParseCIDR("2620:1ec:908::/46"),
			mustParseCIDR("2a01:111:f402::/48"),
		},
		// DNS suffixes from SharePoint endpoint rows (IDs 31, 37)
		dnsSuffixes: []string{
			".sharepoint.com",
			".sharepointonline.com",
		},
	})
}

type SharePointDetector struct {
	nets        []*net.IPNet
	dnsSuffixes []string
}

func (d *SharePointDetector) Name() string {
	return "SharePoint"
}

func (d *SharePointDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, d.dnsSuffixes) {
		return true, nil
	}
	return matchIPRange(endpoint, d.nets)
}
