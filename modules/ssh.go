// LINUX(SSHModule) ok
// WINDOWS(SSHModule) ok
// MACDistribution(SSHModule) ?
// ROOT(SSHModule) yes
package modules

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/praetorian-inc/fingerprintx/pkg/plugins"
	"github.com/praetorian-inc/fingerprintx/pkg/plugins/services/ssh"
	"github.com/situation-sh/situation/store"
	"github.com/situation-sh/situation/utils"
)

func init() {
	RegisterModule(&SSHModule{})
}

// Module definition ---------------------------------------------------------

// SSHModule aims to ...
type SSHModule struct{}

func (m *SSHModule) Name() string {
	return "ssh"
}

func (m *SSHModule) Dependencies() []string {
	return []string{"tcp-scan"}
}

func (m *SSHModule) Run() error {
	logger := GetLogger(m)
	p := &ssh.SSHPlugin{}
	s := plugins.ServiceSSH{}
	machines, apps, endpoints := store.GetMachinesByOpenTCPPort(22)
	for k, endpoint := range endpoints {
		if ipv4 := endpoint.Addr.To4(); ipv4 != nil {
			address := net.JoinHostPort(endpoint.Addr.String(), fmt.Sprintf("%d", endpoint.Port))
			conn, err := net.Dial("tcp", address)
			if err != nil {
				logger.Errorf("fail to dial %s: %v", address, err)
				continue
			}
			service, err := p.Run(conn, time.Second, plugins.Target{})
			if err != nil {
				logger.Errorf("fail to run ssh discovery on %s: %v", address, err)
				continue
			}

			if err := json.Unmarshal(service.Raw, &s); err != nil {
				logger.Errorf("fail to unmarshal discovery results: %v", err)
				continue
			}

			apps[k].Protocol = "ssh"
			apps[k].Config = map[string]interface{}{
				"host_key":              s.HostKey,
				"host_key_fingerprint":  s.HostKeyFingerprint,
				"host_key_type":         s.HostKeyType,
				"algorithms":            parseSSHAlgorithm(s.Algo),
				"password_auth_enabled": s.PasswordAuthEnabled,
			}
			logger.WithField("host_key", s.HostKey).
				WithField("host_key_fingerprint", s.HostKeyFingerprint).
				WithField("host_key_type", s.HostKeyType).
				// WithField("algorithm", s.Algo).
				WithField("password_auth_enabled", s.PasswordAuthEnabled).
				Info("SSH service found")

			if s.Banner != "" {
				banner := analyzeBanner(s.Banner)
				msg := logger

				if machines[k].Platform == "" {
					machines[k].Platform = banner.Platform
					if banner.Platform != "" {
						msg = msg.WithField("platform", banner.Platform)
					}
				}
				if machines[k].Distribution == "" {
					machines[k].Distribution = banner.Distribution
					if banner.Distribution != "" {
						msg = msg.WithField("distribution", banner.Distribution)
					}
				}
				if machines[k].DistributionVersion == "" {
					machines[k].DistributionVersion = banner.DistributionVersion
					if banner.DistributionVersion != "" {
						msg = msg.WithField("distribution_version", banner.DistributionVersion)
					}
				}
				if apps[k].Name == "" {
					apps[k].Name = banner.Product
					if banner.Product != "" {
						msg = msg.WithField("name", banner.Product)
					}
				}
				if apps[k].Version == "" {
					apps[k].Version = banner.Version
					if banner.Version != "" {
						msg = msg.WithField("version", banner.Version)
					}
				}
				if apps[k].CPE == "" {
					apps[k].CPE = banner.CPE
					if banner.CPE != "" {
						msg = msg.WithField("cpe", banner.CPE)
					}
				}

				msg.Info("SSH banner analyzed")
			}

		}

	}
	return nil
}

