# aveva-go

> Read and write AVEVA Marine generic files (`*.gen`) from Go.

[![Test](https://github.com/edanko/aveva-go/actions/workflows/test.yml/badge.svg)](https://github.com/edanko/aveva-go/actions/workflows/test.yml)
[![Lint](https://github.com/edanko/aveva-go/actions/workflows/lint.yml/badge.svg)](https://github.com/edanko/aveva-go/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/edanko/aveva-go.svg)](https://pkg.go.dev/github.com/edanko/aveva-go)
[![Go version](https://img.shields.io/github/go-mod/go-version/edanko/aveva-go)](go.mod)

A `.gen` file is line-oriented `KEY=VALUE` text that a shipyard's nesting software
emits for each plate or profile. There are two document families — `PLATE_PART`
and `LISTED_PROFILE` — with separate writers upstream and different grammars.
This package handles both, in both directions.

It is not a wrapper around a vendor SDK. The grammar was recovered from the
shipped writers plus a large corpus of production output, and the parser is
written to accept what the corpus actually contains rather than an idealised
version of it.

## Features

- **Both families.** `PLATE_PART` nesting and `LISTED_PROFILE` profiles.
- **Reader and writer.** Every modelled section round-trips: parse → write →
  parse is verified `reflect.DeepEqual` over 45,000+ real files.
- **Complete sections.** Cut paths, markings, labels, holes, bevelled edges,
  edges, auxiliary directives, idle and bump paths and profile end cuts.
- **Production quirks tolerated**, because they are real: documents with no
  `PART_DATA`, scopes with an empty `PART_NAME`, more scopes than placements,
  `-` sentinels, and empty numeric values.
- **No silent data loss.** Sections whose schema has no available sample are
  preserved verbatim rather than dropped or guessed.
- **Mirroring** as a first-class operation, including materialising the result.

## Install

```bash
go get github.com/edanko/aveva-go
```

Requires Go 1.27 or later.

## Quick start

The package name is `gen`:

```go
import gen "github.com/edanko/aveva-go"
```

### Reading a plate nesting

```go
f, err := os.Open("nest.gen")
if err != nil {
    return err
}
defer f.Close()

parts, err := gen.ReadParts(f)
if errors.Is(err, gen.ErrNotPlatePart) {
    // Not a nesting: a profile file or something else entirely.
    return nil
}
if err != nil {
    return err
}

for _, p := range parts {
    fmt.Printf("%s  %.2fmm  %q  %d cut paths\n",
        p.Name, p.Thickness, p.Quality, len(p.BurningData))
}
```

### Reading the whole document

`ReadDocument` also returns the blocks that belong to no single part — the
labelling pass a nesting emits at the end carries its own text, so it has no
owning part to hang off:

```go
doc, err := gen.ReadDocument(f)
if err != nil {
    return err
}

fmt.Println(doc.Type, doc.General.Thickness, doc.General.Quality)
fmt.Println(len(doc.Parts), "parts,", len(doc.LabelTexts), "unattributed labels")
```

### Reading profiles

```go
doc, err := gen.ReadProfileDocument(f)
if errors.Is(err, gen.ErrNotListedProfile) {
    return nil
}

for _, p := range doc.Profiles {
    fmt.Printf("%s  L=%.1f  left end cut %d\n", p.Name, p.MLength, p.LeftEnd.EndcutType)
}
```

`COMMON_DATA` sits on the document (`doc.Common`), because the profile's
cross-section and stock lengths are shared by every profile in the file.

### Writing

```go
var buf bytes.Buffer
if err := gen.WriteDocument(&buf, doc); err != nil {
    return err
}
```

`WriteProfiles` is the counterpart for profiles. Both accept optional
`WriteOptions`; the default matches what AVEVA emits, including CRLF line endings
and empty values.

### Mirroring

```go
for _, p := range doc.Parts {
    p.Mirror() // flips about the U axis: contours, marking sides, all U coordinates
}

var mirrored bytes.Buffer
if err := gen.WriteDocument(&mirrored, doc); err != nil {
    return err
}
```

## API

| Function | Purpose |
| --- | --- |
| `ReadParts(io.Reader) ([]*Part, error)` | Parts of a `PLATE_PART` document |
| `ReadDocument(io.Reader) (*Document, error)` | Parts plus document-level blocks |
| `DocumentType(io.Reader) (string, error)` | `TYPE_OF_GENERIC_FILE` without full parsing |
| `ReadProfiles(io.Reader) ([]*Profile, error)` | Profiles of a `LISTED_PROFILE` document |
| `ReadProfileDocument(io.Reader) (*ProfileDocument, error)` | Profiles plus header and `COMMON_DATA` |
| `WriteDocument(io.Writer, *Document, ...WriteOptions) error` | Emit a plate document |
| `WriteParts(io.Writer, []*Part, ...WriteOptions) error` | Emit parts without document blocks |
| `WriteProfiles(io.Writer, *ProfileDocument, ...WriteOptions) error` | Emit a profile document |
| `(*Part).Mirror()` | Flip a part about its U axis |

`ErrNotPlatePart` and `ErrNotListedProfile` report a wrong document family, so a
mixed directory can be skipped rather than treated as a corrupt file. A real
parse failure returns its own error, and the two are never conflated.

## Command line

`cmd/plate2dxf` exports a nesting to DXF, one file per part:

```bash
go install github.com/edanko/aveva-go/cmd/plate2dxf@latest

plate2dxf -i ./gens -o ./dxf
plate2dxf -i ./gens -o ./dxf -mirror=on
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-i` | `.` | Input directory, searched recursively |
| `-o` | `.` | Output directory |
| `-mirror` | `auto` | `auto` flips a part when its OS marking is longer than its TS one, `on` always flips, `off` never flips |

Output is grouped the way the downstream process expects:

```
dxf/
└── 26.11.11 32583 dxf/
    └── 45/
        ├── 272-32583-2.dxf
        └── 272-32583-3.dxf
```

The directory is `<YY.MM.DD> <block> dxf/<thickness>` and each file is
`<ship>-<block>-<position>.dxf`.

## The format, briefly

Enough to read a file without being surprised:

- Line-oriented `KEY=VALUE`, terminated by bare section markers such as
  `PART_DATA` … `END_OF_PART_DATA`. CRLF, no escaping, no quoting.
- **`PART_INFORMATION` is not 1:1 with `PART_DATA`.** A part can be *placed* N
  times and *opened* by M scopes; payload binds positionally and surplus scopes
  accumulate on the last placement. Keying parts by name loses geometry.
- **`-` means "this parameter does not apply"**, not zero. `ENDCUT_TYPE` remains
  authoritative for telling an absent radius from a real zero.
- **`ENDCUT_MASK` is a bit pattern** such as `00010110000`. It is a string here on
  purpose; leading zeros do not survive an integer.
- **Numbers vary by writer.** Empty values and `-` both occur, so parsing is
  tolerant rather than strict.

## Development

```bash
make help     # list every target
make ci       # exactly what CI runs: verify, tidy, build, vet, lint, gofmt, race tests
make check    # a faster local gate
```

Also available: `make test-cover`, `make bench`, `make fuzz`, `make vuln`.

Tests use the fixtures in `testdata/`, which are anonymised real output: the
structure and every format feature are preserved, the identifiers are not.

## Contributing

Issues and pull requests are welcome. Before opening a PR:

```bash
make ci     # must pass; it is the same gate CI runs
```

Two things that make a change easy to accept:

- **Add a fixture or a test that fails first.** Most format work is discovered
  from real output, so a test that reproduces the shape you saw is worth more than
  a description of it.
- **Say what is measured and what is inferred.** Field presence, widths and order
  in the writers come either from the shipped serialisers or from sampling the
  corpus; keep the distinction visible in comments and commit messages.

## Limitations

- **The writers are not byte-identical to AVEVA output.** The corpus contains
  several writer dialects that disagree on field presence and on empty-versus-zero,
  so no single canonical writer can match them all. Structural round-trip equality
  is the contract, and that holds exactly.
- **Some sections have no available sample** — `MARKING_LEFT_END`,
  `REF_LINE_MARKING`, `BENDING_TABLE`, `CURVE_DATA`, `PENETRATING_PROFILE`,
  `DRILLING`, `SLOTTED` and a few others. Their field names are known from the
  binaries, but nothing here can validate an interpretation, so they are captured
  verbatim as `RawSection` instead of being modelled on guesswork.
- **The `LISTED_PROFILE` writer is sample-derived**; the plate-side writer's order
  and widths come from the decompiled serialisers.
- A handful of sections are read but not re-emitted with full fidelity where the
  schema is unknown.

## Provenance

This is an independent, unofficial project. It is not affiliated with or endorsed
by AVEVA. The grammar was reconstructed by static analysis of the shipped
components and by measurement over real production files; anything inferred rather
than measured is marked as such where it is documented.
