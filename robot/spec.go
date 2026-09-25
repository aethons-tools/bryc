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
	KindRange Kind = "range" // an integer between a facet's Min and Max
)

// Shoulders range: 0 is square, positive rounds the corners off, negative
// grows a spike on each shoulder.
const (
	ShouldersMin = -100
	ShouldersMax = 100
)

// Face position range: 0 is neutral, -1 moves the face down toward the
// chin, +1 raises the eyes, +2 also raises the mouth under them (see face.go).
const (
	FaceMin = -1
	FaceMax = 2
)

// Allowed values for each enum facet.
var (
	HeadValues       = []string{"square", "rounded", "dome", "inverted-dome", "trapezoid", "inverted-trapezoid"}
	EyesValues       = []string{"round", "oval", "focused", "visor"}
	EyeCountValues   = []string{"1", "2", "4", "6"}         // round eyes only: one centered eye, or 1-3 stacked pairs
	EyeSizeValues    = []string{"1", "2", "3", "4"}         // round eyes only, capped by the eye count (4 needs a single eye)
	EyeStyleValues   = []string{"glower", "bright", "dead"} // round eyes only
	MouthValues      = []string{"grille", "slot", "line", "jaw"}
	ExpressionValues = []string{"flat", "smile", "frown"}
	AntennaValues    = []string{"none", "ball", "double", "bolt"}
	EarsValues       = []string{"none", "bolts", "dials"}
)

// Spec describes one robot. A zero-valued field is unset and gets randomized
// by Resolve; a resolved Spec has every field set.
type Spec struct {
	Head, Eyes, EyeCount, EyeSize, EyeStyle, Mouth, Expression, Antenna, Ears string
	Face, Shoulders                                                           *int
	Tall, Eyelashes, Rivets, Panels, Blush                                    *bool
	// Colors are "#rrggbb"; Background may also be "none" (transparent).
	Body, Accent, Glow, Background string
}

// field is a name-addressable view of one Spec field, so parsing,
// validation, encoding and randomization can loop over facets.
type field struct {
	name     string
	kind     Kind
	values   []string // KindEnum only
	min, max int      // KindRange only
	str      *string  // KindEnum and KindColor
	flag     **bool   // KindBool
	num      **int    // KindRange
	// dependsOn names the facet values this facet only has an effect with.
	dependsOn *Dependency
}

func (s *Spec) fields() []field {
	return []field{
		{name: "head", kind: KindEnum, values: HeadValues, str: &s.Head},
		{name: "tall", kind: KindBool, flag: &s.Tall},
		{name: "eyes", kind: KindEnum, values: EyesValues, str: &s.Eyes},
		{name: "eyecount", kind: KindEnum, values: EyeCountValues, str: &s.EyeCount,
			dependsOn: roundFamily},
		{name: "eyesize", kind: KindEnum, values: EyeSizeValues, str: &s.EyeSize,
			dependsOn: roundFamily},
		{name: "eyestyle", kind: KindEnum, values: EyeStyleValues, str: &s.EyeStyle,
			dependsOn: roundFamily},
		{name: "eyelashes", kind: KindBool, flag: &s.Eyelashes, dependsOn: roundFamily},
		{name: "mouth", kind: KindEnum, values: MouthValues, str: &s.Mouth},
		{name: "expression", kind: KindEnum, values: ExpressionValues, str: &s.Expression},
		{name: "face", kind: KindRange, min: FaceMin, max: FaceMax, num: &s.Face},
		{name: "antenna", kind: KindEnum, values: AntennaValues, str: &s.Antenna},
		{name: "ears", kind: KindEnum, values: EarsValues, str: &s.Ears},
		{name: "shoulders", kind: KindRange, min: ShouldersMin, max: ShouldersMax, num: &s.Shoulders},
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
	switch f.kind {
	case KindBool:
		return *f.flag != nil
	case KindRange:
		return *f.num != nil
	}
	return *f.str != ""
}

func (f field) value() string {
	switch f.kind {
	case KindBool:
		return strconv.FormatBool(**f.flag)
	case KindRange:
		return strconv.Itoa(**f.num)
	}
	return *f.str
}

// Facet describes one customizable facet, for building UIs.
type Facet struct {
	Name   string   `json:"name"`
	Kind   Kind     `json:"kind"`
	Values []string `json:"values,omitempty"`
	Min    *int     `json:"min,omitempty"` // KindRange only
	Max    *int     `json:"max,omitempty"` // KindRange only
	// DependsOn, if set, is the facet values this facet needs to have any
	// effect; it is still resolved (randomly, if unset) either way.
	DependsOn *Dependency `json:"dependsOn,omitempty"`
}

// Dependency is the set of values of another facet a facet depends on.
type Dependency struct {
	Facet  string   `json:"facet"`
	Values []string `json:"values"`
}

// roundFamily is what the round-eye facets (count, size, style) depend on:
// oval and focused eyes are reshaped round eyes.
var roundFamily = &Dependency{Facet: "eyes", Values: []string{"round", "oval", "focused"}}

// Facets lists every facet in display order.
func Facets() []Facet {
	var out []Facet
	for _, f := range new(Spec).fields() {
		facet := Facet{Name: f.name, Kind: f.kind, Values: f.values, DependsOn: f.dependsOn}
		if f.kind == KindRange {
			facet.Min, facet.Max = &f.min, &f.max
		}
		out = append(out, facet)
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
			case f.kind == KindRange:
				n, err := strconv.Atoi(v)
				if err != nil {
					problems = append(problems, fmt.Sprintf("%s: %q is not an integer", key, v))
					continue
				}
				*f.num = &n
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
		if f.kind == KindRange {
			if n := **f.num; n < f.min || n > f.max {
				problems = append(problems, fmt.Sprintf("%s: %d is outside %d to %d", f.name, n, f.min, f.max))
			}
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
