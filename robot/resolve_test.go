package robot

import (
	"reflect"
	"testing"
)

func TestResolveFillsEverything(t *testing.T) {
	for seed := uint64(0); seed < 200; seed++ {
		s := Resolve(Spec{}, seed)
		for _, f := range s.fields() {
			if !f.isSet() {
				t.Fatalf("seed %d: %s unset", seed, f.name)
			}
		}
		if err := (Request{Spec: s, Size: DefaultSize}).Validate(); err != nil {
			t.Fatalf("seed %d: resolved spec invalid: %v", seed, err)
		}
		if s.Background == "none" {
			t.Fatalf("seed %d: random background must not be none", seed)
		}
	}
}

func TestResolveDeterministic(t *testing.T) {
	a, b := Resolve(Spec{}, 99), Resolve(Spec{}, 99)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("same seed differs:\n%+v\n%+v", a, b)
	}
}

func TestResolveKeepsPinned(t *testing.T) {
	no := false
	s := Resolve(Spec{Head: "dome", Rivets: &no, Body: "#123456", Background: "none"}, 5)
	if s.Head != "dome" || *s.Rivets || s.Body != "#123456" || s.Background != "none" {
		t.Errorf("pinned values changed: %+v", s)
	}
}

func TestResolvePerFacetIndependence(t *testing.T) {
	free := Resolve(Spec{}, 42)
	other := "square"
	if free.Head == other {
		other = "dome"
	}
	pinned := Resolve(Spec{Head: other}, 42)
	pinned.Head = free.Head
	if !reflect.DeepEqual(free, pinned) {
		t.Errorf("pinning head changed other facets:\n%+v\n%+v", free, pinned)
	}
}

func TestResolveSeedsVary(t *testing.T) {
	seen := map[string]bool{}
	for seed := uint64(1); seed <= 20; seed++ {
		seen[Resolve(Spec{}, seed).Query().Encode()] = true
	}
	if len(seen) != 20 {
		t.Errorf("20 seeds gave only %d distinct robots", len(seen))
	}
}

func TestRequestResolveSeed(t *testing.T) {
	_, a := Request{Size: DefaultSize}.Resolve()
	_, b := Request{Size: DefaultSize}.Resolve()
	if a == b {
		t.Errorf("unseeded requests reused seed %d", a)
	}
	seed := uint64(7)
	s, got := Request{Seed: &seed, Size: DefaultSize}.Resolve()
	if got != 7 || !reflect.DeepEqual(s, Resolve(Spec{}, 7)) {
		t.Errorf("seeded request: seed=%d spec=%+v", got, s)
	}
}

func TestRequestResolvePaletteAndOverride(t *testing.T) {
	seed := uint64(1)
	s, _ := Request{Spec: Spec{Accent: "#000000"}, Palette: "mint", Seed: &seed, Size: DefaultSize}.Resolve()
	mint := palettes["mint"]
	if s.Body != mint.Body || s.Glow != mint.Glow || s.Accent != "#000000" {
		t.Errorf("palette not applied correctly: %+v", s)
	}
}

func TestResolveShouldersRange(t *testing.T) {
	var neg, pos bool
	for seed := uint64(0); seed < 200; seed++ {
		v := *Resolve(Spec{}, seed).Shoulders
		if v < ShouldersMin || v > ShouldersMax {
			t.Fatalf("seed %d: shoulders = %d out of range", seed, v)
		}
		neg, pos = neg || v < 0, pos || v > 0
	}
	if !neg || !pos {
		t.Errorf("200 seeds never produced both spiky and rounded shoulders")
	}
}

func TestCellSeed(t *testing.T) {
	seen := map[uint64]bool{}
	for cell := 0; cell < 9; cell++ {
		s := CellSeed(42, cell)
		if s != CellSeed(42, cell) {
			t.Fatalf("cell %d: not deterministic", cell)
		}
		if s == 42 || seen[s] {
			t.Errorf("cell %d: seed %d repeats the grid seed or another cell", cell, s)
		}
		seen[s] = true
	}
	if CellSeed(42, 0) == CellSeed(43, 0) {
		t.Error("different grid seeds gave the same cell seed")
	}
}

func TestRequestResolveCell(t *testing.T) {
	seed, cell := uint64(42), 3
	s, got := Request{Seed: &seed, Cell: &cell, Size: DefaultSize}.Resolve()
	if got != CellSeed(42, 3) || !reflect.DeepEqual(s, Resolve(Spec{}, CellSeed(42, 3))) {
		t.Errorf("cell request resolved with seed %d, want %d", got, CellSeed(42, 3))
	}
}
