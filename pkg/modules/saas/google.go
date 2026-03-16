package saas

import (
	"net"

	"github.com/situation-sh/situation/pkg/models"
)

func init() {
	registerDetector(&GoogleDetector{
		// Google IP ranges - See https://www.gstatic.com/ipranges/goog.json
		// curl 'https://www.gstatic.com/ipranges/goog.json'|jq '.prefixes[]|.ipv4Prefix // .ipv6Prefix'|awk '{print "mustParseCIDR("$0"),"}'
		nets: []*net.IPNet{
			mustParseCIDR("8.8.4.0/24"),
			mustParseCIDR("8.8.8.0/24"),
			mustParseCIDR("8.34.208.0/20"),
			mustParseCIDR("8.35.192.0/20"),
			mustParseCIDR("8.228.0.0/14"),
			mustParseCIDR("8.232.0.0/14"),
			mustParseCIDR("8.236.0.0/15"),
			mustParseCIDR("23.236.48.0/20"),
			mustParseCIDR("23.251.128.0/19"),
			mustParseCIDR("34.0.0.0/15"),
			mustParseCIDR("34.2.0.0/16"),
			mustParseCIDR("34.3.0.0/23"),
			mustParseCIDR("34.3.3.0/24"),
			mustParseCIDR("34.3.4.0/24"),
			mustParseCIDR("34.3.8.0/21"),
			mustParseCIDR("34.3.16.0/20"),
			mustParseCIDR("34.3.32.0/19"),
			mustParseCIDR("34.3.64.0/18"),
			mustParseCIDR("34.4.0.0/14"),
			mustParseCIDR("34.8.0.0/13"),
			mustParseCIDR("34.16.0.0/12"),
			mustParseCIDR("34.32.0.0/11"),
			mustParseCIDR("34.64.0.0/10"),
			mustParseCIDR("34.128.0.0/10"),
			mustParseCIDR("35.184.0.0/13"),
			mustParseCIDR("35.192.0.0/14"),
			mustParseCIDR("35.196.0.0/15"),
			mustParseCIDR("35.198.0.0/16"),
			mustParseCIDR("35.199.0.0/17"),
			mustParseCIDR("35.199.128.0/18"),
			mustParseCIDR("35.200.0.0/13"),
			mustParseCIDR("35.208.0.0/12"),
			mustParseCIDR("35.224.0.0/12"),
			mustParseCIDR("35.240.0.0/13"),
			mustParseCIDR("35.252.0.0/14"),
			mustParseCIDR("64.15.112.0/20"),
			mustParseCIDR("64.233.160.0/19"),
			mustParseCIDR("66.102.0.0/20"),
			mustParseCIDR("66.249.64.0/19"),
			mustParseCIDR("70.32.128.0/19"),
			mustParseCIDR("72.14.192.0/18"),
			mustParseCIDR("74.114.24.0/21"),
			mustParseCIDR("74.125.0.0/16"),
			mustParseCIDR("104.154.0.0/15"),
			mustParseCIDR("104.196.0.0/14"),
			mustParseCIDR("104.237.160.0/19"),
			mustParseCIDR("107.167.160.0/19"),
			mustParseCIDR("107.178.192.0/18"),
			mustParseCIDR("108.59.80.0/20"),
			mustParseCIDR("108.170.192.0/18"),
			mustParseCIDR("108.177.0.0/17"),
			mustParseCIDR("130.211.0.0/16"),
			mustParseCIDR("136.22.2.0/23"),
			mustParseCIDR("136.22.4.0/23"),
			mustParseCIDR("136.22.8.0/22"),
			mustParseCIDR("136.22.160.0/20"),
			mustParseCIDR("136.22.176.0/21"),
			mustParseCIDR("136.22.184.0/23"),
			mustParseCIDR("136.22.186.0/24"),
			mustParseCIDR("136.23.48.0/20"),
			mustParseCIDR("136.23.64.0/18"),
			mustParseCIDR("136.64.0.0/11"),
			mustParseCIDR("136.107.0.0/16"),
			mustParseCIDR("136.108.0.0/14"),
			mustParseCIDR("136.112.0.0/13"),
			mustParseCIDR("136.120.0.0/22"),
			mustParseCIDR("136.124.0.0/15"),
			mustParseCIDR("142.250.0.0/15"),
			mustParseCIDR("146.148.0.0/17"),
			mustParseCIDR("162.120.128.0/17"),
			mustParseCIDR("162.216.148.0/22"),
			mustParseCIDR("162.222.176.0/21"),
			mustParseCIDR("172.110.32.0/21"),
			mustParseCIDR("172.217.0.0/16"),
			mustParseCIDR("172.253.0.0/16"),
			mustParseCIDR("173.194.0.0/16"),
			mustParseCIDR("173.255.112.0/20"),
			mustParseCIDR("192.104.160.0/23"),
			mustParseCIDR("192.158.28.0/22"),
			mustParseCIDR("192.178.0.0/15"),
			mustParseCIDR("193.186.4.0/24"),
			mustParseCIDR("199.36.154.0/23"),
			mustParseCIDR("199.36.156.0/24"),
			mustParseCIDR("199.192.112.0/22"),
			mustParseCIDR("199.223.232.0/21"),
			mustParseCIDR("207.175.0.0/16"),
			mustParseCIDR("207.223.160.0/20"),
			mustParseCIDR("208.65.152.0/22"),
			mustParseCIDR("208.68.108.0/22"),
			mustParseCIDR("208.81.188.0/22"),
			mustParseCIDR("208.117.224.0/19"),
			mustParseCIDR("209.85.128.0/17"),
			mustParseCIDR("216.58.192.0/19"),
			mustParseCIDR("216.73.80.0/20"),
			mustParseCIDR("216.239.32.0/19"),
			mustParseCIDR("216.252.220.0/22"),
			mustParseCIDR("2001:4860::/32"),
			mustParseCIDR("2404:6800::/32"),
			mustParseCIDR("2404:f340::/32"),
			mustParseCIDR("2600:1900::/28"),
			mustParseCIDR("2605:ef80::/32"),
			mustParseCIDR("2606:40::/32"),
			mustParseCIDR("2606:73c0::/32"),
			mustParseCIDR("2607:1c0:241:40::/60"),
			mustParseCIDR("2607:1c0:300::/40"),
			mustParseCIDR("2607:f8b0::/32"),
			mustParseCIDR("2620:11a:a000::/40"),
			mustParseCIDR("2620:120:e000::/40"),
			mustParseCIDR("2800:3f0::/32"),
			mustParseCIDR("2a00:1450::/32"),
			mustParseCIDR("2c0f:fb50::/32"),
		},
	})
}

type GoogleDetector struct {
	accurateName string
	nets         []*net.IPNet
}

func (d *GoogleDetector) Name() string {
	if d.accurateName != "" {
		return d.accurateName
	}
	return "Google"
}

func (d *GoogleDetector) Detect(endpoint *models.ApplicationEndpoint) (bool, error) {
	match, err := matchIPRange(endpoint, d.nets)
	if err != nil {
		return false, err
	}
	if match {
		switch endpoint.Port {
		case 993, 995, 465, 587:
			d.accurateName = "GMail"
		case 53:
			d.accurateName = "Google DNS"
		default:
			// reset
			d.accurateName = ""
		}
	}
	return match, nil
}
