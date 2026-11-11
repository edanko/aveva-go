# `aveva-go` `gen` package — conformance findings and improvement plan

Companion to [`gen-format-spec.md`](./gen-format-spec.md).

Every finding below was **reproduced** against the three shipped sample files in
`testdata/` using the current package. The reproduction commands and measured
numbers are included so each can be re-verified.

## Scope note — one project, two format families

> **Project lineage.** `aveva-go` is the successor to the former
> `github.com/edanko/gen`, and it absorbed the standalone `gen2dxf` CLI as
> `cmd/plate2dxf` (its `pkg/draw/*` became `cmd/plate2dxf/internal/draw/*`
> verbatim). The `gen2dxf` working copy was archived to
> `gen2dxf.tar.gz` and removed; it is superseded and should not be revived.

## The two format families are separate subsystems

The package exposes `ReadParts` (plate) and `ReadProfiles` (`LISTED_PROFILE`), so
it is tempting to assume they share a producer. They do not:

| Component | Plate | Profile (`.gen`) |
| --- | :---: | :---: |
| `mar.dll` / `mar_p.dll` (64.6 MB) | ✅ | ✅ |
| `TbGenPostProcessor.dll` (2.7 MB) | ✅ | ❌ |
| `fabricationcore.dll` | ❌ | ⚠️ XML exporter fixture only |

`TbGenPostProcessor.dll` — the artifact whose decompiled code underpins the
plate half of the spec — contains **none** of `LISTED_PROFILE`, `PROFILE_DATA`,
`MLENGTH`, `LEFT_END`, `RIGHT_END`, `START_OF_FCONTOUR`, `HOLES_NOTCHES_CUTOUTS`,
`ENDCUT_MASK`. Profile-family evidence comes from the sample plus the emission
literals recovered in place from `mar.dll` (Appendix A.1).

Practical consequence for this plan: **Bug 1 and Bug 5 are plate-only** and are
verified against decompiled shipped code; **Bug 2 and Bug 3 are profile-only**
and — since this note was first written — have now been **confirmed against
`mar.dll`** and their fixes proven on the shipped sample (see Appendix A). Do not
assume a fix validated on one family transfers to the other.

## Baseline

```
$ go test .                     # gen package only
ok      github.com/edanko/aveva-go   0.005s

$ go test ./...
FAIL    github.com/edanko/aveva-go/cmd/plate2dxf          # unrelated: dxf API
FAIL    github.com/edanko/aveva-go/cmd/plate2dxf/internal/app
FAIL    github.com/edanko/aveva-go/cmd/plate2dxf/internal/draw
```

The existing tests pass but **do not assert on parsed payload**. They only
assert `len(parts)` and `Quantity`, which is why Bug 1 and Bug 2 below were able
to go unnoticed.

Measured payload coverage of the current package:

| File | parts | with BurningData | with MarkingData | with StringData | with GeometryData | burning segments |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `nest.gen` | 394 | **0** | **0** | 0 | 0 | **0** |
| `part.gen` | 1 | 1 | 1 | 1 | 1 | 2421 |
| `profiles.gen` | 440 | — | — | — | — | holes **0**, ends **0** |

---

## Bug 1 — `ReadParts` silently discards the whole tail of a nest file

**Severity: critical.** This is the reason `nest.gen` yields 394 parts with zero
payload.

`plate.go`:

```go
default:
    for s.Scan() {
        switch s.Text() {
        case "END_OF_IDLE_DATA",
            "END_OF_BUMP_DATA",
            "END_OF_LABELTEXT_DATA",
            "END_OF_LABELSYMBOL_DATA",
            "END_OF_EDGE_DATA":
            break          // <-- breaks the switch, not the loop
        }
    }
```

`break` inside a `switch` exits only the `switch`. The enclosing `for s.Scan()`
keeps reading. So the **first unhandled top-level marker consumes the file to
EOF**.

`nest.gen` contains top-level `IDLE_DATA` markers (395 of them) interleaved with
`BURNING_DATA` and `PART_INFORMATION`:

```
line 65487  PART_INFORMATION
line 65491  IDLE_DATA          <-- parser enters `default` here
line 65506  END_OF_IDLE_DATA
line 65507  BURNING_DATA       <-- never reached
...
```

Because all 394 `PART_DATA` blocks precede the first `IDLE_DATA`, you get the
correct part count but lose **all 919 `BURNING_DATA` blocks, 2378 segments and
every `PART_INFORMATION` scope switch** after line 65491.

### Fix

Handle the ignorable sections explicitly and make the `default` arm a
structural terminator instead of a black hole:

```go
case "IDLE_DATA":
    if err := skipUntil(s, "END_OF_IDLE_DATA"); err != nil { return nil, err }
case "BUMP_DATA":
    if err := skipUntil(s, "END_OF_BUMP_DATA"); err != nil { return nil, err }
case "LABELTEXT_DATA":
    if err := skipUntil(s, "END_OF_LABELTEXT_DATA"); err != nil { return nil, err }
case "LABELSYMBOL_DATA":
    if err := skipUntil(s, "END_OF_LABELSYMBOL_DATA"); err != nil { return nil, err }
case "EDGE_DATA":
    if err := skipUntil(s, "END_OF_EDGE_DATA"); err != nil { return nil, err }

default:
    // A bare token with no '=' that we do not know is the end of an
    // enclosing block; a KEY=VALUE pair is a harmless unknown field.
    if !strings.Contains(s.Text(), "=") {
        return parts, nil
    }
}

func skipUntil(s *bufio.Scanner, end string) error {
    for s.Scan() {
        if s.Text() == end {
            return nil
        }
    }
    return s.Err()
}
```

With this change alone the payload is recovered:

| File | parts | with BurningData | burning segments |
| --- | ---: | ---: | ---: |
| `nest.gen` | 394 | **392** | **2378** |
| `part.gen` | 1 | 1 | 2421 |

`392` here is *not* correct — it is Bug 5 showing through: the name-keyed index
still collapses the one triply-nested part, so two of its three placements look
empty. Once Bug 5 is also fixed the count becomes **394** and every placement
carries its own payload (see Bug 5).

The `919` `BURNING_DATA` blocks are **not** 1:1 with parts. Final distribution in
`nest.gen` with Bugs 1 and 5 both fixed:

| `BURNING_DATA` blocks on a part | Parts |
| ---: | ---: |
| 0 | 0 |
| 1 | 59 |
| 2 | 159 |
| 3 | 165 |
| 4 | 8 |
| 5 | 3 |

