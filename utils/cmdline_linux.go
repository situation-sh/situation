//go:build linux
// +build linux

package utils

import (
	"bytes"
	"fmt"
	"os"
)

func GetCmd(pid int) ([]string, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("the PID is not strictly positive")
	}

	p := fmt.Sprintf("/proc/%d/cmdline", pid)

	if !FileExists(p) {
		return nil, fmt.Errorf("cannot retrieve cmdline file for process with pid=%d", pid)
	}

	// read the file
	buffer, err := os.ReadFile(p) // #nosec G304 -- False positive
	if err != nil {
		return nil, err
	}

	// in general arguments are separated by null bytes
	// so we convert it to spaces first
	slices := bytes.Split(
		bytes.ReplaceAll(buffer, []byte{0}, []byte{32}),
		[]byte{32},
	)

	// specify max capacity to len(slices)
	out := make([]string, 0, len(slices))
	for _, b := range slices {
		if len(b) > 0 {
			out = append(out, string(b))
		}
	}
	return out, nil
}
