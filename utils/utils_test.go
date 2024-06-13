package utils

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func ksStat(data []int, max int, points int) float64 {
	// Kolmogorov Smirnov Test (w.r.t. Uniform Law)

	// sort data
	sort.Ints(data)

	// KS Statistic
	S := 0.0
	length := float64(len(data))
	step := max / (points - 1)
	for i := step; i < max; i += step {
		th := float64(i) / float64(max)
		f := float64(sort.SearchInts(data, i)) / length
		d := math.Abs(f - th)
		if d > S {
			S = d
		}
	}

	return S
}

func TestRandUint16(t *testing.T) {
	var max uint16 = 65535
	n := 10 * max
	data := make([]int, n)
	for i := range data {
		data[i] = int(RandUint16(max))
	}
	S := ksStat(data, int(max), 20)
	t.Log(S)
	if S > 0.01 {
		t.Errorf("The random number generator is suffering, S = %.3f", S)
	}
}

func TestFallback(t *testing.T) {
	max := 255
	n := 20 * max
	data := make([]byte, n)
	intData := make([]int, n)
	fallbackFillRandom(data)
	for i, b := range data {
		intData[i] = int(b)
	}
	S := ksStat(intData, int(max), 20)
	t.Log(S)
	if S > 0.015 {
		t.Errorf("The random number generator is suffering, S = %.3f", S)
	}
}

func TestRandomTCPPort(t *testing.T) {
	var a uint16 = 10000
	var b uint16 = 50000
	var p uint16

	for i := 0; i < 100000; i++ {
		if i%2 == 0 {
			p = RandomTCPPort(a, b)
		} else {
			p = RandomTCPPort(b, a)
		}

		if p < a || p >= b {
			t.Errorf("Generated port (%d) is out of bound [%d, %d[", p, a, b)
		}

	}
}

func TestCopyIP(t *testing.T) {
	ip := net.IPv4(192, 168, 0, 1).To4()
	ip2 := CopyIP(ip)
	for i := range ip {
		if ip[i] != ip2[i] {
			t.Errorf("Bad copy, expect %d, got %d", ip[i], ip2[i])
		}
	}

	// try to modify
	ip[0] = 0
	ip[1] = 1
	ip[2] = 2
	ip[3] = 3
	for i := range ip.To4() {
		if ip[i] == ip2[i] {
			t.Error("Bad copy")
		}
	}
}

func TestIterate(t *testing.T) {
	tot := 32
	ones := 24

	network := &net.IPNet{
		IP:   net.IPv4(192, 168, 3, 10),
		Mask: net.CIDRMask(ones, tot),
	}

	// test range and number of generated IP
	i := 0
	for ip := range Iterate(network) {
		// t.Log(ip)
		i++
		if !network.Contains(ip) {
			t.Errorf("The network %s does not contain %s", network.String(), ip.String())
		}
	}

	if i != 1<<(tot-ones) {
		t.Errorf("The number of IP is bad, expect %d, got %d", 1<<(tot-ones), i)
	}

	// test that we have all the IP
	m := make(map[uint32]bool)
	ones = 20
	network = &net.IPNet{
		IP:   net.IPv4(192, 168, 1, 1),
		Mask: net.CIDRMask(ones, tot),
	}

	// test range and number of generated IP
	for ip := range Iterate(network) {
		// t.Log(ip)
		index := binary.BigEndian.Uint32(ip)
		m[index] = true
	}

	if len(m) != 1<<(tot-ones) {
		t.Errorf("The number of IP is bad, expect %d, got %d", 1<<(tot-ones), len(m))
	}
}

func TestExtractNetworks(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	for i := range ifaces {
		networks := ExtractNetworks(&ifaces[i], true)
		if len(networks) == 0 {
			continue
		}
		for _, network := range networks {
			if (network.IP.To4() != nil) && (network.IP.To4()[3] == 0) {
				t.Errorf("The IPv4 address ends with a 0")
			}
		}
	}
}

func TestIsReserved(t *testing.T) {
	a := net.IPv4(192, 168, 0, 1)
	b := net.IPv4(8, 8, 8, 8)
	c := net.IPv4(192, 168, 0, 0)
	d := net.IPv4(192, 168, 0, 255)
	e := net.IPv6zero
	var f net.IP = nil

	for _, ip := range []net.IP{a, b, e, f} {
		if IsReserved(ip) {
			t.Errorf("the ip %v must not be reserved", ip)
		}
	}

	for _, ip := range []net.IP{c, d} {
		if !IsReserved(ip) {
			t.Errorf("the ip %v must be reserved", ip)
		}
	}

	// if ip4 := ip.To4(); ip4 != nil {
	// 	return ip4[3] == 0 || ip4[3] == 255
	// }
	// return false
}

func randomString(size int) string {
	bytes := RandBytes(size)
	for i, b := range bytes {
		// all the bytes are between 48 (0) and 126 (~)
		bytes[i] = 48 + (b % (127 - 48))
	}
	return string(bytes)
}
func TestFileExists(t *testing.T) {
	file := "/" + randomString(64)
	if FileExists(file) {
		t.Errorf("The file %s must not exist", file)
	}

	f, err := os.Executable()
	if err != nil {
		t.Error(err)
	}
	if !FileExists(f) {
		t.Errorf("The file %s must exist", f)
	}
}

