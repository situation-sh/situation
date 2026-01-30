package saas

import (
	"fmt"
	"net"
	"strings"

	"github.com/situation-sh/situation/pkg/models"
)

type SaaSDetector interface {
	Name() string
	Detect(endpoint *models.ApplicationEndpoint) (bool, error)
}

// matchDNSSuffix checks whether any TLS DNS name on the endpoint
// matches one of the provided suffixes (exact match or suffix match).
func matchDNSSuffix(endpoint *models.ApplicationEndpoint, suffixes []string) bool {
	if endpoint.TLS == nil {
		return false
	}
	for _, dns := range endpoint.TLS.DNSNames {
		for _, suffix := range suffixes {
			if dns == suffix || strings.HasSuffix(dns, suffix) {
				return true
			}
		}
	}
	return false
}

// mustParseCIDR parses a CIDR string and panics if it fails.
// This is intended for use in init() blocks with hardcoded values.
func mustParseCIDR(cidr string) *net.IPNet {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Sprintf("invalid CIDR %q: %v", cidr, err))
	}
	return n
}

// matchIPRange checks whether the endpoint address falls within
// one of the provided IP networks. Returns an error if the address
// cannot be parsed as an IP.
func matchIPRange(endpoint *models.ApplicationEndpoint, nets []*net.IPNet) (bool, error) {
	if endpoint.Addr == "" {
		return false, nil
	}
	ip := net.ParseIP(endpoint.Addr)
	if ip == nil {
		return false, fmt.Errorf("cannot convert %v into net.IP", endpoint.Addr)
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}
