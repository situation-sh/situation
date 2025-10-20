package perf

import "testing"

func TestCollect(t *testing.T) {
	perfs := Collect()
	t.Log(perfs)
}
