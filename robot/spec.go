// Package robot models, randomizes and renders Bored Robots Yacht Club robots.
package robot

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Output size bounds, in pixels.
const (
	DefaultSize = 512
	MinSize     = 64
	MaxSize     = 2048
)

// Kind is the type of value a facet holds.
type Kind string

const (
	KindEnum  Kind = "enum"
	KindBool  Kind = "bool"
	KindColor Kind = "color"
)

// Allowed values for each enum facet.
var (
	HeadValues    = []string{"square", "rounded", "dome", "trapezoid"}
	EyesValues    = []string{"round", "visor", "cyclops", "led"}
	MouthValues   = []string{"grille", "speaker", "smile", "zigzag"}
	AntennaValues = []string{"none", "ball", "double", "bolt"}
	EarsValues    = []string{"none", "bolts", "dials"}
)

// Spec describes one robot. A zero-valued field is unset and gets randomized
// by Resolve; a resolved Spec has every field set.
type Spec struct {
	Head, Eyes, Mouth, Antenna, Ears string
	Rivets, Panels, Blush            *bool
	// Colors are "#rrggbb"; Background may also be "none" (transparent).
	Body, Accent, Glow, Background string
}

// field is a name-addressable view of one Spec field, so parsing,
// validation, encoding and randomization can loop over facets.
type field struct {
	name   string
	kind   Kind
	values []string // KindEnum only
	str    *string  // KindEnum and KindColor
	flag   **bool   // KindBool
}

func (s *Spec) fields() []field {
	return []field{
		{name: "head", kind: KindEnum, values: HeadValues, str: &s.Head},
		{name: "eyes", kind: KindEnum, values: EyesValues, str: &s.Eyes},
		{name: "mouth", kind: KindEnum, values: MouthValues, str: &s.Mouth},
		{name: "antenna", kind: KindEnum, values: AntennaValues, str: &s.Antenna},
		{name: "ears", kind: KindEnum, values: EarsValues, str: &s.Ears},
		{name: "rivets", kind: KindBool, flag: &s.Rivets},
		{name: "panels", kind: KindBool, flag: &s.Panels},
		{name: "blush", kind: KindBool, flag: &s.Blush},
		{name: "body", kind: KindColor, str: &s.Body},
		{name: "accent", kind: KindColor, str: &s.Accent},
		{name: "glow", kind: KindColor, str: &s.Glow},
		{name: "background", kind: KindColor, str: &s.Background},
	}
}

func (f field) isSet() bool {
	if f.kind == KindBool {
		return *f.flag != nil
	}
	return *f.str != ""
}

func (f field) value() string {
	if f.kind == KindBool {
		return strconv.FormatBool(**f.flag)
	}
	return *f.str
}

// Facet describes one customizable facet, for building UIs.
type Facet struct {
	Name   string   `json:"name"`
	Kind   Kind     `json:"kind"`
	Values []string `json:"values,omitempty"`
}

// Facets lists every facet in display order.
func Facets() []Facet {
	var out []Facet
	for _, f := range new(Spec).fields() {
		out = append(out, Facet{Name: f.name, Kind: f.kind, Values: f.values})
	}
	return out
}

// Query encodes the set fields of s as URL query values.
func (s Spec) Query() url.Values {
	q := url.Values{}
	for _, f := range s.fields() {
		if f.isSet() {
			q.Set(f.name, f.value())
		}
	}
	return q
}

// Request is everything a caller can ask for: a partial Spec plus the
// non-facet inputs.
type Request struct {
	Spec    Spec
	Palette string  // "" for none
	Seed    *uint64 // nil means pick a fresh seed
	Size    int
}

// ValidationError lists every problem found in a request.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "invalid robot options:\n  - " + strings.Join(e.Problems, "\n  - ")
}

// ParseQuery reads a Request from query values. Empty values and "random"
// mean unset. All parse and validation problems are returned together.
func ParseQuery(q url.Values) (Request, error) {
	req := Request{Size: DefaultSize}
	byName := map[string]field{}
	for _, f := range req.Spec.fields() {
		byName[f.name] = f
	}

	var problems []string
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		v := strings.TrimSpace(q.Get(key))
		if v == "" || v == "random" {
			continue
		}
		switch key {
		case "palette":
			req.Palette = v
		case "seed":
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				problems = append(problems, fmt.Sprintf("seed: %q is not a non-negative integer", v))
				continue
			}
			req.Seed = &n
		case "size":
			n, err := strconv.Atoi(v)
			if err != nil {
				problems = append(problems, fmt.Sprintf("size: %q is not an integer", v))
				continue
			}
			req.Size = n
		default:
			f, ok := byName[key]
			switch {
			case !ok:
				problems = append(problems, fmt.Sprintf("%s: unknown option", key))
			case f.kind == KindBool:
				b, err := strconv.ParseBool(v)
				if err != nil {
					problems = append(problems, fmt.Sprintf("%s: %q is not true or false", key, v))
					continue
				}
				*f.flag = &b
			case f.kind == KindColor:
				*f.str = strings.ToLower(v)
			default:
				*f.str = v
			}
		}
	}

	problems = append(problems, req.problems()...)
	if len(problems) > 0 {
		return Request{}, &ValidationError{Problems: problems}
	}
	return req, nil
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Validate reports every invalid value in r.
func (r Request) Validate() error {
	if p := r.problems(); len(p) > 0 {
		return &ValidationError{Problems: p}
	}
	return nil
}

func (r Request) problems() []string {
	var problems []string
	for _, f := range r.Spec.fields() {
		if !f.isSet() || f.kind == KindBool {
			continue
		}
		v := *f.str
		switch {
		case f.kind == KindEnum && !slices.Contains(f.values, v):
			problems = append(problems, fmt.Sprintf("%s: unknown value %q (allowed: %s)",
				f.name, v, strings.Join(f.values, ", ")))
		case f.kind == KindColor && !(f.name == "background" && v == "none") && !hexColor.MatchString(v):
			problems = append(problems, fmt.Sprintf("%s: %q is not a #rrggbb color", f.name, v))
		}
	}
	if _, ok := palettes[r.Palette]; r.Palette != "" && !ok {
		problems = append(problems, fmt.Sprintf("palette: unknown value %q (allowed: %s)",
			r.Palette, strings.Join(PaletteNames(), ", ")))
	}
	if r.Size < MinSize || r.Size > MaxSize {
		problems = append(problems, fmt.Sprintf("size: %d is outside %d-%d", r.Size, MinSize, MaxSize))
	}
	return problems
}
