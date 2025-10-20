package modules

import "testing"

var banners = []string{
	"SSH-2.0-OpenSSH_7.9p1 Raspbian-10+deb10u4",            // raspbian 10
	"SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3",               // debian 11
	"SSH-2.0-OpenSSH_for_Windows_7.7",                      // server 2016
	"SSH-2.0-OpenSSH_for_Windows_8.1",                      // server 2019
	"SSH-2.0-OpenSSH_for_Windows_9.8 Win32-OpenSSH-GitHub", // server 2022
	"SSH-2.0-OpenSSH_9.6p1 Ubuntu-3ubuntu13.8",             // ubuntu 24.04
	"SSH-2.0-OpenSSH_9.3",                                  // Fedora 40
	"SSH-2.0-OpenSSH_9.6",                                  // Fedora 40
	"SSH-2.0-OpenSSH_9.9",                                  // Fedora 41 and Fedora 42
	"SSH-2.0-OpenSSH_8.7",                                  // rocky linux 9
	"SSH-2.0-OpenSSH_8.0",                                  // rocky linux 8
}

func TestParseOpenSSHBanner(t *testing.T) {
	for _, banner := range banners {
		out := parseOpenSSHBanner(banner)
		t.Log(out)
		if out.Product != "OpenSSH" {
			t.Errorf("OpenSSH product not found in banner %s", banner)
		}
		if out.Version == "" {
			t.Errorf("version not found in banner %s", banner)
		}
	}
}