// Ubuntu versions mapped to OpenSSH builds
var ubuntuVersions = map[string]string{
	"9.7p1-7":    "24.10", // Ubuntu 24.10 Oracular Oriole
	"9.6p1-3":    "24.04", // Ubuntu 24.04 Noble Numbat
	"9.3p1-1":    "23.10", // Ubuntu 23.10 Mantic Minotaur
	"9.0p1-1":    "22.10", // Ubuntu 22.10 Kinetic Kudu or Ubuntu 23.04 Lunar Lobster
	"8.9p1-3":    "22.04", // Ubuntu 22.04 Jammy Jellyfish
	"8.4p1-6":    "21.10", // Ubuntu 21.10 Impish Indri
	"8.4p1-5":    "21.04", // Ubuntu 21.04 Hirsute Hippo
	"8.3p1-1":    "20.10", // Ubuntu 20.10 Groovy Gorilla
	"8.2p1-4":    "20.04", // Ubuntu 20.04 Focal Fossa
	"8.0p1-6":    "19.10", // Ubuntu 19.10 Eoan Ermine
	"7.9p1-10":   "19.04", // Ubuntu 19.04 Disco Dingo
	"7.7p1-4":    "18.10", // Ubuntu 18.10 Cosmic Cuttlefish
	"7.6p1-4":    "18.04", // Ubuntu 18.04 Bionic Beaver
	"7.5p1-10":   "17.10", // Ubuntu 17.10 Artful Aardvark
	"7.4p1-10":   "17.04", // Ubuntu 17.04 Zesty Zapus
	"7.3p1-1":    "16.10", // Ubuntu 16.10 Yakkety Yak
	"7.2p2-4":    "16.04", // Ubuntu 16.04 Xenial Xerus
	"6.9p1-2":    "15.10", // Ubuntu 15.10 Wily Werewolf
	"6.7p1-5":    "15.04", // Ubuntu 15.04 Vivid Vervet
	"6.6.1p1-8":  "14.10", // Ubuntu 14.10 Utopic Unicorn
	"6.6.1p1-2":  "14.04", // Ubuntu 14.04 Trusty Tahr
	"6.2p2-6":    "13.10", // Ubuntu 13.10 Saucy Salamander
	"6.1p1-4":    "13.04", // Ubuntu 13.04 Raring Ringtail
	"6.0p1-3":    "12.10", // Ubuntu 12.10 Quantal Quetzal
	"5.9p1-5":    "12.04", // Ubuntu 12.04 Precise Pangolin
	"5.8p1-7":    "11.10", // Ubuntu 11.10 Oneiric Ocelot
	"5.8p1-1":    "11.04", // Ubuntu 11.04 Natty Narwhal
	"5.5p1-4":    "10.10", // Ubuntu 10.10 Maverick Meerkat
	"5.3p1-3":    "10.04", // Ubuntu 10.04 Lucid Lynx
	"5.1p1-6":    "9.10",  // Ubuntu 9.10 Karmic Koala
	"5.1p1-5":    "9.04",  // Ubuntu 9.04 Jaunty Jackalope
	"5.1p1-3":    "8.10",  // Ubuntu 8.10 Intrepid Ibex
	"4.7p1-8":    "8.04",  // Ubuntu 8.04 Hardy Heron
	"4.6p1-5":    "7.10",  // Ubuntu 7.10 Gutsy Gibbon
	"4.3p2-8":    "7.04",  // Ubuntu 7.04 Feisty Fawn
	"4.3p2-5":    "6.10",  // Ubuntu 6.10 Edgy Eft
	"4.2p1-7":    "6.06",  // Ubuntu 6.06 Dapper Drake
	"4.1p1-7":    "5.10",  // Ubuntu 5.10 Breezy Badger
	"3.9p1-1":    "5.04",  // Ubuntu 5.04 Hoary Hedgehog
	"3.8.1p1-11": "4.10",  // Ubuntu 4.10 Warty Warthog
}

// Debian versions mapped to OpenSSH builds
var debianVersions = map[string]string{
	"9.2p1-2":   "12", // "Debian 12.x \"Bookworm\""
	"8.4p1-5":   "11", // "Debian 11.x \"Bullseye\""
	"7.9p1-10":  "10", // "Debian 10.x \"Buster\""
	"7.4p1-10":  "9",  // "Debian 9.x \"Buster\""
	"7.4p-9":    "9",
	"6.7p1-5":   "8",
	"6.0p1-4":   "7",
	"6.0p1-2":   "7",
	"5.8p1-4":   "6",
	"5.5p1-6":   "6",
	"5.1p1-5":   "5",
	"4.3p2-9":   "4",
	"3.8.1p1-8": "3.1",
	"3.4p1-1":   "3.0",
}

// FreeBSD versions mapped to OpenSSH builds
var freebsdVersions = map[string]string{
	"20240806": "14.2", //"FreeBSD 14.2-RELEASE"
	"20240318": "14.1", //"FreeBSD 14.1-RELEASE"
	"20230316": "13.2", // "FreeBSD 13.2-RELEASE"
	"20200214": "12.2", //"FreeBSD 12.2-RELEASE"
	"20130515": "9.2",  // "FreeBSD 9.2-RELEASE"
}

// windows versions mapped to OpenSSH builds
var windowsVersions = map[string]string{
	"7.7": "Microsoft Windows Server 2016",
	"8.1": "Microsoft Windows Server 2019",
	"9.8": "Microsoft Windows Server 2022",
}

type SSHBanner struct {
	Product             string
	Version             string
	CPE                 string
	Platform            string
	Distribution        string
	DistributionVersion string
}

