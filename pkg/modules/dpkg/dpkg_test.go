package dpkg

import "testing"

func TestParseLogLine(t *testing.T) {
	installed := []string{
		"2022-11-29 08:46:43 status installed linux-generic:amd64 5.15.0.53.53",
		"2022-11-29 08:46:43 status installed libc-bin:amd64 2.35-0ubuntu3.1",
		"2022-11-29 08:47:02 status installed initramfs-tools:all 0.140ubuntu13",
	}
	others := []string{
		"2022-11-29 08:46:43 status unpacked linux-generic:amd64 5.15.0.53.53",
		"2022-11-29 08:46:43 status half-configured linux-generic:amd64 5.15.0.53.53",
		"2022-11-29 08:46:43 trigproc libc-bin:amd64 2.35-0ubuntu3.1 <none>",
		"2022-11-29 08:46:43 status half-configured libc-bin:amd64 2.35-0ubuntu3.1",
		"2022-11-29 08:46:43 trigproc initramfs-tools:all 0.140ubuntu13 <none>",
		"2022-11-29 08:46:43 status half-configured initramfs-tools:all 0.140ubuntu13",
		"2022-11-29 08:47:02 trigproc linux-image-5.15.0-53-generic:amd64 5.15.0-53.59 <none>",
		"2022-11-29 08:47:02 status half-configured linux-image-5.15.0-53-generic:amd64 5.15.0-53.59",
	}
	for _, line := range installed {
		if parseLogLine(line) == nil {
			t.Errorf("cannot parse the following line: %s", line)
		}
	}
	for _, line := range others {
		if parseLogLine(line) != nil {
			t.Errorf("the following line must be rejected: %s", line)
		}
	}
}