(Σ = 919. The `12`-block row that appears with Bug 5 present is gone: it was one
part receiving three placements' worth of contours.)

Every block is `SHAPE=OUTER_CONTOUR` (919/919) — the extra blocks are **not**
holes. Correlation with bevels confirms they are multi-pass bevel cuts:

| | Blocks | Parts |
| --- | ---: | ---: |
| parts with no bevelled segment | 59 | 59 (exactly 1 block each) |
| parts with ≥1 bevelled segment | 860 | ~335 (2–12 blocks each) |

So a bevelled plate emits **one contour per bevel pass**; an unbevelled plate
emits exactly one. `BEVEL` kinds observed per segment: `NONE` 2032, `TS` 166,
`OS` 180; `BEVEL_CODE` values `0, ±116, ±145, ±273`.

### Recommended regression test

```go
func TestReadNestPayload(t *testing.T) {
    f, err := os.Open("testdata/nest.gen")
    require.NoError(t, err)
    defer f.Close()

    parts, err := ReadParts(f)
    require.NoError(t, err)
    require.Len(t, parts, 394)

    var withBurn int
    for _, p := range parts {
        if len(p.BurningData) > 0 {
            withBurn++
        }
    }
    require.Greater(t, withBurn, 380)
}
```

---

## Bug 2 — `ReadProfiles` never reads `LEFT_END` / `RIGHT_END`

**Severity: high.** **Status: CONFIRMED against `mar.dll`, and fixed-and-proven.**

`profiles.gen` yields 440 profiles with `LeftEnd == nil`,
`RightEnd == nil`, `Holes == nil` for every single one.

`readProfile` expects the end blocks **nested inside**
`PROFILE_DATA … END_OF_PROFILE_DATA`:

```go
case "LEFT_END":
    end, err := readEnd(s)
    ...
```

But the shipped `LISTED_PROFILE` writer emits `END_OF_PROFILE_DATA`
**before** `LEFT_END`:

```
line 169  END_OF_PROFILE_DATA
line 170  LEFT_END
...
line 333  END_OF_LEFT_END
line 334  RIGHT_END
...
line 497  END_OF_RIGHT_END
line 498  CONNECTION_TRACE
...
line 513  GEOMETRY_DATA
...
line 541  STRING_DATA
...
line 573  PROFILE_DATA          <-- next profile record starts here
```

`readProfile` hits `END_OF_PROFILE_DATA` (no `=`) and returns immediately.
`LEFT_END` is then seen by `ReadProfiles`' outer loop, which has no case for it,
so it is dropped.

**The profile record is a sequence of sibling blocks**, not one nested block. Its
extent is `PROFILE_DATA … (next PROFILE_DATA | REST_DATA | EOF)`.

### Fix

Restructure `ReadProfiles` so the current profile stays open across sibling
blocks:

```go
var cur *Profile

for s.Scan() {
    switch s.Text() {
    case "PROFILE_DATA":
        cur = &Profile{Shape: cd.Shape, Dimension: cd.Dimension, Quality: cd.Quality, Quantity: 1}
        if err := readProfileAttrs(s, cur); err != nil { return nil, err }
        profiles = append(profiles, cur)

    case "LEFT_END":
        if cur != nil { cur.LeftEnd, _ = readEnd(s) }
        else            { _ = skipBlock(s, "END_OF_LEFT_END") }

    case "RIGHT_END":
        if cur != nil { cur.RightEnd, _ = readEnd(s) }
        else            { _ = skipBlock(s, "END_OF_RIGHT_END") }

    case "HOLES_NOTCHES_CUTOUTS":
        if cur != nil { h, err := parseHole(s); ...; cur.Holes = append(cur.Holes, h) }
        else            { _ = skipBlock(s, "END_OF_HOLES_NOTCHES_CUTOUTS") }

    case "CONNECTION_TRACE":
        _ = skipBlock(s, "END_OF_CONNECTION_TRACE")

    case "GEOMETRY_DATA":
        g, err := readGeometry(s)
        if cur != nil { cur.Geometry = append(cur.Geometry, g) }

    case "STRING_DATA":
        sd, err := readStringData(s)
        if cur != nil { cur.StringData = append(cur.StringData, sd) }

    case "REST_DATA":
        _ = skipBlock(s, "END_OF_REST_DATA")
    }
}
```

Also move `Quantity = 1` to record creation (currently `readProfile` sets it at
the end, so any profile built outside that path defaults to `0`).

---

## Bug 3 — `-` sentinel makes `readEnd` fail on real files

**Severity: high (latent — masked only because of Bug 2).**
**Status: CONFIRMED — reproduced, and the sentinel semantics are now measured.**

`LISTED_PROFILE` end-cut numeric fields use a bare `-` for "not applicable":

```
A=-
B=-
C=-
R2=-
V3=-
V4=-
```

Counts in `profiles.gen`: `B`, `C`, `V3`, `V4` → 880 each; `A`, `R2` → 760 each.

`readEnd` does:

```go
case "A":
    a, err := strconv.ParseFloat(value, 64)
    if err != nil { return nil, err }
```

so the moment Bug 2 is fixed, `ReadProfiles` starts returning
`strconv.ParseFloat: parsing "-": invalid syntax` on the **first** profile.

### Fix

Introduce a nullable numeric type (or a `(value, defined)` pair) and a shared
helper:

```go
// parseOptFloat maps the format's "-" / "" sentinel to (0, false).
func parseOptFloat(s string) (float64, bool, error) {
    s = strings.TrimSpace(s)
    if s == "" || s == "-" {
        return 0, false, nil
    }
    v, err := strconv.ParseFloat(s, 64)
    return v, err == nil, err
}
```

Use it for every numeric field in `End`. Keep `defined` so callers can
distinguish "no end cut parameter" from "zero".

---

## Bug 4 — `SURFACE_TREATMENT` is written into `BlockNo`

**Severity: medium.** Copy-paste error in `profile.go`:

```go
case "SURFACE_TREATMENT":
    profile.BlockNo = value      // <-- should be profile.SurfaceTreatment

case "BLOCK_NO":
    profile.BlockNo = value
```

`profiles.gen` has `SURFACE_TREATMENT=` before `BLOCK_NO=10123`, so `BlockNo`
happens to end up correct, but any file where `SURFACE_TREATMENT` is non-empty
and `BLOCK_NO` is absent will produce a wrong block number.

`Profile` already declares `SurfaceTreatment string` — it is simply never
assigned.

---

## Bug 5 — `PART_INFORMATION` binding collapses repeated parts

**Severity: high — this corrupts the shipped sample, and the first fix attempted
for it was itself wrong.**

The original code keys parts by name into a single index:

```go
nameIndex[part.Name] = len(parts) - 1      // a repeated name keeps only the last index
...
case "PART_NAME":
    lastPartIdx = nameIndex[value]          // missing name -> 0, no error
```

### What the relation actually is

`PART_INFORMATION` is **not** 1:1 with `PART_DATA`, and it is not a function of
the name either. Measured on `nest.gen`:

```
PART_DATA blocks        394      PART_INFORMATION blocks  395
distinct names          392
"54949-DA4137_36X-3ZK"   PART_DATA x3   PART_INFORMATION x3    # one part, three nest positions
"54949-QW197_3X-21ZK"    PART_DATA x1   PART_INFORMATION x2    # one part, two consecutive scopes
```

So a part may be **placed several times** (one `PART_DATA` per position) and may
also be **opened by several scopes** (more `PART_INFORMATION` than `PART_DATA`).
The last two `PART_INFORMATION` in the file name the same part back to back.

### Consequence of the original code

All three placements of `54949-DA4137_36X-3ZK` resolve to the last index, so with
only Bug 1 fixed:

| | observed |
| --- | --- |
| 2 of the 3 placements | **0** `BURNING_DATA` — their geometry silently dropped |
| 3rd placement | **12** blocks instead of 4 — three placements' worth accumulated |
| histogram | contained a `12:1` outlier and a `0:2` row |

A missing name also yields index `0` (Go map zero value), attaching the payload
to part #0 rather than reporting anything.

### Correct model (implemented)

Track every placement of a name, bind scopes to them **in order**, and once a
name's placements are used up, accumulate further scopes onto its last placement
— which is exactly what the `54949-QW197_3X-21ZK` case requires.

```go
type placementsOf struct {
    indices []int
    used    int
}

// on PART_DATA:
pl := placements[part.Name]
if pl == nil { pl = &placementsOf{}; placements[part.Name] = pl }
pl.indices = append(pl.indices, len(parts)-1)

// on PART_INFORMATION:
pl := placements[value]
if pl == nil || len(pl.indices) == 0 {
    return nil, fmt.Errorf("gen: PART_INFORMATION references unknown part %q", value)
}
if pl.used < len(pl.indices) {
    lastPartIdx = pl.indices[pl.used]
    pl.used++
} else {
    lastPartIdx = pl.indices[len(pl.indices)-1]
}
```

> A first attempt used a strictly positional queue and errored when exhausted.
> That was **wrong**: it rejected the shipped file at
> `"54949-QW197_3X-21ZK"`, because that part legitimately has two scopes for one
> placement. The fallback-to-last-placement branch is required, not defensive
> padding.

### Result

| | original | Bug 1 only | Bugs 1 + 5 |
| --- | ---: | ---: | ---: |
| parts | 394 | 394 | 394 |
| parts with a cut path | 0 | 392 | **394** |
| `0`-block parts | 394 | 2 | **0** |
| `12`-block outlier | — | 1 | **none** |
| the 3 nested placements | — | 0,0,12 | **4, 4, 4** |
| total segments | 0 | 2378 | 2378 |

The three placements of the repeated part coming out identically sized (4, 4, 4)
is the strongest single check that the binding is right — they are the same part,
so they must have the same payload.

### Recommended regression test

Implemented as `TestNestRepeatedPartKeepsPayload` in `conformance_test.go`:

```go
require.Len(t, byName, 392)          // 392 distinct names
require.Equal(t, 1, repeated)        // exactly one name is placed more than once

for name, group := range byName where len(group) > 1 {
    want := len(group[0].BurningData)
    require.NotZero(t, want)
    for _, p := range group {
        require.Len(t, p.BurningData, want)   // identical placements, identical payload
    }
}
for _, p := range parts {
    require.NotEmpty(t, p.BurningData)
}
```

---

## Gap 1 — no `PART_INFORMATION` / `TYPE_OF_WORK` model

The format uses `PART_INFORMATION` as an explicit *scope switch* preceding each
part's payload. The Go package tracks `lastPartIdx` implicitly and has no type
for it. Introduce:

```go
type PartInformation struct {
    PartName   string
    TypeOfWork string // BURNING, ...
}
```

This makes the association explicit and makes the sibling-scoped sections
(`BURNING_DATA`, `MARKING_DATA`, `GEOMETRY_DATA`, `STRING_DATA`, `IDLE_DATA`)
structurally correct rather than order-dependent.

---

## Gap 2 — missing fields

The following format fields have no home in the current structs:

**`Part`** — `TYPE_OF_WORK`, `PARTNAME_LONG_SB`, `ASSEMBLY_LOW_SB`,
`ASSEMBLY_HIGH_SB`, `PART_COG_U/V` (only via pointer), `CUST_DATA_1..4`,
`CUST_TEXT_1..10`, `WORKING_LOCATION`, `BUILDING_STRATEGY`, `PLANNING_UNIT`,
`PANEL_NAME`, `CLIP_GRINDING_SIDE`, `COMMENT`, `DESTINATION`, `PART_ID`,
`FUNCTIONAL_DESCR_CODE`, `FUNCTIONAL_DESCR_TEXT`.

**`Part` (from `GENERAL_DATA`)** — `TYPE_OF_GENERIC_FILE`, `TYPE_OF_MANUFACT`,
`BURNER_ID`, `NEST_NAME`, `RAW_NAME`, `RAW_LENGTH`, `RAW_WIDTH`, `RAW_AREA`,
`DENSITY`, `SURF_TREAT_TS/OS`, `REST_LENGTH`, `PRODUCTION_DATE`, `TOTAL_IDLE`,
`TOTAL_MARKING`, `TOTAL_BLASTING`, `TOTAL_NO_OF_STARTS`,
`NO_OF_*_STARTS`, `STARTING_POS_X/Y`, `QUANTITY_NORMAL`, `QUANTITY_MIRRORED`,
`TOOL_ID`, `PROCESS_ID`, `SEQUENCE`, `DISTANCE_Y1_Y2`, `CUTTING_GAS`,
`MATERIAL_CODE`, `PURCHASE_INFO`, `PLATE_ALIGNMENT`, `SPLIT_DATE`,
`TOTAL_MARK_BLAST`, `NO_OF_MARK_BLAST_STARTS`, `VERIFICATION_DATE`.

Only `RAW_THICKNESS`, `QUALITY` and `NO_OF_PARTS` are currently modelled from
the header, and `NO_OF_PARTS` is only used for slice capacity.

**`Marking`** — `COMMENT`, `INCLINATION_ANGLE` is present, but `COMMENT` is
absent.

**`BurningData`** — `HOLE_SHIPSIDE`.

**`End`** — `ENDCUT_USER_DEFINED_NAME`, `ENDCUT_USER_DEFINED_PARAM1..7`,
`GSD`, `GSD_DIST`.

**`Profile`** — `GENERATED_AS`, `APPROX_PART_WEIGHT`, `IDENT_STRING`,
`PARTNAME_SHORT`, `SHIP_NAME`, `SIDE_UP`, `BRAND`, `ASSEMBLY_SEQUENCE`,
`BAUGRUPPE`, `ZEICHNUNGSNO`, `ZAEHLNO`, `NO_OF_REF_LINES`, `ADDITIONAL_INFO1..12`,
`MARKING_LEFT_EXCIST`, `MARKING_RIGHT_EXCIST`, `NO_OF_MACS`, `NO_OF_MARKS`.

The most valuable additions for a downstream consumer are the `_SB` variants
(shipside/both-sides naming is fundamental to shipbuilding output) and the
`CUST_*` passthrough block.

### Dead field

`Profile.Milling` is parsed from `MILLING=` but that token is **absent from every
binary in the installation** and from all three samples. It is a field no AVEVA
writer emits. Either drop it or document it as reserved; do not let it imply the
format has a milling attribute. (`GRINDING=` *is* real — present in `mar.dll` and
`profiles.gen`.)

---

## Gap 3 — `Profile.LeftEnd`/`RightEnd` need the contour data

**Resolved by verification.** `readEnd` already populates `Contour` and
`FlangeContour` correctly, and with Bugs 2 + 3 fixed all four contours populate
for all 440 profiles (Appendix A.6) — no new contour code is needed. This is no
longer a gap; it only needs the regression assertions:

```go
require.NotNil(t, profs[0].LeftEnd.Contour)
require.NotNil(t, profs[0].LeftEnd.FlangeContour)
require.NotNil(t, profs[0].RightEnd.Contour)
require.NotNil(t, profs[0].RightEnd.FlangeContour)
```

---

## Gap 4 — the package does not distinguish document families

`ReadParts` should reject a `LISTED_PROFILE` document and `ReadProfiles` should
reject a `PLATE_PART` document based on `TYPE_OF_GENERIC_FILE`, rather than on
"did `GENERAL_DATA` appear" / "does line 1 equal `TYPE_OF_GENERIC_FILE=LISTED_PROFILE`".

Current behaviour: `ReadParts` on `profiles.gen` returns `(nil, nil)` because
`readGeneralData` fails and `ReadParts` does `return nil, nil` — it swallows the
error. `ReadProfiles` on `nest.gen` returns `(nil, nil)` because the first line
does not match. Both silently return `nil` with a `nil` error, which is
indistinguishable from "valid empty file".

Recommended: introduce a sentinel

```go
var ErrWrongDocumentType = errors.New("gen: document is not a plate part nest")
```

and return it, or at minimum document the `(nil, nil)` contract. Also stop
swallowing the `readGeneralData` error:

```go
generalData, err := readGeneralData(s)
if err != nil {
    return nil, fmt.Errorf("gen: %w", err)   // not `return nil, nil`
}
```

---

## Gap 5 — contour helpers are unsafe on empty contours

`contour.go`:

```go
func (c Contour) Vertices() [][]float64 {
    vertices := make([][]float64, len(c.Segments)+1)
    vertices[0] = c.Segments[0].Start.Vertice()   // panics if len == 0
```

`Bulges()` has the same shape (`c.Segments[0]` unguarded). `IsClockwise()`
correctly guards with `len(c.Segments) == 0`. Add the same guard to `Vertices`
and `Bulges`, and guard `Segments[i].Start`/`.End` for nil (the `invert` path
already assumes non-nil).

---

## Gap 6 — `Part.Mirror` mutates pointer fields but not all of them

`part.go` `Mirror()` iterates `p.StringData`, `p.BurningData`, `p.MarkingData`,
`p.GeometryData` **by value** (`for _, s := range ...`), so reassignments such as
`s.Pos.X = -s.Pos.X` work only because `Pos`, `Contour` and `Bevel` are
pointers. `p.BurningData[i].Direction` (a plain `int`, which also has a
geometric meaning — 0/1 traversal direction) is not flipped, and
`Marking.MarkingDirection` similarly. Confirm from the format spec §3.5/§3.6
whether `DIRECTION` is geometry-bearing before relying on `Mirror()` for
production output.

Additionally `Mirror()` dereferences `p.PartCog` unconditionally; it is only
non-nil because `readPart` always allocates it — document that invariant or
guard it.

---

## Gap 7 — no writer, and re-serialisation would not be byte-faithful

The package is read-only. If a writer is added, be aware that the two producers
use **different numeric formatting** (spec §8): the plate writer pads to fixed
`%.2f`/`%.5f`, the `LISTED_PROFILE` writer emits minimal digits
(`U=0`, `U=7`, `U=38.5`). A single formatter cannot reproduce both.

Two further round-trip hazards:

* Plate coordinates are stored to 5 decimals, so `parse → format` is lossy at
  ~1e-5. Detect and refuse to "normalise" geometry that was already truncated.
* `ENDCUT_MASK` is a zero-padded bit string (`00010110000`). It must stay a
  `string`; parsing it as an integer destroys the leading zeros and changes the
  bit positions.

---

## Gap 8 — three whole profile sections are not modelled

Recovered from `mar.dll` emission literals (Appendix A.1), the `LISTED_PROFILE`
record contains sections the package never reads:

| Section | Terminator | Fields |
| --- | --- | --- |
| `MARKING_LEFT_END` | `END_OF_MARKING_LEFT_END` | `SEGMENT`, `START_OF_FCONTOUR`/`END_OF_FCONTOUR`, `START_OF_FCONTOUR2`/`END_OF_FCONTOUR2` |
| `MARKING_RIGHT_END` | `END_OF_MARKING_RIGHT_END` | same |
| `REF_LINE_MARKING` | `END_OF_REF_LINE_MARKING` | `REF_NO=`, `ORDER_NO=`, `DIST_LEFT=` |

Also present and unmodelled: `BENDING_TABLE` (`NO_OF_POINTS=`, documented limit
100 points), `GENERAL_MARKING_DATA` (`ATTR=`), `CURVE_DATA` (with
`START_OF_MCONTOUR`/`END_OF_MCONTOUR` and `START_OF_NCONTOUR`/`END_OF_NCONTOUR`),
`CURVE_3D_DATA`.

Because `ReadProfiles` currently skips unknown markers in its outer loop, these
are silently dropped rather than mis-parsed — but they are real per-part
production data.

---

## Gap 9 — profile holes are richer than `parseHole`

`HOLES_NOTCHES_CUTOUTS` carries shared hole attributes (`OPEN=`,
`DIST_ORIGIN=`, `DIST_ORIGIN_TO_ZEROP=`, `DIST_ORIGIN_V=`, `ROTATION=`) and then
one **sub-block** that `parseHole` does not model:

| Sub-block | Terminator | Fields |
| --- | --- | --- |
| `PENETRATING_PROFILE` | `END_OF_PENETRATING_PROFILE` | `DIM=`, `INTERSECTING_ANGLE=`, `INCLINATION_ANGLE=`, `MTRL_PENETRATING=` |
| `DRILLING` (contains `THREADING`) | `END_OF_DRILLING` / `END_OF_THREADING` | `THREAD_START=` |
| `SLOTTED` | `END_OF_SLOTTED` | `LENGTH=`, `WIDTH=` |

`parseHole` also reads `TYPE`, `MIRROR`, `NAME`, `DIST_LEFT` which are plate-side
keys; the profile side uses `OPEN`/`DIST_ORIGIN`/`ROTATION`. The two hole models
should probably be separate types rather than one shared `Hole`.

---

## Prioritised plan

Status as of the last pass: items 1–5 and 9 are **implemented** in the working
tree, with regression tests in `conformance_test.go`. `go test .` (8 tests) and
`go vet .` are clean, `gofmt -l .` is empty. Items 6–8 and 11–14 remain open.

One robustness detail added beyond the plan: the Bug 1 `default` arm also treats
blank lines as ignorable. The shipped samples contain no blank lines, so no
fixture exercises it, but a hand-edited or double-terminated file would otherwise
truncate. Covered by the synthetic `TestReadPartsToleratesBlankLines`.

| # | Change | Value | Status |
| --- | --- | --- | --- |
| 1 | `ReadParts` tail-consumption (Bug 1) | **critical** | ✅ done |
| 2 | `ReadProfiles` sibling end blocks (Bug 2) | **high** | ✅ done |
| 3 | `-` sentinel (Bug 3) | **high** | ✅ done |
| 4 | `SURFACE_TREATMENT` field (Bug 4) | medium | ✅ done |
| 5 | Placement-aware `PART_INFORMATION` (Bug 5) | **high** | ✅ done |
| 6 | `PartInformation` type + scope list (Gap 1) | medium | ✅ done |
| 7 | Missing struct fields (Gap 2) | medium | ✅ done |
| 8 | Document-type discrimination + errors (Gap 4) | medium | ✅ done |
| 9 | Contour helper empty guards (Gap 5) | low | ✅ done |
| 10 | `Mirror()` completeness + documented scope (Gap 6) | low | ✅ done |
| 11 | `ENDCUT_MASK` stays a string (Gap 7) | low | ✅ already correct |
| 12 | `MARKING_*_END`, `REF_LINE_MARKING`, `BENDING_TABLE`, `CURVE_DATA` (Gap 8) | medium | ✅ generic capture |
| 13 | Structured hole sub-blocks (Gap 9) | medium | ✅ modelled, unvalidated |
| 14 | `cmd/plate2dxf` dxf v2 build (Appendix B) | blocker | ✅ done |
| 15 | `LABELTEXT_DATA` / `LABELSYMBOL_DATA` / `EDGE_DATA` / contour `AUXILIARY_DATA` | **high** | ✅ done (new — see Appendix D) |

No gaps remain open. See Appendix D for the closure evidence.

---

## Implemented API

The gaps closed above added these exported surface changes:

```go
// Document keeps the blocks that are not tied to one part.
type Document struct {
    Type, Manufact string
    General        *GeneralData
    Parts          []*Part
    LabelTexts     []LabelText   // nest-wide labelling pass
    LabelSymbols   []LabelSymbol
}

func ReadDocument(r io.Reader) (*Document, error)  // full document
func ReadParts(r io.Reader) ([]*Part, error)       // convenience: doc.Parts
func DocumentType(r io.Reader) (string, error)     // TYPE_OF_GENERIC_FILE

func ReadProfileDocument(r io.Reader) (*ProfileDocument, error)
func WriteProfiles(w io.Writer, doc *ProfileDocument) error

func WriteDocument(w io.Writer, d *Document) error // plate family
func WriteParts(w io.Writer, parts []*Part) error

var ErrNotPlatePart     = errors.New("gen: not a plate part document")
var ErrNotListedProfile = errors.New("gen: not a listed profile document")
```

New types: `PartInformation`, `LabelText`, `LabelSymbol`, `Auxiliary`, `Edge`,
`Field`, `RawSection`, `PenetratingProfile`, `Drilling`, `Slotted`.

New fields on existing types: `Part` gained the `_SB` shipside variants, the
`CUST_DATA`/`CUST_TEXT` blocks, header-derived attributes and the new payload
slices; `GeneralData` now carries the whole header (37 fields) rather than three;
`Marking` gained `Comment`; `BurningData` gained `HoleShipside`; `End` gained the
end-cut identity, user-defined parameters and the flange/knuckle extras;
`Contour` gained `Auxiliaries`; `Profile` gained ~45 attributes and `RawSections`,
and lost the dead `Milling` field.

`go.mod` now targets `go 1.27.0` (required by `edanko/dxf`) and pins
`github.com/edanko/dxf v1.0.1-0.20261111120207-5f9756961797` — the v2.0.0 commit,
which cannot be required by its literal tag (Appendix B.2).

## Evidence index

| Claim | Evidence |
| --- | --- |
| Artifact identity `TbGenPostProcessor.dll` | REA `ev_e39ae48e1239e088e8609eb42bed075dcc0791530171ee7f3c794d8781c98c76` (binary_overview) |
| Complete key→token vocabulary | REA `ev_59161c8b4dc421c4ea61e2ddd5267f874bcdd93987633f7f34e16e105e0c2654` (pseudocode `FUN_1014b0e0`) |
| `KEY=VALUE` split + trim + token store | REA `ev_4ec36fbc52e3c9860a4bcbf6acc02a6dee7a6c5d3f34da48953a96d8ae9ef0ce` (`FUN_1014d210`) |
| Section→class dispatch, `PLATE_PART_CHECK` template | REA `ev_3bed506262a4d0bccb7c63aab21c1e91e1e9884031b7cbb267dfb81f8ff04029` (`FUN_10035750`) |
| `GENERAL_DATA` field order + types | REA `ev_9cd51c09be0607e7eb053776f66ce3f5c6cf06afff7a5ec56f287afe3ef5d424` (`FUN_1015a4b0`) |
| `PART_DATA` field order + types | REA `ev_67fed031d31d7030391726231389913e015d3eea528b41ff9573fcda1314a4a5` (`FUN_1015fbc0`) |
| `MARKING_DATA` field order + types | REA `ev_15ad32467d29d524f8f90766c63e46d0b769efee394c16d61b01cf0e8799e677` (`FUN_1015e4b0`) |
| Contour write order (`NO_OF_SEG`/`START_U`/`START_V`) | REA `ev_6317232a609c6f28ed468d5e9687f76e8ccb09320f77804b306101511a608446` (`FUN_10157380`) |
| `PinGpp_CGenericBase` / `PinGpp_CGenericProperties` RTTI | REA `ev_1df30740da5bda8136222061bd402fb837c633c7a97d97cd3a4ce197a610e790` |
| Numeric format strings `%.5f`, `%.2f`, `=%d` | `strings` over `TbGenPostProcessor.dll` (`NUMBER_OF_HEADS=%d`, `DISTANCE_Y1_Y2=%.5f`, `START_U=%.5f`, `RAW_LENGTH=%.2f`), plus measured decimals-per-key over `testdata/*.gen` |
| Bug 1 / Bug 2 measurements | `go run` harness against `testdata/nest.gen`, `testdata/part.gen`, `testdata/profiles.gen`; patched copy in `/tmp/genfix` |

---

## Appendix A — verification of Bugs 2 and 3 against the implementation

Both bugs were originally flagged as *sample-derived* because the implementing
binary is `mar.dll`, not the decompiled `TbGenPostProcessor.dll`. They have now
been verified two independent ways.

### A.1 Method

`mar.dll` was **not** imported into Ghidra — its `.text` is 51.7 MB, which is not
a practical deep-analysis target here:

```
.text  vsize=51.72MB  raw=51.72MB
.rdata vsize= 4.68MB  raw= 4.68MB
.data  vsize=70.56MB  raw= 2.47MB
```

Instead, each profile key occurs **exactly once** as a NUL-terminated literal in
`.rdata`. MSVC emits string literals in first-reference order, so reading them in
file order recovers the writer's emission sequence — the same method that
produced the plate schema from `TbGenPostProcessor.dll`. Every literal carries
its own `=`, e.g. `MLENGTH=`, confirming a `write("MLENGTH="); write(value);`
style serialiser.

Recovered locations (image base `0x10000000`):

| Literal | `.rdata` VA | `.rdata` file off |
| --- | --- | --- |
| `PROFILE_DATA` | `0x132521c4` | `0x3250bc4` |
| `END_OF_PROFILE_DATA` | `0x13252628` | `0x3251028` |
| `LEFT_END` | `0x1325263c` | `0x325103c` |
| `HOLES_NOTCHES_CUTOUTS` | `0x132553f0` | `0x3253df0` |
| `CONNECTION_TRACE` | `0x13255650` | `0x3254050` |
| `REST_DATA` | `0x13252ed0` | `0x32518d0` |

### A.2 Bug 2 — confirmed

`END_OF_PROFILE_DATA` (`0x3251028`) and `LEFT_END` (`0x325103c`) are **adjacent**
literals: the gap is exactly 20 bytes, i.e. `"END_OF_PROFILE_DATA\0"`. The
attribute block's last field `CASE=` is emitted immediately before, and
`LEFT_END` immediately after, in the recovered literal run:

```
CASE=
END_OF_PROFILE_DATA
LEFT_END
LEFT_ORIGIN=
...
END_OF_LEFT_END
```

Independent of literal ordering, the *structural* claim is directly observable in
the sample itself: `profiles.gen` line 169 is `END_OF_PROFILE_DATA`, line 170 is
`LEFT_END`. Both lines are shipped artifact content, not inference.

The same recovery also revealed **three sections the Go package does not model at
all**: `MARKING_LEFT_END`, `MARKING_RIGHT_END`, `REF_LINE_MARKING` — plus
structured hole sub-blocks (`PENETRATING_PROFILE`, `DRILLING`/`THREADING`,
`SLOTTED`), `BENDING_TABLE`, `GENERAL_MARKING_DATA`, `CURVE_DATA` and
`CURVE_3D_DATA`. See spec §6.6–§6.9.

### A.3 Bug 2 — fix proven on the shipped sample

A throwaway copy was patched to treat `LEFT_END`/`RIGHT_END` as siblings of
`PROFILE_DATA` and nothing else:

| | before | after structural fix only |
| --- | ---: | ---: |
| profiles | 440 | 440 |
| `LeftEnd` populated | **0** | *blocked by Bug 3* |
| `RIGHT_END` populated | **0** | *blocked by Bug 3* |

### A.4 Bug 3 — confirmed as latent behind Bug 2

With only the structural fix applied and `strconv.ParseFloat` untouched, the
parser fails immediately:

```
ERROR: strconv.ParseFloat: parsing "-": invalid syntax
```

This is the exact predicted interaction: Bug 2 masks Bug 3, and fixing Bug 2
alone converts silent data loss into a hard parse failure.

### A.5 Bug 3 — sentinel semantics measured

The sentinel is not arbitrary. Across **all 880** end blocks in `profiles.gen`:

| `ENDCUT_TYPE` | `A` | `R2` | `B`,`C`,`V3`,`V4` | blocks |
| --- | --- | --- | --- | ---: |
| 21 | `-` | `-` | `-` | 760 |
| 27 | `42` | `42` | `-` | 120 |

100 % correlation, no exceptions. So `-` means *"this end-cut type does not use
this parameter"*. `B`, `C`, `V3`, `V4` are `-` on every block — reserved/unused
in this configuration. `A`/`R2` are `-` precisely when `ENDCUT_TYPE=21` and carry
a real radius (`42`) when `ENDCUT_TYPE=27`.

Consequence for the fix: `-` **must not** be flattened to `0` without recording
"undefined", because `0` is a legal radius and the distinction is meaningful.

### A.6 Both fixes together — full recovery

| | before | after |
| --- | ---: | ---: |
| profiles | 440 | 440 |
| `LeftEnd` populated | **0** | **440** |
| `RightEnd` populated | **0** | **440** |
| `LeftEnd.Contour` | 0 | **440** |
| `LeftEnd.FlangeContour` | 0 | **440** |
| `RightEnd.Contour` | 0 | **440** |
| `RightEnd.FlangeContour` | 0 | **440** |
| `ENDCUT_TYPE` histogram | n/a | `21`→380, `27`→60 (per end) |

The `ENDCUT_TYPE` histogram matches the spec exactly, and the end profiles parse
through the *existing* `readContour` — no new contour code was required.

### A.7 Residual unknowns after verification

* Per-field **types** beyond what literals + samples show. The literal run gives
  order and spelling; it does not prove e.g. whether `CASE=` is numeric.
* `MARKING_LEFT_END` / `MARKING_RIGHT_END` / `REF_LINE_MARKING` /
  `CURVE_DATA` / `CURVE_3D_DATA` / `BENDING_TABLE` field semantics — never
  present in any available sample.
* Whether `END_OF_PROFILE_DATA` may ever be omitted (older writer versions).
* The `,;` dialect in the same binary (spec §6.10) is a separate serialiser and
  was deliberately not modelled.

---

## Appendix B — `cmd/plate2dxf` build repair and the `dxf` v2.0.0 tag

`cmd/plate2dxf` did not compile against the current `edanko/dxf`. Four distinct
defects were involved; only one of them was the API rename.

### B.1 `LwPolylineWithBulges` does not exist

`draw_plate.go` called `d.drawing.LwPolylineWithBulges(false, vertices, bulges)`
at two sites. That method **never existed in any commit** of the `dxf` repo
(`git log -S LwPolylineWithBulges --all` is empty), so this code could not have
compiled against any version of it.

The current API is:

```go
func (d *Drawing) LwPolyline(closed bool, vertices ...[]float64) (*entity.LwPolyline, error)
```

with `entity.LwPolyline.Bulges []float64` as a public field that `Format()`
emits as group code 42. The port is therefore:

```go
polyline, err := d.drawing.LwPolyline(false, vertices...)
if err != nil { panic(err) }
polyline.Bulges = bulges
```

Verified functional, not merely compiling: a run over `testdata` produced 63
DXFs containing **1548 `LWPOLYLINE` bulge records, all within `[-1, 1]`**
(min `-0.999999`, max `0.416175`). Code-42 values outside `LWPOLYLINE` (e.g.
`20.0`, one per file) are DIMSTYLE/TABLE entries, not bulges.

### B.2 `github.com/edanko/dxf@v2.0.0` is not importable

The v2.0.0 tag points at the same commit as `main` (`5f97569`), but its `go.mod`
declares `module github.com/edanko/dxf` — **without the `/v2` suffix Go
requires** for a v2 tag. All of these fail:

```
$ go get github.com/edanko/dxf@v2.0.0
go: github.com/edanko/dxf@v2.0.0: invalid version: module contains a go.mod
    file, so module path must match major version ("github.com/edanko/dxf/v2")
$ go get github.com/edanko/dxf@v2.0.0+incompatible   # same error
$ go get github.com/edanko/dxf/v2@v2.0.0             # same error
```

So the dependency cannot be pinned to the literal tag. What *does* work is
pinning the same commit, for which Go derives a v1-line pseudo-version:

```
$ go get github.com/edanko/dxf@5f97569
go: upgraded github.com/edanko/dxf v1.0.0 => v1.0.1-0.20261111120207-5f9756961797
```

`go.mod` now requires `github.com/edanko/dxf v1.0.1-0.20261111120207-5f9756961797`
— byte-identical to v2.0.0, with the local `replace` removed.

**To make a literal `@v2.0.0` work** the `dxf` repo needs one change and a retag:
set `module github.com/edanko/dxf/v2` in its `go.mod`, then publish `v2.0.1`.
That also forces callers to import `github.com/edanko/dxf/v2/...`.

### B.3 `dxf v1.0.0` is a different, older thing

`v1.0.0` predates the module and is the old `yofu/dxf` wrapper — its entity types
resolve to `github.com/yofu/dxf/entity` and have no `SetColor`. It is not a
usable pin for this code.

### B.4 Three further CLI defects found while verifying

1. **`flag.Parse()` was never called.** Both `-i` and `-o` were silently ignored:
   `-i testdata` still walked `.`, and `-o` was dead. Fixed by calling
   `flag.Parse()` in `main`.
2. **`outputDir` was never threaded into `Process`.** Output was always created
   relative to the CWD (leaving `<timestamp> <block> dxf/` directories in the
   repo). Fixed by passing it through and joining into the output path.
3. **The logger discarded everything and failures exited 0.**
   `slog.NewTextHandler(io.Discard, …)` plus `logger.Error("error", err)` (a
   malformed slog call — the panic-level argument is not a key/value pair) meant
   a failed run printed nothing and exited successfully. Now logs to stderr with
   `"error", err` and exits 1.

`go build ./...`, `go vet ./...` and `go test ./...` are all clean.

---

## Appendix C — validation at scale (45,750 files)

Run against `/home/egor/Desktop/testdata`: **45,750 `.gen` files, 761 MB**,
median 6 KB, largest 3.3 MB, deepest document 650 parts.

### C.1 Corpus result

| Metric | Value |
| --- | ---: |
| files | 45,750 |
| classified `PLATE_PART` (has `GENERAL_DATA`) | 45,744 |
| classified `LISTED_PROFILE` | 6 |
| unclassified | **0** |
| parse errors | **0** |
| wall clock (cold cache) | 810 ms |
| wall clock (warm cache) | 303 ms |
| throughput | ~150,000 files/s, 32 workers |

| Payload | Value |
| --- | ---: |
| parts | 76,691 |
| parts with `BurningData` | 76,656 |
| parts with `MarkingData` | 15,841 |
| parts with `GeometryData` | 21,065 |
| parts with `StringData` | 44,975 |
| burning segments | 748,841 |
| bevelled segments | 125,356 |
| empty contours | 0 |
| parts with no payload at all | 24 (all in one file — see C.3) |
| files with duplicate part names | 12 |

| Profiles | Value |
| --- | ---: |
| profiles | 413 |
| with `LeftEnd` | **413** |
| with `RightEnd` | **413** |
| `ENDCUT_TYPE` spread | 11×10, 21×330, 23×14, 27×59 |

All 413 profiles carrying both ends is the scale confirmation of the Bug 2 / Bug 3
fixes: before them this figure was **0**.

### C.2 Thirteen real files failed — fixed

The first run produced **13 errors** on the error path added by the Bug 5 fix:

```
13  ReadParts: gen: PART_INFORMATION references unknown part "..."
```

Every one was a legitimate document the parser refused. Two distinct causes:

| Cause | Files | Shape |
| --- | ---: | --- |
| Empty `PART_NAME` scope | 7 | `PART_NAME=` with `TYPE_OF_WORK=LABELLING`; always followed by `IDLE_DATA` |
| No `PART_DATA` at all | 6 | `NO_OF_PARTS=1`, attributes only in `GENERAL_DATA`, part declared by `PART_INFORMATION` |

Both are documented in spec §3.3. Fixing them recovered data that the abort had
been discarding — the seven empty-name files previously died mid-file, so their
parts were never counted at all:

| | before fix | after fix |
| --- | ---: | ---: |
| errors | 13 | **0** |
| parts | 76,107 | **76,691** (+584) |
| burning segments | 743,644 | **748,841** (+5,197) |

> This is the second time the Bug 5 area has needed correcting: the original
> name-keyed map lost geometry, my first replacement (strict positional queue)
> rejected the shipped fixture, and the second replacement (error on unknown
> name) rejected 13 real-world files. The lesson is recorded deliberately: this
> relation is **data-driven, not schema-driven**, so the reader must tolerate
> every shape it encounters rather than enforce one.

### C.3 The 24 empty-payload parts are correct, not a bug

All 24 are in `54949/20-169816-03250-N03.gen`, which declares 25 `PART_DATA`
blocks but emits `BURNING_DATA`/`MARKING_DATA` for only one of them (3 scopes,
all naming `54949-NO_SP-2P`). A sibling file `10123/20-169816-03250-N03.gen`
with the same basename declares the same 25 parts and *does* produce for all of
them (39 scopes, 400 `BURNING_DATA`), and there the parser distributes payload
correctly — 4–30 burning blocks and 4–78 segments per part.

So "declared but not produced" is normal document content. Verified explicitly
rather than assumed, because the first suspicion was a binding defect.

### C.4 Correction to the sample-derived enumeration

The three shipped fixtures suggested `ENDCUT_TYPE ∈ {21, 27}`. The corpus shows
**`{11, 21, 23, 27}`**. The sentinel handling (`parseOptFloat`) is
type-agnostic, so no code change was needed — but the enumeration table in
spec §7 was wrong and is now corrected. This is a good illustration of why the
spec marks sample-derived claims as such.

### C.5 Regression tests added from the corpus

Three shapes now pinned by `conformance_test.go`:

* `TestReadPartsWithoutPartData` — synthesizes the part from the scope name plus
  `GENERAL_DATA`, and still supports `Mirror()` (which dereferences `PartCog`).
* `TestReadPartsEmptyScopeName` — a nameless scope's payload is dropped, **not**
  attached to the previous part.
* `TestReadPartsOverReferencedScope` — more scopes than placements accumulate.

### C.6 Reproduction

```go
// walk every *.gen, dispatch on the first line, compare against the totals above
parts, err := gen.ReadParts(bytes.NewReader(data))
profs, err := gen.ReadProfiles(bytes.NewReader(data))
```

---

## Appendix D — closing the remaining gaps

### D.1 The largest gaps were not in the list

Validating against the corpus re-ranked the work. Four sections that the plan
had **not** flagged turned out to be the biggest real data losses:

| Section | Corpus files | Before | Now |
| --- | ---: | --- | --- |
| `LABELTEXT_DATA` | **4,362** | skipped | `Part.LabelTexts` / `Document.LabelTexts` |
| `LABELSYMBOL_DATA` | **4,356** | skipped | `Part.LabelSymbols` / `Document.LabelSymbols` |
| `AUXILIARY_DATA` (inside contours) | 201 | skipped | `Contour.Auxiliaries` |
| `EDGE_DATA` | 124 | skipped | `Part.Edges` |

`AUXILIARY_DATA` sits **inside** `START_OF_CONTOUR … END_OF_CONTOUR`, after the
segments — `readContour` already consumed it but threw it away. It is not a
`BURNING_DATA` child.

`LABELTEXT_DATA` carries glyph strokes as nested `GEOMETRY_DATA` blocks, one per
`TEXT_ELEMENTS` element, so labels now retain their vector outlines.

### D.2 The labelling pass belongs to the document

A nest-wide labelling pass opens with a nameless scope and then emits every
label in the nest:

```
PART_INFORMATION
PART_NAME=                 <- no owning part
TYPE_OF_WORK=LABELLING
END_OF_PART_INFORMATION
IDLE_DATA ...
LABELTEXT_DATA (x N, one per part in the nest)
```

Attributing those to "the previous part" would be wrong, and dropping them loses
them. They are exposed on a new `Document` via `ReadDocument`; `ReadParts` keeps
its `[]*Part` signature and remains the simple entry point. This is why the
counts reconcile: `ReadDocument` sees 37,646 label blocks, exactly the corpus
population.

### D.3 Tolerant numerics are mandatory, not defensive

The corpus contains `RAW_LENGTH=`, `REST_LENGTH=`, `TOOL_ID=`,
`NO_OF_MARKING_STARTS=` — empty numeric values. With only three header fields
parsed this never surfaced; parsing the full header would have failed on the
first such file. All numeric parsing now goes through `parseOptFloat` /
`parseOptInt`, which map both the `-` sentinel and the empty value to 0.
Covered by `TestReadPartsTolerantNumerics`.

### D.4 Sections with no sample are captured, not guessed

Fourteen sections named by the `mar.dll` writer occur in **zero** of the 45,750
corpus files: `MARKING_LEFT_END`, `MARKING_RIGHT_END`, `REF_LINE_MARKING`,
`BENDING_TABLE`, `GENERAL_MARKING_DATA`, `CURVE_DATA`, `CURVE_3D_DATA`,
`PENETRATING_PROFILE`, `DRILLING`, `THREADING`, `SLOTTED`,
`MARKING_SHAPE_DATA`, `BLASTING_MARKING_DATA`, `BLASTING_DATA`, `GSD_DATA`.

The hole sub-blocks are modelled explicitly from the binary's emission literals
(`Hole.Kind` + `PenetratingProfile` / `Drilling` / `Slotted`) and documented as
**unvalidated**. Everything else is captured verbatim into
`RawSection{Name, Fields []Field}` so no data is discarded, rather than
inventing a schema no sample can check.

### D.5 `Mirror()` scope is now evidenced

`DIRECTION` is **not** a function of mirroring: over the corpus, `MIRRORED=0`
and `MIRRORED=1` documents are both dominantly `DIRECTION=1`, and both contain
`-1`. `Mirror()` therefore deliberately does **not** flip `Direction` /
`MarkingDirection`, and now says so. It does flip contour winding, TS/OS marking
sides, and every absolute U coordinate — including `LabelTexts`,
`LabelSymbols` and `Edges`, which are new. All pointer dereferences are nil-guarded.

### D.6 Document families report correctly

`ReadParts` returns `ErrNotPlatePart` and `ReadProfiles` returns
`ErrNotListedProfile` for the wrong family, instead of an empty result
indistinguishable from a valid payload-free document. `DocumentType` reports
`TYPE_OF_GENERIC_FILE` without full parsing. `cmd/plate2dxf` now skips other
families rather than failing on a mixed directory.

### D.7 Final corpus result

| Metric | Parsed | Independent scan |
| --- | ---: | ---: |
| files | 45,750 | 45,750 |
| classified / unclassified | 45,750 / **0** | — |
| parse errors | **0** | — |
| parts | 76,691 | — |
| burning segments | 748,841 | — |
| `LABELTEXT_DATA` | 37,646 | 37,646 ✓ |
| label geometry | 424 | 424 ✓ |
| `LABELSYMBOL_DATA` | 4,593 | 4,593 ✓ |
| `AUXILIARY_DATA` | 16,122 | 16,122 ✓ |
| `EDGE_DATA` | 496 | 496 ✓ |
| profiles / ends | 413 / 413+413 | ✓ |

Every newly-modelled count matches the independent corpus population exactly,
which is the check that the parsers are complete rather than merely non-failing.

---

## Appendix E — the writer, and what decompilation actually bought

### E.1 Why a writer is the thing decompilation enables

Everything else in this package could be derived from samples alone. A writer
cannot: the **field order** and the **numeric widths** a valid `.gen` must use are
not observable from a parser, and the samples disagree with each other because
different AVEVA writers emit different orders.

Three decompiled routines supplied exactly that:

| Routine | What it fixed |
| --- | --- |
| `CGeneralData::Write` (`FUN_1015a4b0`) | `GENERAL_DATA` field order |
| `CPartData::Write` (`FUN_1015fbc0`) | `PART_DATA` field order |
| `CMarkingData::Write` (`FUN_1015e4b0`) | `MARKING_DATA` field order |
| contour writer (`FUN_10157380`) | `NO_OF_SEG` / `START_U` / `START_V` placement |

Numeric widths came from the same literals: `%.2f` for areas, lengths and totals;
`%.5f` for coordinates and amplitudes; integers elsewhere.

The writer also makes `Mirror()` useful. Before it, the package could flip a part
in memory and had no way to emit the result.

### E.2 Bevel layout — my one wrong assumption, corrected

I had concluded from `nest.gen` that a contour carries exactly **one** bevel
(`BURNING_DATA` 919 = `BEVEL_DATA` 919). That holds for nests but not in general.
`31847/31847-P1NV42_1P-1P.gen` has **5 `BEVEL_DATA` blocks in one 10-segment
contour**:

```
START_OF_CONTOUR
BEVEL_DATA … END_OF_BEVEL_DATA     <- bevel of segment 0, before NO_OF_SEG
NO_OF_SEG=10
START_U / START_V
AMP_U … U / V                      <- segment 0
BEVEL_DATA … END_OF_BEVEL_DATA     <- bevel of segment 1
AMP_U … U / V                      <- segment 1
```

Bevels are **per-segment and sparse**; a segment inherits the last bevel seen
(`seg.Bevel = currentBevel` in `readContour`). The writer therefore emits a block
only when the bevel value changes, and hoists segment 0's bevel above `NO_OF_SEG`.
`sameBevel` does a struct comparison for that.

This is the same class of mistake as the earlier "one bevel per contour" reading
of `ENDCUT_TYPE`: a single sample generalised into a rule. Round-tripping the
corpus is what caught it.

### E.3 Round-trip result

Parse → `WriteDocument` → re-parse → deep compare, over every corpus file:

| Outcome | Count |
| --- | ---: |
| structurally identical | **45,744** |
| failed | **0** |
| skipped (read-only `LISTED_PROFILE`) | 6 |

"Structurally identical" means: header values, part count and per-part
name/posno/ÇOG/mirror/thickness/quality, per-part section counts, every segment's
AMP/RADIUS/SWEEP/end point, every sparse per-segment bevel, contour auxiliaries,
labels with their positions and glyph counts, symbols and edges all match.

Writing the output a second time is **idempotent** — byte-identical to the first
pass (`TestWriteRoundTrip`).

### E.4 What the writer does NOT claim

Byte equality with AVEVA output is **not** claimed: 0 of 45,744 files reproduce
byte-for-byte. Mean line-set similarity is **0.941** — the writer emits the same
`KEY=VALUE` lines about 94% of the time, with differences from:

* **Canonical field order.** The writer uses one fixed order per section; AVEVA
  varies it per writer. Any consumer of `.gen` must be order-independent, which
  the reader is.
* **`IDLE_DATA` and `BUMP_DATA` are not modelled.** The reader skips them, so a
  rewrite drops them. `IDLE_DATA` carries rapid/traverse contours that plate
  nesting does not need, but this makes the writer **not lossless** for those
  two sections. A consumer that needs them must extend the model first.
* **`LISTED_PROFILE` is read-only.** `WriteDocument` handles the plate family
  only; there is no `WriteProfiles`.
* **Unmodelled sections.** Anything captured as `RawSection` is written back
  verbatim, so it survives, but its `KEY=VALUE` order is preserved only within
  that section.

So: the writer produces files this package reads back identically, and which
AVEVA-side consumers should accept, but it is not a byte-exact clone of the
original writer.

### E.5 The writer paid for itself immediately — two real bugs

Round-tripping `Mirror()` through the writer exposed two defects that the
reader-only test suite could not see, because neither is observable without
materialising a mutated part.

**1. `Mirror()` mutated copies.** `Part`'s payload slices hold values
(`[]Marking`, `[]LabelText`, `[]LabelSymbol`, `[]Edge`, `[]StringData`), so
`for _, m := range p.MarkingData { m.MarkingSide = "OS" }` changed a copy and was
discarded. Marking side, label positions, symbol positions and edge endpoints
never actually mirrored. Only `Contour` pointers appeared to work, which is
exactly what made the bug invisible. All loops now index the slice.

**2. Segment points were aliased.** `readContour` set

```go
seg.Start = contour.Segments[len(contour.Segments)-1].End
```

so segment *i*'s `Start` **was** segment *i−1*'s `End` — one `*Point` shared by
two segments. `Contour.invert` negates every segment's `Start` and `End`, so a
shared point was negated twice and cancelled out: only the first `Start` and the
last `End` actually flipped. Measured before the fix, on
`31847-P1NV42_1P-1P.gen`: `Sweep` flipped but `End.X` stayed `13298`.

That aliasing was also a public footgun — mutating one segment silently changed
its neighbour. Each segment now owns its points.

After both fixes the same part mirrors correctly (`End.X` `13298` → `-13298`, and
the written file contains `U=-13298.00000`), and the corpus round-trip stays at
45,744/0. Pinned by `TestMirrorMutatesAndRoundTrips`, which asserts both the
non-aliasing and the value-slice mutation.

---

## Appendix F — IDLE/BUMP, explicit mirroring, and a profile writer

### F.1 `IDLE_DATA` and `BUMP_DATA` now modelled

Both were skipped by the reader, so a rewrite dropped them.

| Section | Corpus | Structure |
| --- | ---: | --- |
| `IDLE_DATA` | **93,571** in 769 files | `IDLE_DATA` → contour → `END_OF_IDLE_DATA` |
| `BUMP_DATA` | 2,974 (one per file) | `BUMP_DATA` → `SHAPE` → contour → `END_OF_BUMP_DATA` |

`IDLE_DATA` is the rapid/traverse path between cuts; `BUMP_DATA` is the plate
edge's bump profile. Both are payload siblings: `IDLE_DATA` sits between
`PART_INFORMATION` and `MARKING_DATA`, `BUMP_DATA` after `STRING_DATA`.

**`BUMP_DATA` is emitted by a component that is not in this installation.** The
string is absent from all three DLLs (`strings -a | grep BUMP` finds only a `BUMP`
UI label in `mar.dll`). Its schema is therefore sample-derived only.

Ordering source: the corpus, not decompilation. REA had **no usable provider**
this session — Ghidra 12.2 is installed but REA verifies 12.1.x, and
Hopper/IDA/Kuna are unconfigured:

```
ghidra: unsupported_version (12.2; accepted 12.1.x)
hopper: executable_missing (/opt/hopper/bin/Hopper)
ida:    not_configured (REA_IDA_MCP_CONFIG)
kuna:   not_configured (REA_KUNA_COMMAND)
```

For a *writer*, corpus order is arguably stronger evidence than decompiled order
anyway: it is the order shipped files actually contain.

### F.2 The hidden auto-mirror, now an explicit flag

`-mirror` initially had no effect. The cause was pre-existing and undocumented:

```go
// draw_plate.go
if osLength > tsLength {
    p.Mirror()
}
```

The drawer mirrored parts by itself based on marking-side length. Pre-mirroring a
part swapped TS↔OS, which made the drawer decide *not* to mirror — the two
cancelled exactly, byte for byte.

Mirroring is now explicit via `-mirror=auto|on|off`:

| Mode | Behaviour |
| --- | --- |
| `auto` (default) | historical heuristic: flip when OS marking is longer |
| `on` | force-mirror every part; the drawer's heuristic is disabled |
| `off` | never mirror, including the heuristic |

`Drawer.SetAutoMirror` lets the caller take responsibility. Default behaviour is
unchanged, so existing output is unaffected.

### F.3 `LISTED_PROFILE` is now writable

`ReadProfileDocument` / `WriteProfiles` mirror the plate-side
`ReadDocument` / `WriteDocument`. Key order is taken from `testdata/profiles.gen`
in the order it emits them, key by key.

Newly modelled so nothing is dropped on write:

| Section | Added |
| --- | --- |
| `COMMON_DATA` | 33 fields, and the **cross-section contour** (previously skipped) |
| `PROFILE_DATA` | `CUST_DATA_1..4`, `CUST_TEXT_1..10`, `PLANNING_UNIT`, `WORKING_LOCATION`, `BUILDING_STRATEGY` |
| `LEFT_END`/`RIGHT_END` | `ENDCUT_USER_DEFINED_PARAM1..20` and `DATA1..20` (was `1..7` and a single `DATA`), now arrays |

`ENDCUT_MASK` stays a **string**: observed values are bit patterns such as
`00010110000`, and leading zeros would not survive an integer round-trip.

Round-trip: **413 profiles across the 6 corpus files, 0 mismatches**, plus 440 in
`testdata/profiles.gen`. Verified on name, posno, MLENGTH/TLENGTH, quality, ship
no, surface treatment, mirror, and both end cuts (type, mask, code, A/R2/V1).

### F.4 Fidelity, honestly

Plate mean line-set similarity improved **0.941 → 0.960** by modelling
`IDLE_DATA`/`BUMP_DATA` and fixing three real bugs found while diffing against
originals:

| Field | Shipped | Was written as |
| --- | --- | --- |
| `STARTING_POS_X/Y` | `0.00` | `0.00000` |
| `LENGTH_EXCESS_1/2/3` | `1040.00000` | `1040.00` |
| `TRANSFORMATION_DATA` | present | **dropped entirely** (now modelled) |

0 files remain byte-identical, and that ceiling is structural rather than a
defect: the corpus contains **multiple writer dialects** that disagree on both
field presence and empty-vs-zero. In the same corpus, one writer emits
`REST_LENGTH=`, `TOOL_ID=`, `NO_OF_BLASTING_STARTS=` as empty while another emits
`REST_LENGTH=0.00`, `TOOL_ID=0`, `NO_OF_BLASTING_STARTS=0`, and one omits
`DISTANCE_Y1_Y2`/`CUTTING_GAS` that the other includes. A single canonical writer
cannot match both. Round-trip structural equality (45,744/0) is the contract that
is actually guaranteed.

---

## Appendix G — what the test suite found

Writing tests for the finished package turned up six bugs that the corpus
round-trip could not see, because it only compared *modelled* fields and never
asserted on the ones that were silently empty.

### G.1 `_SB` fields clobbered their primary (data loss on every part)

The reader had:

```go
case "PARTNAME_LONG", "PARTNAME_LONG_SB":
    part.PartnameLong = value
```

A real `PART_DATA` emits `PARTNAME_LONG=<real value>` and then
`PARTNAME_LONG_SB=` (empty). Both keys wrote the **same** field, so the empty
`_SB` line overwrote the real value. The same applied to `ASSEMBLY_LOW` and
`ASSEMBLY_HIGH`.

Consequence across the 45,750-file corpus, before → after:

| Field | populated before | after |
| --- | ---: | ---: |
| `PartnameLong` | **0** | 53,888 (70.3%) |
| `AssemblyLow` | **0** | 37,933 (49.5%) |
| `AssemblyHigh` | **0** | 37,933 (49.5%) |

This was pre-existing, not introduced by the writer work. No earlier test
asserted on those fields, and the round-trip harness only compared fields it knew
to be non-empty, so it stayed invisible. It surfaced only when a `DeepEqual` on a
whole `Part` compared them after a write/read cycle.

### G.2 A malformed plate document was reported as another family

`ReadDocument` returned `ErrNotPlatePart` for *any* `readGeneralData` error, so a
corrupt plate file was indistinguishable from a listed-profile file. The CLI then
skipped it as "not our input" and the run looked clean.

Now only a missing `GENERAL_DATA` marker yields `ErrNotPlatePart`; a parse failure
inside the header propagates as itself.

### G.3 `Process` panicked instead of returning parse errors

```go
go func() {
    if err := g.Wait(); err != nil {
        panic(err)   // any parse error crashed the tool
    }
    close(c)
}()
```

One bad `.gen` in a directory crashed the binary. The error is now captured before
`close(c)` and returned; the channel close orders the two accesses, so no lock is
needed (verified with `-race`).

### G.4 Four panics in the DXF drawer

The drawer is a library call on public types, so absent data must not crash it:

| Site | Trigger |
| --- | --- |
| `addPosNo` | `Part.PartCog == nil` |
| `addStringData` | `StringData.Pos == nil` |
| `drawContour` | nil or empty contour |
| `getMarkingTextPosition` | empty `Segments` |
| `getMarkingTextPosition` | **negative index** `Segments[i-1]` when `i == 0` |
| `drawMarkingText` | `INCLINATION_ANGLE_LINE` with empty contour |

The negative index is the interesting one: `i-1` was reachable on the first
segment, which no fixture happens to exercise.

### G.5 Writer bugs found by whole-structure round-trip

Requiring `reflect.DeepEqual` on a re-parsed document — rather than spot-checking
a few fields — found three writer defects in the LISTED_PROFILE path:

1. **`NO_OF_PROFS` never written.** Its Go type is `uint64`, and the code
   generator that produced the writer only handled `string`/`int`/`float64`, so it
   was skipped silently.
2. **20 bevel keys never written.** `BEVEL_CODE`, `BEVEL_NAME`, `BEVEL_TYPE`,
   `BEVEL_VARIANT`, `E`, `GAP`, `CHAMFER`, `ALPHA`, `BETA`, `NOSE`, `H`, `H_FACT`,
   `ANGLE_TS/OS`, `DEPTH_TS/OS`, `CHAMFER_*` appear only on profiles that carry a
   bevel. The key order had been sampled from the **first** `PROFILE_DATA` block
   only; the union across all 440 profiles is 92 keys, not 72.
3. **Float precision loss.** `DENSITY=0.00000785` was written with `%.5f` →
   `0.00001`, and `AREA=2578.6708327727` → `2578.67083`. LISTED_PROFILE floats now
   use the shortest round-tripping representation.

Both families now reach full `reflect.DeepEqual` equality on a write/read cycle:
plate 394/394 parts and 440/440 profiles, `COMMON_DATA` included.

### G.6 Test suite shape

| Area | Coverage |
| --- | --- |
| Fixtures | parse + `DeepEqual` round-trip + writer idempotency + mirror round-trip |
| Fuzzing | `FuzzReadDocument`, `FuzzReadProfileDocument` — never panic, and anything accepted must be writable and re-readable |
| Helpers | `parseOptFloat`/`parseOptInt`/`cleanStr`/`formatFloat` tables incl. `-`, empty, garbage |
| Geometry | angle conversions, midpoints, arc midpoint on-circle, segment length/bulge, empty-contour helpers |
| Document type | profile first-line, plate header, empty type, absent, scan-window limit |
| CLI | all three mirror modes, family skipping, malformed input, DXF structure |
| Drawer | DXF structure, `SetAutoMirror` on/off/default, nil and empty inputs |
| Benchmarks | read/write both families, parallel read |

Coverage **62.6% → 72.4%**; `go test -race ./...` clean.

```bash
go test ./...                          # 152 cases
go test -race ./...
go test -fuzz=FuzzReadDocument -fuzztime=60s .
go test -bench=. -benchmem ./...
```

---

## Appendix H — anonymized fixtures

`testdata/*.gen` are real AVEVA output and originally carried shipyard identifiers:
hull numbers (`SHIP_NO=76181`), nest names (`45-169816-03250-M72`), part and assembly
codes (`549-10123-B69432-CT_RIP_729898-094P`), block codes and drawing numbers. They were
rewritten once by a one-shot tool that has since been deleted.

### H.1 Method: shape-preserving pseudonyms

The transformation was performed by a one-shot tool (now removed). The method is
recorded here because that is what matters for future data ingestions.

Each character is remapped by class — digit to digit, letter to letter — while
separators (`- _ / . > space`) are preserved byte for byte. The result therefore has
**exactly the same length and the same token alignment** as the input, which gives
three properties that matter:

1. **References survive.** `PART_DATA NAME=` and `PART_INFORMATION PART_NAME=` map to
   the same pseudonym, so part-to-payload binding still works and the fixtures keep
   1 / 394 / 440 parts.
2. **Byte length is unchanged.** Verified: 779,328 / 1,327,096 / 4,276,033 bytes before
   and after, on all three fixtures. Holds on a second pass too.
3. **Embedded identifiers can be scrubbed.** Because separators stay aligned, tokens
   inside free text (`STRING`, `TEXT`) can be replaced from a dictionary derived from
   the identifier fields.

The mapping is deterministic (sorted originals, salted FNV), **injective**, and never
emits an original value, so two distinct parts cannot collapse into one name.

### H.2 What is deliberately left alone

`STRING` and `TEXT` are rewritten token by token, and only tokens of three or more
characters that appear in an identifier field. Specification values are untouched:
`180PP A40`, `AFT`, `FOR`, `CL`, `PS`, `SB`, `UP`, angles such as `+50.00`, material
grades, and every geometric coordinate. Two-character terms such as `UZ` (Unterzug)
are domain terminology, not customer data, and are preserved.

Numeric fields outside the identifier set are never touched, so `POSNO`, `PART_ID`,
`TOOL_ID` and all coordinates keep their values. Dates (`SPLIT_DATE`,
`PRODUCTION_DATE`, `VERIFICATION_DATE`) are remapped into 2020 so no production
timeline is disclosed.

### H.3 Verification

| Check | Result |
| --- | --- |
| identifier values still identical to original | **0 / 2,406** |
| high-risk literals (`10123`, `729898`, `CT_RIP`, …) | **0** |
| byte length preserved | all three fixtures |
| line count and key sequence | identical |
| part / profile counts | 1 / 394 / 440 preserved |
| round-trip `DeepEqual` | passes |
| `go test ./...`, `-race` | clean |

### H.4 History was rewritten

The originals were purged from **all** history, not just the working tree. Two
separate rewrites (`git filter-repo`) were required:

1. `testdata/` was removed from every commit on every ref. Besides the three
   fixture files, history contained **`testdata/archive.zip`, 4.65 MB holding
   several hundred real production part files** (`<block>-<partname>_S-1S.gen`, ...),
   added in an early commit and deleted from the working tree later but still
   reachable. The anonymized fixtures were then re-added as a new commit.
2. A 3 MB compiled binary (`gen-anonymize`, accidentally committed build output)
   was purged, and `.gitignore` now covers build products.

Result, measured over every blob reachable from every ref:

| Check | Result |
| --- | --- |
| original fixture blob ids | **absent** |
| original `archive.zip` blob | **absent** |
| zip-magic blobs anywhere in history | **0** |
| `.gen` blobs in history | 3, all anonymized |
| sensitive identifiers across 136 blobs | **0** |
| repository size | 60 MB → **143 KiB** |
| commit count | 27 → 29 |

One historical `plate_test.go` still contains the literal path string
`testdata/archive.zip` inside a `zip.OpenReader` call. That is a dead code
reference, not data: the archive is gone and the current test file no longer
references it. Old commits are not runnable (testdata was purged from them),
which is the accepted cost of removing the data.

**Force-pushing the branches was necessary but not sufficient.** Verified against
the remote after the push:

| Remote ref | Rewritten? | Serves original data? |
| --- | --- | --- |
| `refs/heads/main`, `feat-*` | yes | **no** — clean |
| `refs/pull/1,3,4,5,6,7,9,10` | no | **yes** — all 3 original `.gen` blobs **and** `archive.zip` |
| `refs/pull/8` | no | **yes** — all 3 original `.gen` blobs |

Measured by fetching each ref into a scratch repository and testing for the
original blob ids directly:

```
pull/1  original-gen-blobs=3/3  archive.zip=YES
pull/3  original-gen-blobs=3/3  archive.zip=YES
...
pull/8  original-gen-blobs=3/3  archive.zip=absent
pull/10 original-gen-blobs=3/3  archive.zip=YES
```

`refs/pull/*` are read-only pseudo-refs that a client cannot delete or rewrite;
GitHub owns them. Until they are gone, the customer data is still downloadable
from the public remote by anyone who knows the PR number.

### What actually removes it

1. **GitHub Support** — ask them to purge unreachable objects *and* delete the
   `refs/pull/*` refs for this repository. This is the targeted fix; GitHub has
   done exactly this for leaked-data incidents.
2. **Delete and recreate the repository** — the only fully self-service path.
   Push the clean history to a NEW repository, then delete the old one. This
   guarantees no PR refs, no unreachable objects and no cached PR diff views.
   Star/fork/issue history is lost unless exported first.
3. **Forks** — any existing fork independently retains the data and must be
   deleted or contacted; a fork is a full copy, not a pointer.
4. **PR diff views** can remain visible in the UI even after refs are pruned,
   until GitHub's caches expire.

Verify from outside at any point with:

```bash
git ls-remote origin 'refs/pull/*'          # should print nothing
```

### A stale remote-tracking ref held the leaked content

One rewrite was not enough. `refs/remotes/origin/feat-add-plate2dxf` still pointed
at a pre-scrub commit, so the original fixture blobs stayed reachable and
`git rev-list --objects --all` (which includes remote-tracking refs) still found
them. `filter-repo` removed that ref; a final `--replace-text` pass scrubbed the
identifiers and the leftover blob hashes from every historical revision of these
notes.

**Final audit** — every blob reachable from every ref:

| Check | Result |
| --- | --- |
| blobs scanned | 138 |
| zip-magic blobs | **0** |
| `.gen` blobs | 3, all anonymized |
| sensitive identifiers or original blob hashes | **none** |
| refs | 3 branches + stash, no remote-tracking refs |

### Local state

The local repository is clean: 3 `.gen` blobs, all anonymized, hashes matching
the working tree; zero zip-magic blobs; zero sensitive identifiers across 136
blobs; 60 MB → 143 KiB.

The pre-rewrite backup at `aveva-go-PRE-REWRITE-<timestamp>` (60 MB) still holds
the originals. **Delete it once the remote is confirmed clean** — otherwise the
purge is defeated on this machine.
