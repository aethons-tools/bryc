# BRYC 3D mode (phase 1) — Design

Date: 2026-09-28
Builds on: `2026-09-24-bryc-design.md` (the 2D generator).

## Goal

Every robot the 2D generator can make can also be produced as a 3D model:

- a `.glb` (binary glTF 2.0) file from the server (`GET /robot.glb`) and the
  CLI (`bryc --format=glb`), usable in Blender, game engines and AR viewers;
- a spinnable 3D view in the designer page (Single view), with a
  **Download .glb** button.

Phase 1 is "2.5D": parts are built mostly from the same 2D outlines, pushed
out into rounded slabs, with simple solids (ellipsoids, cylinders, tubes) for
eyes, rivets, antennas and lashes. The look is soft, toy-like shading (vinyl
and metal materials, self-lit glow).

## Non-goals (this phase)

- Cartoon outline shells ("inverted hull"). The design must leave room to add
  them later as an option; nothing is built now.
- True rounded volumes (a dome that is really a dome; eyes seated on curved
  surfaces).
- 3D in grid mode, animation, printability (wall thickness, a single manifold
  solid).
- Painted 2D effects in 3D: glints, halos and hot spots are 2D-only; real
  lighting and emissive materials replace them. The background color is not
  part of the model.

## Architecture

Three units, each testable on its own:

```
robot/mesh/   generic 3D geometry; knows nothing about robots
robot/glb/    generic .glb writer; knows nothing about robots
robot/        existing package; gains model.go + model_*.go
```

### `robot/mesh`

- `Vec2{X, Y}`, `Vec3{X, Y, Z}`; `Mesh{Positions, Normals []Vec3; Indices []uint32}`.
- Mesh operations: `Append`, `Translate`, `Scale`, `RotateZ`, `Bounds`.
- `Triangulate(poly []Vec2) ([]int, error)`: ear clipping for simple polygons
  (either winding; output triangles are counter-clockwise).
- `Extrude(poly []Vec2, depth, bevel float64) (Mesh, error)`: a closed slab
  centred on z=0 whose side edges are rounded with a quarter-circle of radius
  `bevel` (a few rings; inset by offsetting the outline along vertex
  bisectors, miter length clamped). Analytic smooth normals on the rounded
  edges and walls, flat normals on the caps.
- `Ellipsoid(rx, ry, rz)`, `Cylinder(r, length)` (along Z, capped),
  `Tube(path []Vec3, radius func(t) float64)` (capped; radius may taper to
  ~0 at the end).
- Test helpers exported for other packages' tests: `(Mesh).Closed() bool`
  (every edge shared by exactly two triangles, in opposite directions) and
  `(Mesh).Volume() float64` (signed; positive means outward-facing).

### `robot/glb`

- `Scene{Nodes []Node}`; `Node{Name string; Mesh mesh.Mesh; Material Material}`;
  `Material{Name string; Color [4]float64 (linear RGBA); Metallic, Roughness
  float64; Emissive [3]float64 (linear); Blend bool}`.
- `Encode(Scene) ([]byte, error)`: glTF 2.0 binary. One buffer; per node
  float32 POSITION (with min/max) and NORMAL accessors and uint32 indices;
  4-byte aligned buffer views and chunks; materials deduplicated in first-use
  order; all nodes children of a root node `robot`; `asset.generator`
  "bryc". Output is deterministic (no maps iterated during encoding).

### Model (in `robot`)

- `Model(s Spec) glb.Scene` for a resolved spec. One node per part, named
  (`head`, `neck`, `shoulders`, `chest-light`, `eye-0…`, `eye-ring-0…`,
  `eye-core-0…`, `visor`, `visor-bar`, `mouth`, `mouth-bar-0…`, `jaw`,
  `jaw-bolt-0/1`, `rivet-0…`, `seam-0/1`, `blush-0/1`, `lash-0…`,
  `ear-0/1` (+ `ear-cap-*`, `ear-tick-*`), `antenna-*`).
- `WriteGLB(w io.Writer, s Spec) error` convenience.
- Building a model never fails for a resolved, validated spec (geometry
  errors are programming errors and panic, as `Render` would).

**Coordinates.** Canvas units (1000 = the image) → metres: `X = (x-500)/1000`,
`Y = (1000-y)/1000`, `Z = z/1000`. So the model is ~1 m tall, stands on Y=0
(good for AR), faces +Z.

**Shared placement.** 2D and 3D take every position and size from the same
code. Existing shared functions are reused (`headLayout`, `faceBox`,
`facePos`, `roundEyeCenters`, `roundEyeR`, `eyeTilt`, `leftEdge`, `headTop`,
`rivetCenters`, `panelSeams`, `lashFor`, `curve`, `blushY`, `bowRatio`,
`shoulder*` constants). Numbers now written inline in 2D part code (ear
positions and sizes, antenna geometry, visor size, grille/slot sizes, jaw
key heights, chest light, neck) move into small shared functions or
constants. **The 2D output must not change**: all saved golden PNGs must
match byte-for-byte after the refactor.

New shared outline samplers (2D canvas coordinates, used by 3D and tested
against the 2D edge functions):

- `headOutline(s, b) []mesh.Vec2`: the head's full outline, matching
  `headPaths` (rounded corners, dome cap, bowed edges, tall/narrow boxes).
- `shoulderOutline(shoulders) []mesh.Vec2`: matching `shoulderPath`, cut at
  the canvas bottom (y=1000).
