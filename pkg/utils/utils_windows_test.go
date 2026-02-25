//go:build windows

package utils

import (
	"os/exec"
	"testing"
)

func TestGetCmd(t *testing.T) {
	exe := "cmd.exe"
	args := []string{"/c", "timeout 30"}

	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		t.Errorf("error while starting command: %v\n", err)
	}

	defer cmd.Process.Kill()

	cmdline, err := GetCmd(cmd.Process.Pid)
	t.Logf("CMDLINE: %v", cmdline)
	if err != nil {
		t.Fatal(err)
	}
	if cmdline[0] != exe {
		t.Errorf("bad exe name: %v != %v", cmdline[0], exe)
	}
	for i, arg := range cmdline[1:] {
		if arg != args[i] {
			t.Errorf("bad command line, expect %s, got %s", args[i], arg)
		}
	}
}
