package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func tree(t *testing.T, paths ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range paths {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestFind(t *testing.T) {
	root := tree(t,
		"outside.tfvars",
		"repo/.git",
		"repo/b.tfvars",
		"repo/a.tfvars",
		"repo/svc/svc.tfvars",
		"repo/svc/dev/dev.tfvars",
		"repo/svc/dev/notes.txt",
	)
	dev := filepath.Join(root, "repo/svc/dev")
	abs := func(paths ...string) []string {
		out := make([]string, len(paths))
		for i, p := range paths {
			out[i] = filepath.Join(root, p)
		}
		return out
	}

	tests := []struct {
		name string
		s    search
		want []string
		top  string
	}{
		{"stops at marker, outermost first, sorted within a directory",
			search{".tfvars", -1, ".git"},
			abs("repo/a.tfvars", "repo/b.tfvars", "repo/svc/svc.tfvars", "repo/svc/dev/dev.tfvars"),
			"repo"},
		{"depth zero is the start directory", search{".tfvars", 0, ".git"},
			abs("repo/svc/dev/dev.tfvars"), "repo/svc/dev"},
		{"depth one climbs once", search{".tfvars", 1, ".git"},
			abs("repo/svc/svc.tfvars", "repo/svc/dev/dev.tfvars"), "repo/svc"},
		{"other extension", search{".txt", -1, ".git"},
			abs("repo/svc/dev/notes.txt"), "repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, top, unbounded, err := find(dev, tt.s)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("files = %v, want %v", got, tt.want)
			}
			if want := filepath.Join(root, tt.top); top != want {
				t.Errorf("top = %s, want %s", top, want)
			}
			if unbounded {
				t.Error("unexpected unbounded")
			}
		})
	}
}

func TestFindMissingMarkerStaysInStartDir(t *testing.T) {
	root := tree(t, "stray.tfvars", "dev/dev.tfvars")
	got, _, unbounded, err := find(filepath.Join(root, "dev"), search{".tfvars", -1, ".no-such-marker"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{filepath.Join(root, "dev/dev.tfvars")}; !reflect.DeepEqual(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
	if !unbounded {
		t.Error("want unbounded")
	}
}

func TestCliArgsQuotes(t *testing.T) {
	got := cliArgs([]string{"/a b/x.tfvars", "/it's/y.tfvars"})
	want := `-var-file='/a b/x.tfvars' -var-file='/it'\''s/y.tfvars'`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseTerraform(t *testing.T) {
	chdir, cmd, rest := parseTerraform([]string{"-chdir=svc/dev", "plan", "-out=p"})
	if chdir != "svc/dev" || cmd != "plan" || !reflect.DeepEqual(rest, []string{"-out=p"}) {
		t.Errorf("got %q %q %v", chdir, cmd, rest)
	}
}

func TestAppliesSavedPlan(t *testing.T) {
	tests := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"tfplan"}, true},
		{[]string{"-auto-approve"}, false},
		{[]string{"-auto-approve", "tfplan"}, true},
		{[]string{"-parallelism=2", "tfplan"}, true},
		{[]string{"-target", "aws_s3_bucket.a"}, false},
		{[]string{"-var-file", "x.tfvars"}, false},
	}
	for _, tt := range tests {
		if got := appliesSavedPlan(tt.args); got != tt.want {
			t.Errorf("%v: got %v, want %v", tt.args, got, tt.want)
		}
	}
}
