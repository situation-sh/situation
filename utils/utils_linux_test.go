//go:build linux
// +build linux

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func TestGetCmd(t *testing.T) {
	exe := "/usr/bin/sleep"
	args := []string{"3000000"}

	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		t.Errorf("error while starting command: %v\n", err)
	}
	t.Logf("X: %#+v\n", cmd.Process)
	b, e := os.ReadFile(fmt.Sprintf("/proc/%d/status", cmd.Process.Pid))
	t.Logf("ERR: %v, BUFFER: %v\n", e, string(b))

	ok := false
	for !ok {
		p, err := process.NewProcess(int32(cmd.Process.Pid))
		if err != nil {
			continue
		}
		s, err := p.Status()
		if err != nil || len(s) == 0 {
			continue
		}
		ok = s[0] == "running" || s[0] == "sleep"
		t.Log(s)
		time.Sleep(time.Second)
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
