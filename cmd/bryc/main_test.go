package main

import (
	"bytes"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aethons-tools/bryc/robot"
)

func TestParseArgsEmpty(t *testing.T) {
	opts, err := parseArgs(nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(opts.req.Spec, robot.Spec{}) || opts.req.Seed != nil ||
		opts.req.Size != robot.DefaultSize || opts.out != "" {
		t.Errorf("got %+v", opts)
	}
}

func TestParseArgsPinned(t *testing.T) {
	opts, err := parseArgs([]string{
		"--head=dome", "--shoulders=-40", "--rivets=false", "--body=#112233", "--palette=mint",
		"--seed=42", "--size=128", "-o", "x.png",
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	r := opts.req
	if r.Spec.Head != "dome" || *r.Spec.Shoulders != -40 || *r.Spec.Rivets || r.Spec.Body != "#112233" ||
		r.Palette != "mint" || *r.Seed != 42 || r.Size != 128 || opts.out != "x.png" {
		t.Errorf("got %+v", opts)
	}
}

func TestParseArgsInvalid(t *testing.T) {
	_, err := parseArgs([]string{"--head=blob"}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), `head: unknown value "blob"`) {
		t.Errorf("err = %v", err)
	}
}

func TestRunWritesPNG(t *testing.T) {
	out := filepath.Join(t.TempDir(), "bot.png")
	var stderr bytes.Buffer
	if code := run([]string{"--seed=5", "--size=64", "-o", out}, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 64 {
		t.Errorf("width = %d", img.Bounds().Dx())
	}
	for _, want := range []string{"seed: 5\n", "spec: accent=%23", "wrote " + out} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr missing %q:\n%s", want, stderr.String())
		}
	}
}

func TestRunSameSeedSameBytes(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")
	run([]string{"--seed=9", "--size=64", "-o", a}, io.Discard)
	run([]string{"--seed=9", "--size=64", "-o", b}, io.Discard)
	ab, _ := os.ReadFile(a)
	bb, _ := os.ReadFile(b)
	if len(ab) == 0 || !bytes.Equal(ab, bb) {
		t.Error("same seed produced different files")
	}
}

func TestRunInvalidExitsNonZero(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"--size=1"}, &stderr); code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "size: 1 is outside 64-2048") {
		t.Errorf("stderr = %s", stderr.String())
	}
}
