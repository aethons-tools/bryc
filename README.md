# Bored Robots Yacht Club (bryc)

Cartoon robot portraits, drawn entirely in code. Pin the parts you want;
everything you leave unset is random. The seed is always reported, so any
robot can be reproduced.

## CLI

```bash
go run ./cmd/bryc                                  # a random robot → bryc-<seed>.png
go run ./cmd/bryc --head=dome --palette=mint       # pin some parts
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
