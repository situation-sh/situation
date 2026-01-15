//go:build linux

package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
	if len(buffer) == 0 {
		return nil, fmt.Errorf("cmdline is empty")
	}
	// in general arguments are separated by null bytes
	// so we convert it to spaces first
	spaceByte := []byte(" ")
	slices := bytes.Split(
		bytes.ReplaceAll(buffer, []byte{0}, spaceByte),
		spaceByte,
	)

	// specify max capacity to len(slices)
	out := make([]string, 0, len(slices))
	for _, b := range slices {
		if len(b) > 0 {
			out = append(out, string(b))
		}
	}

	if len(out) > 0 {
		p = fmt.Sprintf("/proc/%d/exe", pid)
		bin, err := filepath.EvalSymlinks(p)
		if err != nil {
			return nil, err
		}
		out[0] = bin
	}
	return out, nil
}
