package utils

import (
	"encoding/binary"
	"math"
	"net"
	"os"
	"sort"
	"testing"
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

func TestRandomTCPPort(t *testing.T) {
	var a uint16 = 10000
	var b uint16 = 50000

	for i := 0; i < 100000; i++ {
		p := RandomTCPPort(a, b)
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
