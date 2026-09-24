# at-bryc (Bored Robots Yacht Club) — Design

Date: 2026-09-24

## Goal

A small Go program that generates cartoonish robot portraits (head and
shoulders, front-facing) from a set of options, drawn entirely in code —
no art assets. A lightweight "character design screen", not a full editor.

Two front ends share one core package:

- **CLI** (`bryc`): flags in, PNG file out.
- **Web** (`bryc-web`): a local server with a single HTML page of controls and
  a live preview.

## Non-goals

- No asset/sprite layers, no user-supplied art.
- No accounts, persistence, or gallery. The URL/seed *is* the save format.
- No JS framework or front-end build step.
- No SVG output (possible later; not in v1).

## Module layout

Module path: `github.com/aethons-tools/bryc`.
Location: `~/local-repos/at-bryc`.

```
at-bryc/
├── go.mod
├── robot/                 # core; knows nothing about CLI or HTTP
│   ├── spec.go            # Spec, option enums, Validate, query/JSON encode/decode
│   ├── resolve.go         # Resolve(partial Spec, seed) → complete Spec
│   ├── palette.go         # named palettes
│   ├── render.go          # Render(complete Spec, size) → image.Image
│   └── parts/             # one file per part (head.go, eyes.go, mouth.go, ...)
├── cmd/bryc/              # CLI
└── cmd/bryc-web/          # web server
    └── static/index.html  # embedded via go:embed
```

Single external dependency: `github.com/fogleman/gg` (anti-aliased 2D drawing,
PNG encoding).

## Facets

Each facet is either **pinned** (set by the user) or **unset** (randomized).

| Facet        | Values                                              |
|--------------|-----------------------------------------------------|
| `head`       | `square`, `rounded`, `dome`, `trapezoid`            |
| `eyes`       | `round`, `visor`, `cyclops`, `led`                  |
| `mouth`      | `grille`, `speaker`, `smile`, `zigzag`              |
| `antenna`    | `none`, `ball`, `double`, `bolt`                    |
| `ears`       | `none`, `bolts`, `dials`                            |
| `rivets`     | `true`, `false`                                     |
| `panels`     | `true`, `false`                                     |
| `blush`      | `true`, `false`                                     |
| `body`       | hex color `#rrggbb`                                 |
| `accent`     | hex color `#rrggbb`                                 |
| `glow`       | hex color `#rrggbb` (eye glow)                      |
| `background` | hex color `#rrggbb` or `none` (transparent)         |

`palette` is a convenience input, not a facet: a named palette (e.g. `rusty`,
`mint`, `chrome`, plus a handful more) that sets `body`, `accent` and `glow`
together. Explicit color facets override the palette. Specifying an unknown
palette is a validation error.

Output settings are not facets and are never randomized:

- `size`: output edge length in pixels, square. Default `512`, allowed
  `64`–`2048`.

## Randomization and seeds

**Rule: every unset facet is random.** An empty spec produces a different
robot on each invocation.

- `seed` is an optional `uint64`. If absent, a fresh one is drawn from a
  non-deterministic source (`math/rand/v2` default source).
- The seed actually used is **always reported**: the CLI prints it to stderr
  (`seed: 1234...`), the web server returns it in an `X-Bryc-Seed` response
  header, and the web page displays it.
- Given the same seed and the same pinned facets, output is identical
  (deterministic across runs and platforms for the same program version).
- **Per-facet independence:** each facet draws from its own PRNG stream
  derived from `(seed, facet name)` (e.g. PCG seeded with `seed` and a hash of
  the facet name). Pinning one facet therefore does not change the random
  choices for any other facet.
- Random colors are picked from a curated set of palette-friendly colors
  (random hue with bounded saturation/lightness), not uniformly from all of
  RGB, so random robots look good.
- Random `background` is always a color (never `none`); `none` only when
  pinned.

`Resolve(partial Spec, seed uint64) Spec` is the single function that
implements this: it fills every unset facet and returns a complete Spec.
Unset is represented by the zero value (empty string / nil pointer for bools).

## Rendering

- All drawing happens on a virtual 1000×1000 coordinate space; `Render` scales
  to the requested `size`. Part code never deals in output pixels.
- Fixed layout: head box centered, with headroom above for the antenna and
  margins left/right for ears; a simple neck + shoulders at the bottom edge.
  Eyes and mouth are positioned relative to the head box, so every
  combination aligns.
- Draw order: background → shoulders/neck → ears → head → panels → rivets →
  eyes (with glow halo) → mouth → blush → antenna.
- Consistent cartoon style: thick dark outline on all shapes, flat body fill,
  accent color on antenna tips / dials / blush / shoulder trim, soft radial
  glow behind eyes.
- Each part is a function `func(dc *gg.Context, s Spec, box Layout)` registered
  in a map keyed by facet value; adding a variant = one function + one map
  entry.
- `Render` never fails for a complete, validated Spec.

## CLI (`bryc`)

```
bryc [--head=dome] [--eyes=visor] ... [--palette=mint] [--seed=N]
     [--size=512] [-o bot.png]
```

- Flags map 1:1 to facets; unset flags are random.
- `-o` defaults to `bryc-<seed>.png` in the current directory.
- Prints the seed used to stderr. Also prints the fully resolved spec as a
  query string (e.g. `head=dome&eyes=visor&...`) so a robot can be recreated
  exactly with pinned facets.
- Invalid input: prints the validation error to stderr, exits 1.

## Web (`bryc-web`)

```
bryc-web [--addr=localhost:8080]
```

- `GET /` — the single embedded HTML page.
- `GET /robot.png?<facets>&seed=&size=` — renders a PNG. Sets `X-Bryc-Seed`
  and `X-Bryc-Spec` (resolved spec as query string) headers. `Cache-Control:
  no-store` when no seed was given.
- `GET /options.json` — lists all facets and their allowed values and palettes,
  so the page builds its controls from the server's source of truth.
- Invalid input: HTTP 400 with the validation message as plain text.

Page behavior:

- One control per facet: dropdowns for enums, checkbox-style tri-state
  (random / on / off) for booleans, color picker + "random" toggle for colors.
  Every control has a "random" state, which is the default.
- Changing a control updates the `<img>` src; no page reload.
- **Randomize** button: clears the seed (new robot, pinned facets kept).
- **Lock seed** toggle: keep the current seed while tweaking facets.
- Shows the current seed; the page URL mirrors the current settings so it can
  be shared/bookmarked.
- **Download** link for the current PNG.
- On a 400, shows the error text in place of the image.

## Errors

`Spec.Validate()` collects all problems (unknown enum value, malformed hex
color, unknown palette, size out of range) into a single error listing each
problem and the allowed values. Validation runs before `Resolve`; after
resolve the spec is complete and rendering cannot fail.

## Testing

- `robot` unit tests: validation (good/bad inputs, aggregated messages),
  query-string round-trip, `Resolve` determinism (same seed + pins → same
  spec), per-facet independence (pinning facet A leaves facet B's random value
  unchanged), unseeded resolve differs across calls, palette/override
  precedence.
- Render smoke test: every value of every enum facet renders without panic at
  a small size.
- Golden images: a few fixed-seed robots compared pixel-for-pixel against
  `testdata/*.png`; `go test ./robot -update` regenerates.
- Web: `httptest` — `/robot.png` valid → 200 + `image/png` + seed header;
  invalid → 400; `/options.json` → 200 and parses.
- CLI: test the flag-to-Spec parsing function directly (no subprocess).
