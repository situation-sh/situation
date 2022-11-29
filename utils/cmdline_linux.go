//go:build linux
// +build linux

package utils

import (
	"bytes"
	"fmt"
	"os"
)

func GetCmd(pid int) ([]string, error) {
	p := fmt.Sprintf("/proc/%d/cmdline", pid)
	if !FileExists(p) {
		return nil, fmt.Errorf("cannot retrieve cmdline file for process with pid=%d", pid)
	}

	// read the file
	buffer, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	// arguments are separated by null bytes
	slices := bytes.Split(buffer, []byte{0})
	// specify max capacity to len(slices)
	out := make([]string, 0, len(slices))
	for _, b := range slices {
		if len(b) > 0 {
			out = append(out, string(b))
		}
	}
	return out, nil
}
