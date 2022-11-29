//go:build windows
// +build windows

package utils

import (
	"strings"

	"github.com/winlabs/gowin32"
)

func GetCmd(pid int) ([]string, error) {
	buffer, err := gowin32.GetProcessCommandLine(uint(pid))
	if err != nil {
		return nil, err
	}
	// arguments are separated by spaces
	slices := strings.Split(buffer, " ")
	// specify max capacity to len(slices)
	out := make([]string, 0, len(slices))
	for _, s := range slices {
		if len(s) > 0 {
			out = append(out, s)
		}
	}
	return out, nil
}
