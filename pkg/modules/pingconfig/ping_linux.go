//go:build linux
// +build linux

package pingconfig

import (
	"os/user"
	"regexp"
	"strconv"
	"strings"

	sysctl "github.com/lorenzosaino/go-sysctl"
)

// Retrieve net.ipv4.ping_group_range
func getPingGroupRange() (int, int, error) {
	start := -1
	end := -1
	value, err := sysctl.Get("net.ipv4.ping_group_range")
	if err != nil {
		return start, end, err
	}

	value = strings.TrimSpace(value)
	re := regexp.MustCompile(`^([0-9]+)[ ]+([0-9]+)$`)
	subm := re.FindStringSubmatch(value)
	if len(subm) < 3 {
		return start, end, err
	}
	start, err = strconv.Atoi(subm[1])
	if err != nil {
		return start, end, err
	}
	end, err = strconv.Atoi(subm[2])
	return start, end, err
}

func isUserAllowedToPing(usr *user.User) (bool, error) {
	groups, err := usr.GroupIds()
	if err != nil {
		return false, err
	}
	start, end, err := getPingGroupRange()
	if err != nil {
		return false, err
	}

	if start > end {
		return false, nil
	}

	for gid := range groups {
		if (start <= gid) && (gid <= end) {
			return true, nil
		}
	}
	return false, nil
}

func UseICMP() bool {
	usr, err := user.Current()
	if err != nil {
		return false
	}

	// check sysctl
	allowed, err := isUserAllowedToPing(usr)
	if err != nil {
		return false
	}

	return !allowed
}