// Extracts OpenSSH version and build from banner
func parseOpenSSHBanner(banner string) *SSHBanner {
	out := SSHBanner{Product: "OpenSSH"}

	// windows case
	windowsRe := regexp.MustCompile(`OpenSSH_for_Windows_(\d+\.\d+)`)
	if windowsRe.MatchString(banner) {
		out.Platform = "windows"
		matches := windowsRe.FindStringSubmatch(banner)
		out.Version = matches[1]
		if version, exist := windowsVersions[out.Version]; exist {
			out.Distribution = version
		}
		return &out
	} else {
		out.Platform = "linux"
	}

	// Regular expressions to identify OpenSSH versions
	longVersionRegex := regexp.MustCompile(`OpenSSH[-_](\d+\.\d+\.\d+(p\d+)?)`)
	shortVersionRegex := regexp.MustCompile(`OpenSSH[-_](\d+\.\d+(p\d+)?)`)

	// Identify the OpenSSH version (long format first, then fallback to short format)
	if longVersionRegex.MatchString(banner) {
		matches := longVersionRegex.FindStringSubmatch(banner)
		out.Version = matches[1]
	} else if shortVersionRegex.MatchString(banner) {
		matches := shortVersionRegex.FindStringSubmatch(banner)
		out.Version = matches[1]
	} else {
		return &out
	}

	// split version along p
	chunks := strings.Split(out.Version, "p")
	if len(chunks) == 1 {
		out.CPE = fmt.Sprintf("cpe:2.3:a:openbsd:openssh:%s:-:*:*:*:*:*:*", chunks[0])
	} else {
		out.CPE = fmt.Sprintf("cpe:2.3:a:openbsd:openssh:%s:p%s:*:*:*:*:*:*", chunks[0], chunks[1])
	}

	fullVersion := ""
	// Calculate start offset to find the build number
	startOffset := strings.Index(banner, out.Version) + len(out.Version)

	// Extract build version (e.g., `-5` in `Debian-5+deb11u3`)
	buildVersionRegex := regexp.MustCompile(`-(\d+)`)
	if matches := buildVersionRegex.FindStringSubmatch(banner[startOffset:]); len(matches) > 1 {
		buildVersion := matches[1] // Extract build number
		fullVersion = fmt.Sprintf("%s-%s", out.Version, buildVersion)
	}

	if strings.Contains(banner, "Ubuntu") {
		out.Distribution = "ubuntu"
		if version, exists := ubuntuVersions[fullVersion]; exists {
			out.DistributionVersion = version
		}
	} else if strings.Contains(banner, "Debian") {
		out.Distribution = "debian"
		if version, exists := debianVersions[fullVersion]; exists {
			out.DistributionVersion = version
		}
	} else if strings.Contains(banner, "FreeBSD") {
		out.Distribution = "freebsd"
		re := regexp.MustCompile(`(\d{8})`)
		matches := re.FindStringSubmatch(banner)
		if len(matches) >= 2 {
			if version, exists := freebsdVersions[matches[1]]; exists {
				out.DistributionVersion = version
			}
		}
	} else if strings.Contains(banner, "Raspbian") {
		out.Distribution = "raspbian"
		if version, exists := debianVersions[fullVersion]; exists {
			out.DistributionVersion = version
		}
	}

	return &out
}

func analyzeBanner(banner string) *SSHBanner {
	out := SSHBanner{}
	bannerL := strings.ToLower(banner)
	if strings.Contains(bannerL, "openssh") {
		return parseOpenSSHBanner(banner)
	}

	return &out
}

// Function to parse the map-like string. Here are the map to reconstruct
//
//	info := map[string]string{
//		"Cookie":                  cookie,
//		"KexAlgos":                kexAlgos,
//		"ServerHostKeyAlgos":      serverHostKeyAlgos,
//		"CiphersClientServer":     ciphersClientServer,
//		"CiphersServerClient":     ciphersServerClient,
//		"MACsClientServer":        macClientServer,
//		"MACsServerClient":        macServerClient,
//		"CompressionClientServer": compressionClientServer,
//		"CompressionServerClient": compressionServerClient,
//		"LanguagesClientServer":   languagesClientServer,
//		"LanguagesServerClient":   languagesServerClient,
//	}
//
// This function cannot work in all cases. Ex: a := map[string]string{"a": "b", "c": "d or e", "f or g": "7"}
func parseSSHAlgorithm(algo string) map[string]interface{} {
	out := make(map[string]interface{})
	// Remove the surrounding `map[` and `]`
	trimmed := strings.TrimSuffix(strings.TrimPrefix(algo, "map["), "]")

	re := regexp.MustCompile(`[A-Z][a-zA-Z]+[:]`)
	// Find all matches with their start positions
	matches := re.FindAllStringIndex(trimmed, -1)
	for k, location := range matches {
		start := location[0]
		end := location[1]
		key := utils.ConvertCamelToSnake(trimmed[start : end-1])

		value := ""
		if k == len(matches)-1 {
			value = strings.TrimSpace(trimmed[location[1]:])
		} else {
			nextLocation := matches[k+1]
			value = strings.TrimSpace(trimmed[location[1]:nextLocation[0]])

		}

		values := strings.Split(value, ",")
		if len(values) == 1 {
			out[key] = values[0]
		} else {
			out[key] = values
		}
		// fmt.Printf("KEY: %s | VALUE: %v\n", key, value)
	}

	return out
}
