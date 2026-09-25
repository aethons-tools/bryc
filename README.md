# Bored Robots Yacht Club (bryc)

Cartoon robot portraits, drawn entirely in code. Pin the parts you want;
everything you leave unset is random. The seed is always reported, so any
robot can be reproduced.

## CLI

```bash
go run ./cmd/bryc                                  # a random robot → bryc-<seed>.png
go run ./cmd/bryc --head=dome --palette=mint       # pin some parts
go run ./cmd/bryc --head=inverted-dome --tall=true  # heads: square rounded dome inverted-dome trapezoid inverted-trapezoid
go run ./cmd/bryc --shoulders=-80                  # spiky shoulders (-100..100, 0 = square)
go run ./cmd/bryc --mouth=jaw --expression=smile   # mouths: grille slot line jaw; flat smile frown
go run ./cmd/bryc --face=2                          # face position -1 (low) .. 2 (high and compact)
go run ./cmd/bryc --eyes=round --eyecount=6        # 2, 4 or 6 round eyes
go run ./cmd/bryc --eyes=round --eyesize=1         # smaller round eyes (1..4, capped by eyecount)
go run ./cmd/bryc --eyecount=1 --eyesize=4 --eyestyle=glower  # the HAL look
go run ./cmd/bryc --eyes=focused --eyestyle=bright  # eyes: round oval focused visor
go run ./cmd/bryc --seed=42 --size=1024 -o bot.png # reproduce a robot
go run ./cmd/bryc -h                               # all options
```

## Web designer

```bash
go run ./cmd/bryc-web          # http://localhost:8080
```

`GET /robot.png?head=dome&seed=42&size=512` returns a PNG; the seed and the
fully resolved options come back in `X-Bryc-Seed` and `X-Bryc-Spec`.

## Development

```bash
go test ./...
go test ./robot -update   # regenerate golden images after an intended visual change
```
