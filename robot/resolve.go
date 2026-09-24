package robot

import (
	"hash/fnv"
	"math/rand/v2"
)

// hslRange bounds the saturation and lightness of a random color so random
// robots stay pleasant: pastel backgrounds, mid-tone bodies, vivid glow.
type hslRange struct{ satMin, satMax, lightMin, lightMax float64 }

var colorRanges = map[string]hslRange{
	"body":       {0.25, 0.60, 0.55, 0.75},
	"accent":     {0.60, 0.90, 0.50, 0.62},
	"glow":       {0.85, 1.00, 0.55, 0.65},
	"background": {0.20, 0.45, 0.85, 0.93},
}

// Resolve returns partial with every unset facet filled in. Each facet draws
// from its own PRNG stream seeded by (seed, facet name), so pinning one facet
// never changes another facet's random value.
func Resolve(partial Spec, seed uint64) Spec {
	s := partial
	for _, f := range s.fields() {
		if f.isSet() {
			continue
		}
		r := facetRand(seed, f.name)
		switch f.kind {
		case KindEnum:
			*f.str = f.values[r.IntN(len(f.values))]
		case KindBool:
			b := r.IntN(2) == 0
			*f.flag = &b
		case KindColor:
			cr := colorRanges[f.name]
			h := r.Float64() * 360
			sat := cr.satMin + (cr.satMax-cr.satMin)*r.Float64()
			light := cr.lightMin + (cr.lightMax-cr.lightMin)*r.Float64()
			*f.str = hexString(hsl(h, sat, light))
		}
	}
	return s
}

func facetRand(seed uint64, name string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(name))
	return rand.New(rand.NewPCG(seed, h.Sum64()))
}

// Resolve applies the palette to unset colors, picks a fresh seed if none was
// given, and fills every remaining unset facet. It returns the resolved Spec
// and the seed used.
func (r Request) Resolve() (Spec, uint64) {
	seed := rand.Uint64()
	if r.Seed != nil {
		seed = *r.Seed
	}
	s := r.Spec
	if p, ok := palettes[r.Palette]; ok {
		for _, c := range []struct {
			dst *string
			src string
		}{
			{&s.Body, p.Body}, {&s.Accent, p.Accent}, {&s.Glow, p.Glow},
		} {
			if *c.dst == "" {
				*c.dst = c.src
			}
		}
	}
	return Resolve(s, seed), seed
}
