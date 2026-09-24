package robot

import "sort"

// Palette is a named set of colors applied to any unset color facets.
type Palette struct {
	Body, Accent, Glow string
}

var palettes = map[string]Palette{
	"bubblegum": {Body: "#ffb3d9", Accent: "#8a4fff", Glow: "#7cf3ff"},
	"chrome":    {Body: "#c0c7d1", Accent: "#4a6fa5", Glow: "#ff3b3b"},
	"midnight":  {Body: "#3b4a6b", Accent: "#f2a541", Glow: "#9dff6b"},
	"mint":      {Body: "#98e2c6", Accent: "#ff8fa3", Glow: "#3ae0ff"},
	"rusty":     {Body: "#b5651d", Accent: "#6b8e23", Glow: "#ffd23f"},
	"sunny":     {Body: "#ffd166", Accent: "#ef476f", Glow: "#06d6a0"},
}

// PaletteNames returns the palette names in sorted order.
func PaletteNames() []string {
	names := make([]string, 0, len(palettes))
	for name := range palettes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
