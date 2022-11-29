package rpm

import "testing"

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
	out := keepLeaves(files)

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
