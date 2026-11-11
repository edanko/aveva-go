package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	gen "github.com/edanko/aveva-go"
	"github.com/edanko/aveva-go/cmd/plate2dxf/internal/draw"
)

// MirrorMode selects how part orientation is decided.
type MirrorMode string

const (
	// MirrorAuto keeps the historical heuristic: a part is flipped when its OS
	// marking length exceeds the TS marking length.
	MirrorAuto MirrorMode = "auto"
	// MirrorOn forces every part to be mirrored about its U axis.
	MirrorOn MirrorMode = "on"
	// MirrorOff disables mirroring entirely, including the automatic heuristic.
	MirrorOff MirrorMode = "off"
)

// Options configures a run.
type Options struct {
	Mirror MirrorMode
}

type App struct {
	logger *slog.Logger
	opts   Options
}

func New(logger *slog.Logger, opts Options) *App {
	return &App{
		logger: logger,
		opts:   opts,
	}
}

func (a *App) Process(ctx context.Context, source fs.FS, outputDir string) error {
	g, ctx := errgroup.WithContext(ctx)

	files := make(chan string, 100)

	g.Go(func() error {
		defer close(files)

		return fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			if filepath.Ext(d.Name()) != ".gen" {
				return nil
			}

			select {
			case files <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	c := make(chan *gen.Part)
	for file := range files {
		g.Go(func() error {
			f, err := source.Open(file)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil {
					a.logger.Error("close failed", "path", file, "error", cerr)
				}
			}()

			parts, err := gen.ReadParts(f)
			if err != nil {
				if errors.Is(err, gen.ErrNotPlatePart) {
					// A mixed directory may hold listed-profile or other
					// documents; they are not this tool's input.
					a.logger.Debug("skipping non plate-part document", "path", file)
					return nil
				}
				return err
			}

			switch a.opts.Mirror {
			case MirrorOn:
				// Mirror() rewrites contours, marking sides and every absolute
				// U coordinate; the drawer then exports the mirrored part.
				for _, p := range parts {
					p.Mirror()
				}
			case MirrorOff:
				// Orientation is explicit; the drawer must not override it.
			}

			a.logger.Info("successfully parsed", "path", file)
			for _, part := range parts {
				select {
				case c <- part:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	// parseErr is written before close(c) and read after the range over c
	// completes, so the channel close orders the two accesses.
	var parseErr error

	go func() {
		parseErr = g.Wait()
		close(c)
	}()

	parts := make([]*gen.Part, 0, 100)
	for p := range c {
		parts = append(parts, p)
	}

	// A malformed document must be reported, not turned into a panic.
	if parseErr != nil {
		return parseErr
	}

	parts = lo.UniqBy(parts, func(p *gen.Part) string {
		return p.ShipNo + p.BlockNo + p.PosNo
	})

	g2, ctx := errgroup.WithContext(ctx)

	feeder := make(chan *gen.Part, 18)

	g2.Go(func() error {
		defer close(feeder)

		for _, p := range parts {
			feeder <- p
		}
		return nil
	})

	for elem := range feeder {
		g2.Go(func() error {
			dir := time.Now().Format("06.01.02") + " " + elem.BlockNo + " dxf"
			subdir := strconv.FormatFloat(elem.Thickness, 'g', -1, 64)
			dir = path.Join(outputDir, dir, subdir)
			file := fmt.Sprintf("%s-%s-%s.dxf", elem.ShipNo, elem.BlockNo, elem.PosNo)
			out := path.Join(dir, file)

			// 0750 rather than fs.ModePerm: output directories have no reason to
			// be group- or world-writable.
			err := os.MkdirAll(dir, 0o750)
			if err != nil {
				return err
			}

			// out is built from this tool's own output directory, not from
			// untrusted input, so G304 does not apply. #nosec is understood by
			// both standalone gosec and gosec inside golangci-lint.
			f, err := os.Create(out) // #nosec G304 -- path is built locally

			if err != nil {
				return err
			}
			defer func() {
				if cerr := f.Close(); cerr != nil {
					a.logger.Error("close failed", "path", out, "error", cerr)
				}
			}()

			drawer, err := draw.NewDrawer(f)
			if err != nil {
				return err
			}
			// Only 'auto' lets the drawer pick the orientation itself.
			drawer.SetAutoMirror(a.opts.Mirror == MirrorAuto)

			return drawer.PartToDXF(elem)
		})
	}

	if err := g2.Wait(); err != nil {
		return err
	}

	return nil
}