- `jawOutline(s, b) []mesh.Vec2`: the U plate: the head outline scaled by
  `jawOverhang` about the head centre, kept below the jaw's top, with the
  face cut out down to the mouth curve (same geometry as `drawJaw`).

### Parts in 3D

Depths in canvas units; "front" is +Z; the head's front face is at z=+150.

| part | form | depth / placement |
|---|---|---|
| head | `Extrude(headOutline)` | 300 thick, bevel 12 |
| neck | rounded box (extruded rect) | 140 thick, bevel 8 |
| shoulders | `Extrude(shoulderOutline)` | 360 thick, bevel 12 |
| chest light | ellipsoid, accent (rx = ry = 30, rz 15) | centred on the shoulders' front face (z +180), so it bulges 15 |
| bright eye | ellipsoid, glow (emissive) | rx r, ry r·aspect, rz 0.5r, centre on face (z 150); oval/tilt as 2D |
| glower eye | metal cylinder ring r, length 0.4r; black lens ellipsoid (r−ring, bulge 0.3r); glow core sphere (core radius) just under the lens surface | on face |
| dead eye | ellipsoid, screen | as bright |
| visor | `Extrude(stadium)`, screen | 40 thick, front at z 150+20; glow bar `Extrude` 10 thick just proud of it |
| grille | `Extrude(bent rect)`, metal, 20 thick; bars as thin ink slabs on its front | on face |
| slot | `Extrude(bent rect)`, screen, 6 thick | on face |
| line | `Tube(curve)`, ink, radius 7 | on face |
| jaw | `Extrude(jawOutline)`, body shade `jawShade`, 330 thick, bevel 10; hinge bolts metal spheres on arms | wraps lower face front and back |
| rivets | metal ellipsoid r 12, bulge 6 | on face |
| panel seams | `Tube` along seam curve, ink, radius 4 | on face |
| blush | `Extrude(ellipse)`, accent, 60% alpha (Blend), 2 thick | just proud of face |
| lashes | `Tube(lash curve)`, ink, radius = `halfWidth(t)` | root tucked at the eye edge, on face |
| ears: bolts | rounded box (`Extrude`) accent, from the head side | 240 deep (z ±120) |
| ears: dials | accent `Cylinder` along X out of the head side (r 62, 40 long), body-coloured cap (r 30), ink tick | at head side |
| antenna | stems metal `Tube`s along the 2D segments; balls accent ellipsoids; base rounded box; bolt accent `Tube` along the zigzag (radius 9) | on the head top, centred in z |

### Materials (glTF metallic-roughness; colours converted sRGB → linear)

| material | colour | metallic | roughness | notes |
|---|---|---|---|---|
| body | body | 0 | 0.45 | jaw uses body × `jawShade`, neck body × 0.7 |
| accent | accent | 0 | 0.45 | |
| metal | `metal` | 0.8 | 0.3 | |
| screen | `screen` | 0 | 0.15 | |
| glow | glow | 0 | 0.4 | emissive = glow |
| ink | `ink` | 0 | 0.5 | |
| blush | accent, alpha 0.6 | 0 | 0.45 | Blend |

## Server, CLI, page

- `GET /robot.glb`: same query parameters as `/robot.png` (including `cell`;
  `size` is ignored), same `X-Bryc-Seed` / `X-Bryc-Spec` headers and
  `Cache-Control` rule, `Content-Type: model/gltf-binary`; 400 with the same
  messages on invalid input.
- CLI: `--format=png|glb` (default png). Default output `bryc-<seed>.<ext>`.
  Unknown format → error, exit 1.
- Page (Single view): a **2D / 3D** toggle. 3D view: three.js (pinned
  version) from jsDelivr via an import map (no build step); GLTFLoader,
  OrbitControls (drag to spin, scroll/pinch to zoom), auto-rotate until the
  user interacts, `RoomEnvironment` lighting; camera framed from the model's
  bounds; backdrop = robot background colour (neutral grey for `none`).
  Changing a control swaps in the new model and keeps the camera angle.
  **Download .glb** in 3D view. `?view=3d` in the URL. Grid mode stays 2D;
  choosing Grid switches to 2D. If WebGL is unavailable the 3D toggle is
  disabled with a short explanation; if loading fails, a message replaces
  the viewer.

## Testing

- `mesh`: triangulated area equals polygon area (convex, concave, either
  winding); `Extrude`, `Ellipsoid`, `Cylinder`, `Tube` are closed with
  positive volume; extrusions stay within the outline's bounds and depth;
  bevel ≤ 0 or degenerate input returns an error.
- `glb`: encode, then parse back in the test: header magic/version/length,
  JSON chunk then BIN chunk, 4-byte alignment, every accessor's buffer view
  and byte range inside the buffer, POSITION min/max correct, index count ÷
  3, material indices valid, deterministic bytes.
- Model: every value of every facet builds; expected node names present for
  representative specs; all meshes closed; same spec → identical bytes;
  head outline x-extent matches `leftEdge` at several heights for every head
  variant; face parts lie in front of z=150−(their embed); whole model
  within X ±0.5 m, Y 0..1 m; golden `.glb` files for a few specs
  (`-update`); budget: model builds in < 100 ms and encodes to < 1 MB.
- 2D unchanged: existing golden PNGs match exactly.
- Server/CLI: `/robot.glb` status, content type, headers, determinism, 400;
  `--format=glb` writes a file that parses; unknown format errors.
- Browser (manual, by the implementer): 3D loads, toggle, spin/zoom,
  control change keeps angle, download link, `?view=3d`, grid switches to 2D,
  phone width.
