//go:build windows
// +build windows

package pingconfig

// According to the authors of go-ping
// the pinger must be privileged on windows
// even if we do not run the agent as admin/root
func UseICMP() bool {
	return false
}
