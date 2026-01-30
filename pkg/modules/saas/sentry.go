package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	// see https://docs.sentry.io/security-legal-pii/security/ip-ranges/
	// or curl -L 'https://docs.sentry.io/api/ip-ranges' | jq
	registerDetector(&SentryDetector{
		nets: []*net.IPNet{
			// dashboard
			mustParseCIDR("35.186.247.156/32"), // sentry.io / us.sentry.io
			mustParseCIDR("34.36.122.224/32"),  // de.sentry.io
			mustParseCIDR("34.36.87.148/32"),   // de.sentry.io
			// event ingestion (apex)
			// 35.186.247.156/32 already listed above
			// event ingestion (organization subdomains)
			mustParseCIDR("34.120.195.249/32"),
			mustParseCIDR("34.160.81.0/32"),
			mustParseCIDR("34.102.210.18/32"),
			mustParseCIDR("2600:1901:0:5e8a::/64"),
			mustParseCIDR("2600:1901:0:7edb::/64"),
			mustParseCIDR("34.120.62.213/32"),
			// 34.160.81.0/32, 34.102.210.18/32 already listed above
			// event ingestion (legacy)
			mustParseCIDR("34.96.102.34/32"),
			// outbound requests (US)
			mustParseCIDR("35.184.238.160/32"),
			mustParseCIDR("104.155.159.182/32"),
			mustParseCIDR("104.155.149.19/32"),
			mustParseCIDR("130.211.230.102/32"),
			// outbound requests (EU)
			mustParseCIDR("34.141.31.19/32"),
			mustParseCIDR("34.141.4.162/32"),
			mustParseCIDR("35.234.78.236/32"),
			// email delivery
			mustParseCIDR("167.89.86.73/32"),
			mustParseCIDR("167.89.84.75/32"),
			mustParseCIDR("167.89.84.14/32"),
			// uptime monitoring
			mustParseCIDR("34.123.33.225/32"),
			mustParseCIDR("34.41.121.171/32"),
			mustParseCIDR("34.169.179.115/32"),
			mustParseCIDR("35.237.134.233/32"),
			mustParseCIDR("34.85.249.57/32"),
			mustParseCIDR("34.159.197.47/32"),
			mustParseCIDR("35.242.231.10/32"),
			mustParseCIDR("34.107.93.3/32"),
			mustParseCIDR("35.204.169.245/32"),
		},
		dnsSuffixes: []string{
			".sentry.io",
		},
	})
}

type SentryDetector struct {
	nets        []*net.IPNet
	dnsSuffixes []string
}

func (d *SentryDetector) Name() string {
	return "Sentry"
}

func (d *SentryDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	if matchDNSSuffix(endpoint, d.dnsSuffixes) {
		return true, nil
	}
	return matchIPRange(endpoint, d.nets)
}