func TestKeepLeaves(t *testing.T) {
	files := []string{
		"/d/e/f/g",
		"/d/f",
		"/d/e/f",
		"/a",
		"/a/b/c",
		"/a/c",
		"/b/c/d",
		"/b",
		"/b/c",
		"/c",
	}
	expect := []string{
		"/d/f",
		"/d/e/f/g",
		"/c",
		"/b/c/d",
		"/a/c",
		"/a/b/c",
	}
	out := KeepLeaves(files)

	if len(out) != len(expect) {
		t.Errorf("output has not the right size. Expected results: %v (got %v)",
			expect, out)
	}
	for i, e := range out {
		if expect[i] != e {
			t.Errorf("bad pruning, expect %s, got %s", expect[i], e)
		}
	}
}

func TestKeepLeavesError(t *testing.T) {
	files := []string{}
	out := KeepLeaves(files)
	if len(out) != 0 {
		t.Errorf("Output must be empty: %v", out)
	}
}

func TestIncludes(t *testing.T) {
	slice := []string{"a", "b", "c", "d"}
	notslice := []string{"e", "f", "g", "h"}

	for _, s := range slice {
		if !Includes(slice, s) {
			t.Errorf("the element %s is not in %v", s, slice)
		}
	}

	for _, s := range notslice {
		if Includes(slice, s) {
			t.Errorf("the element %s is in %v", s, slice)
		}
	}
}

func TestGetKeys(t *testing.T) {
	trueKeys := []string{"a", "b", "c", "d"}
	m := map[string]interface{}{
		"a": 0,
		"b": time.Now(),
		"c": "xxx",
		"d": []string{"o", "oo", "ooo"},
	}
	keys := GetKeys(m)
	if len(keys) != len(trueKeys) {
		t.Errorf("bad array length (expect %d): %v", len(trueKeys), keys)
	}

	for _, k := range trueKeys {
		if !Includes(keys, k) {
			t.Errorf("keys does not include %s: %v", k, keys)
		}
	}
}

func TestGetLines(t *testing.T) {
	f, err := os.CreateTemp("", "getlines")
	if err != nil {
		t.Error(err)
	}
	fp := f.Name()
	// close and remove that files
	// in the end
	defer func() {
		os.Remove(fp)
	}()

	content := "line"
	for i := 1; i <= 10; i++ {
		f.WriteString(fmt.Sprintf("%s%d\n", content, i))
	}
	f.Close()

	lines, err := GetLines(fp)
	if err != nil {
		t.Error(err)
	}
	for i, line := range lines {
		expected := fmt.Sprintf("%s%d", content, i+1)
		if line != expected {
			t.Errorf("bad line, expect %s, got %s", expected, line)
		}
	}

	trimNumber := func(s string) string {
		out := ""
		for _, r := range s {
			if r >= 48 && r < 58 {
				continue
			}
			out += string(r)
		}
		return out
	}

	lines, err = GetLines(fp, trimNumber)
	if err != nil {
		t.Error(err)
	}
	for _, line := range lines {
		if line != content {
			t.Errorf("bad line, expect %s, got %s", content, line)
		}
	}

}

func TestGetLinesError(t *testing.T) {
	if out, err := GetLines("/file/not/found"); err == nil {
		t.Errorf("Error must raise, got non nil output: %v", out)
	}

}

func TestGetCmd(t *testing.T) {
	exe := "/usr/bin/sleep"
	args := []string{"3000000"}
	if runtime.GOOS == "windows" {
		exe = "cmd.exe"
		args = []string{"/c", "'timeout 30'"}
	}
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

func TestGetCmdErrors(t *testing.T) {
	if cmdline, err := GetCmd(-5); err == nil {
		t.Errorf("error must raise, got %v", cmdline)
	}

	if cmdline, err := GetCmd(1<<32 - 1); err == nil {
		t.Errorf("error must raise, got %v", cmdline)
	}
}

func TestFlags(t *testing.T) {
	type unknown struct{}

	flags := map[string]interface{}{
		"bool":        true,
		"int":         int(0),
		"int64":       int64(0),
		"string":      "s",
		"stringslice": []string{"a", "b"},
		"duration":    time.Second,
		"other":       unknown{},
	}

	for k, v := range flags {
		BuildFlag(k, v, "", nil)
	}

	for k, v := range BuiltFlags() {
		switch value := v.(type) {
		case []string:
			f, ok := flags[k].([]string)
			if !ok {
				t.Errorf("bad flag type: %v", flags[k])
			}
			for i, e := range value {
				if f[i] != e {
					t.Errorf("bad flags, expected %v, got %v", flags[k], v)
				}
			}
		default:
			if flags[k] != v {
				t.Errorf("bad flags, expected %v, got %v", flags[k], v)
			}
		}
	}

}

func TestEnforceMask(t *testing.T) {
	ipnet := net.IPNet{
		IP:   net.ParseIP("192.168.1.68"),
		Mask: net.CIDRMask(24, 32),
	}
	out := EnforceMask(&ipnet)
	truth := net.ParseIP("192.168.1.0")
	if !out.IP.Equal(truth) {
		t.Errorf("%v != %v", out.IP, truth)
	}

	ipnet6 := net.IPNet{
		IP:   net.ParseIP("fe80::b636:f478:575f:59a6"),
		Mask: net.CIDRMask(64, 128),
	}
	out6 := EnforceMask(&ipnet6)
	truth6 := net.ParseIP("fe80::")
	if !out6.IP.Equal(truth6) {
		t.Errorf("%v != %v", out6.IP, truth6)
	}

}
