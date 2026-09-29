// Command bryc renders a Bored Robots Yacht Club robot to a PNG or .glb file.
// Every option left unset is random; the seed used is printed so a robot
// can be reproduced.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aethons-tools/bryc/robot"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

type options struct {
	req    robot.Request
	out    string // "" means bryc-<seed>.<format>
	format string // "png" or "glb"
}

// flagError marks errors the FlagSet has already printed along with usage.
type flagError struct{ error }

func parseArgs(args []string, stderr io.Writer) (options, error) {
	fs := flag.NewFlagSet("bryc", flag.ContinueOnError)
	fs.SetOutput(stderr)
	for _, f := range robot.Facets() {
		var usage string
		switch f.Kind {
		case robot.KindEnum:
			usage = strings.Join(f.Values, " | ")
		case robot.KindBool:
			usage = "true | false"
		case robot.KindRange:
			usage = fmt.Sprintf("integer %d to %d", *f.Min, *f.Max)
		case robot.KindColor:
			usage = "#rrggbb"
			if f.Name == "background" {
				usage += " | none"
			}
		}
		fs.String(f.Name, "", usage+" (random if unset)")
	}
	fs.String("palette", "", strings.Join(robot.PaletteNames(), " | "))
	fs.String("seed", "", "random seed (random if unset)")
	fs.String("size", strconv.Itoa(robot.DefaultSize), "output width and height in pixels")
	format := fs.String("format", "png", "png | glb")
	out := fs.String("o", "", "output file (default bryc-<seed>.<format>)")

	if err := fs.Parse(args); err != nil {
		return options{}, flagError{err}
	}
	if fs.NArg() > 0 {
		return options{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	if *format != "png" && *format != "glb" {
		return options{}, fmt.Errorf("unknown format %q (want png or glb)", *format)
	}
	q := url.Values{}
	fs.Visit(func(f *flag.Flag) {
		if f.Name != "o" && f.Name != "format" {
			q.Set(f.Name, f.Value.String())
		}
	})
	req, err := robot.ParseQuery(q)
	if err != nil {
		return options{}, err
	}
	return options{req: req, out: *out, format: *format}, nil
}

func run(args []string, stderr io.Writer) int {
	opts, err := parseArgs(args, stderr)
	var fe flagError
	switch {
	case errors.As(err, &fe) && fe.error == flag.ErrHelp:
		return 0
	case errors.As(err, &fe):
		return 2
	case err != nil:
		fmt.Fprintln(stderr, err)
		return 1
	}

	spec, seed := opts.req.Resolve()
	out := opts.out
	if out == "" {
		out = fmt.Sprintf("bryc-%d.%s", seed, opts.format)
	}
	var werr error
	if opts.format == "glb" {
		werr = writeGLB(out, spec)
	} else {
		werr = writePNG(out, robot.Render(spec, opts.req.Size))
	}
	if werr != nil {
		fmt.Fprintln(stderr, werr)
		return 1
	}
	fmt.Fprintf(stderr, "seed: %d\nspec: %s\nwrote %s\n", seed, spec.Query().Encode(), out)
	return 0
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeGLB(path string, spec robot.Spec) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := robot.WriteGLB(f, spec); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
