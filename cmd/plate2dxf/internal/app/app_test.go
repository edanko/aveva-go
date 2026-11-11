package app

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	gen "github.com/edanko/aveva-go"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

// plateDocument builds a single-part plate document. The marking sides make the
// drawer's auto-mirror heuristic testable: OS is deliberately longer than TS.
func plateDocument(osLonger bool) string {
	tsLen, osLen := 100.0, 10.0
	if osLonger {
		tsLen, osLen = 10.0, 100.0
	}

	lines := []string{
		"GENERAL_DATA",
		"TYPE_OF_GENERIC_FILE=PLATE_PART",
		"RAW_THICKNESS=10.00",
		"QUALITY=A40",
		"NO_OF_PARTS=1",
		"END_OF_GENERAL_DATA",
		"PART_DATA",
		"NAME=P1",
		"POSNO=1",
		"SHIP_NO=123",
		"BLOCK_NO=B1",
		"PART_AREA=1000.00",
		"MIRRORED=0",
		"END_OF_PART_DATA",
		"PART_INFORMATION",
		"PART_NAME=P1",
		"TYPE_OF_WORK=BURNING",
		"END_OF_PART_INFORMATION",
	}

	marking := func(side string, length float64) []string {
		return []string{
			"MARKING_DATA",
			"MARKING_SIDE=" + side,
			"MARKING_TYPE=PANEL",
			"START_OF_CONTOUR",
			"NO_OF_SEG=1",
			"START_U=0", "START_V=0",
			"AMP_U=0", "AMP_V=0", "AMP=0", "RADIUS=0", "SWEEP=0",
			"ORIGIN_U=0", "ORIGIN_V=0",
			"U=" + formatFloat(length), "V=0",
			"END_OF_CONTOUR",
			"END_OF_MARKING_DATA",
		}
	}

	lines = append(lines, marking("TS", tsLen)...)
	lines = append(lines, marking("OS", osLen)...)

	return strings.Join(lines, "\r\n")
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func TestProcessWritesDXF(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"P1.gen": &fstest.MapFile{Data: []byte(plateDocument(false))},
	}
	out := t.TempDir()

	a := New(discardLogger(), Options{Mirror: MirrorAuto})
	require.NoError(t, a.Process(context.Background(), source, out))

	dxfs := findFiles(t, out, ".dxf")
	require.Len(t, dxfs, 1, "one part must produce one DXF")

	content, err := os.ReadFile(dxfs[0])
	require.NoError(t, err)

	body := string(content)
	require.Contains(t, body, "SECTION")
	require.Contains(t, body, "ENTITIES")
	require.Contains(t, body, "LWPOLYLINE", "the cut/marking contours must be drawn")
}

// A mixed directory holds listed-profile documents too; they are not this tool's
// input and must be skipped rather than failing the run.
func TestProcessSkipsOtherDocumentFamilies(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"P1.gen":        &fstest.MapFile{Data: []byte(plateDocument(false))},
		"profile.gen":   &fstest.MapFile{Data: []byte("TYPE_OF_GENERIC_FILE=LISTED_PROFILE\r\nVERSION=1.0\r\nUSAGE=LIST\r\nCOMMON_DATA\r\nEND_OF_COMMON_DATA\r\n")},
		"unrelated.txt": &fstest.MapFile{Data: []byte("not a generic file")},
	}
	out := t.TempDir()

	a := New(discardLogger(), Options{Mirror: MirrorAuto})
	require.NoError(t, a.Process(context.Background(), source, out),
		"a non plate-part document must not fail the run")

	require.Len(t, findFiles(t, out, ".dxf"), 1)
}

func TestProcessMirrorModes(t *testing.T) {
	t.Parallel()

	// OS marking longer than TS, so the auto heuristic chooses to mirror.
	doc := plateDocument(true)

	render := func(mode MirrorMode) string {
		t.Helper()

		source := fstest.MapFS{"P1.gen": &fstest.MapFile{Data: []byte(doc)}}
		out := t.TempDir()

		a := New(discardLogger(), Options{Mirror: mode})
		require.NoError(t, a.Process(context.Background(), source, out))

		dxfs := findFiles(t, out, ".dxf")
		require.Len(t, dxfs, 1)

		content, err := os.ReadFile(dxfs[0])
		require.NoError(t, err)

		return string(content)
	}

	auto := render(MirrorAuto)
	on := render(MirrorOn)
	off := render(MirrorOff)

	// `on` and `auto` both mirror here, so they must agree; `off` must not.
	require.Equal(t, on, auto,
		"with OS longer than TS the auto heuristic mirrors, matching forced mirroring")
	require.NotEqual(t, auto, off,
		"disabling mirroring must change the drawing")
}

// When TS is longer than OS the heuristic declines to mirror, so auto and off
// agree while forced mirroring differs.
func TestProcessMirrorModesNoHeuristicMirror(t *testing.T) {
	t.Parallel()

	doc := plateDocument(false)

	render := func(mode MirrorMode) string {
		t.Helper()

		source := fstest.MapFS{"P1.gen": &fstest.MapFile{Data: []byte(doc)}}
		out := t.TempDir()

		a := New(discardLogger(), Options{Mirror: mode})
		require.NoError(t, a.Process(context.Background(), source, out))

		dxfs := findFiles(t, out, ".dxf")
		require.Len(t, dxfs, 1)

		content, err := os.ReadFile(dxfs[0])
		require.NoError(t, err)

		return string(content)
	}

	require.Equal(t, render(MirrorAuto), render(MirrorOff))
	require.NotEqual(t, render(MirrorAuto), render(MirrorOn))
}

// A parse failure must propagate rather than leave a partial run looking healthy.
func TestProcessPropagatesParseErrors(t *testing.T) {
	t.Parallel()

	// GENERAL_DATA with an unparseable numeric value.
	broken := "GENERAL_DATA\r\nRAW_THICKNESS=not-a-number\r\nEND_OF_GENERAL_DATA\r\n"

	source := fstest.MapFS{"broken.gen": &fstest.MapFile{Data: []byte(broken)}}
	out := t.TempDir()

	a := New(discardLogger(), Options{Mirror: MirrorAuto})
	require.Error(t, a.Process(context.Background(), source, out))
}

func findFiles(t *testing.T, root, ext string) []string {
	t.Helper()

	var found []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ext) {
			found = append(found, path)
		}

		return nil
	})
	require.NoError(t, err)

	return found
}

// BenchmarkProcess measures a full parse-and-render pass over a in-memory
// document, which is what the tool does per file.
func BenchmarkProcess(b *testing.B) {
	doc := []byte(plateDocument(false))
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	for _, mode := range []MirrorMode{MirrorAuto, MirrorOn, MirrorOff} {
		b.Run(string(mode), func(b *testing.B) {
			a := New(logger, Options{Mirror: mode})
			b.ReportAllocs()

			for b.Loop() {
				source := fstest.MapFS{"P1.gen": &fstest.MapFile{Data: doc}}
				if err := a.Process(context.Background(), source, b.TempDir()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

var _ = gen.ErrNotPlatePart
