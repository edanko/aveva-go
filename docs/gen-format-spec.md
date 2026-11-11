# AVEVA Marine `*.gen` — Generic File Format Specification

Reverse-engineered from the shipped **AVEVA Marine 12.1 SP5** installation
(`OH12.1.SP5/`) and cross-checked against the sample files shipped in
`testdata/`.

### Artifacts

There is **no single component** that handles both document families. Ownership
was determined by presence-testing the format's discriminating keys
(`PART_DATA`/`BURNING_DATA` for the plate family, `MLENGTH`/`LEFT_END`/
`PROFILE_DATA` for the listed-profile family) across every binary in the
installation.

| Component | Size | Plate family | Profile family (`.gen`) | Role |
| --- | ---: | :---: | :---: | --- |
| `mar.dll` | 64.6 MB | ✅ | ✅ | **The real owner of both.** Aveva Marine broker library — every Marine module links it. Emission literals recovered in place (§6). |
| `mar_p.dll` | 64.6 MB | ✅ | ✅ | Same as `mar.dll`; identical size and section layout, sibling build (ANSI/MBCS variant). |
| `TbGenPostProcessor.dll` | 2.7 MB | ✅ | ❌ | **Plate-only.** The NC post processor. Primary artifact analysed here. |
| `FEModellerDll.dll`, `inithull.dll`, `marHullVerify.dll`, `sj912.dll`, `TbBracketWizard.dll` | 33–51 MB | ✅ | ❌ | Marine feature DLLs that embed the plate reader. |
| `fabricationcore.dll` | 3.1 MB | ❌ | ⚠️ | Contains the token `LISTED_PROFILE` **once**, in an *XML* exporter fixture (`generic_data`, lowercase `type_of_generic_file`, `Test xml document for profile generic xml exporter`). Not the `.gen` text parser. |
| `TbGeneric2Plate.exe`, `tbgenpreview.dll/.exe` | 24–28 KB | — | — | .NET thin wrappers; no format logic. |

Presence test result for `TbGenPostProcessor.dll` — **all absent**:
`LISTED_PROFILE`, `PROFILE_DATA`, `START_OF_FCONTOUR`, `HOLES_NOTCHES_CUTOUTS`,
`MLENGTH`, `TLENGTH`, `LEFT_END`, `RIGHT_END`, `PROF_SIDE`, `SIDE_UP`,
`DIRECTION_TRACE`, `BEVEL_TRACE`, `GENERATED_AS`, `NO_OF_PROFS`, `ENDCUT_MASK`,
`LEFT_ORIGIN`. It is exclusively a plate-family reader.

`TbGenPostProcessor.dll` identity: SHA-256
`a886adb9c5c936d63414d770cb51224d4db29a39d4265cdd96de2dc9a910b525`,
PE32 x86, MSVC C++ with PDB `TbGenPostProcessor.pdb`, analysed with Ghidra
12.1.2 via REA 6.8.0. RTTI classes recovered include `PinGpp_CGenericBase`,
`PinGpp_CGenericProperties`, `PinGpp_CPartData`, `PinGpp_CGeneralData`,
`PinGpp_CMarkingData` and `PinGpp_CContour`.

> **Provenance note.** Claims marked *(observed)* are direct reads of shipped
> artifact bytes, decompiled code, or the shipped sample files. Claims marked
> *(inferred)* follow from those observations but were not proven by a single
> instruction. Claims marked *(unknown)* are explicit gaps.
>
> **Attribution.** The per-part schema (§3.2–§3.8, §4, §5) is evidenced by
> decompiled `TbGenPostProcessor.dll` code __and__ by the samples. The listed
> profile family (§6) is evidenced by `testdata/profiles.gen` **and by the
> emission literals recovered in place from `mar.dll`** (§6 preamble) — it is
> **not** implemented in `TbGenPostProcessor.dll`. §6.2 (sibling order) and §6.5
> (`-` sentinel) are now confirmed against the implementation, not inferred from
> the sample.
>
> **Method note.** `mar.dll` was not opened in Ghidra (51.7 MB of `.text`). Its
> emission order was recovered by locating each `KEY=` literal in `.rdata` and
> reading the literals in file order, which for MSVC is first-reference order —
> the same method used to derive the plate schema. This yields field **order and
> spelling** reliably; it does not yield per-field type information beyond what
> the literal and the samples show.

---

## 1. What the format is

`*.gen` ("generic file") is AVEVA Marine's **manufacturing-neutral intermediate
format** for plate parts and profile parts. A downstream **generic post
processor** (`TbGenPostProcessor.dll`, CESSI/ESSI writers, etc.) consumes a
`.gen` file and emits machine-specific NC / tape code.

The format is a **line-oriented, `KEY=VALUE` text format with bare section
markers**. There is no header, no length prefix, no binary framing, and no
compression. It is designed to be streamed in a single forward pass.

Two primary document families share the same lexical rules but different
top-level schemas, and — importantly — **different implementing components**:

1. **Plate nest / plate part** — `TYPE_OF_GENERIC_FILE=PLATE_PART`
   (a nest of many parts) or an empty value (a single-part export).
   Implemented by the plate reader inside `mar.dll` and consumed by
   `TbGenPostProcessor.dll` (which is plate-only).
2. **Listed profile** — `TYPE_OF_GENERIC_FILE=LISTED_PROFILE`
   (a batch of profile parts with end-cut and end-profile data).
   Implemented in `mar.dll`/`mar_p.dll`; **no** support in
   `TbGenPostProcessor.dll`.

Consequently the two families cannot be validated by testing one binary. A
mapping of which component owns which family is in the Artifacts table above.

---

## 2. Lexical grammar

```
file        := line*
line        := comment? | marker | pair | blank
marker      := UPPER_SNAKE_TOKEN
pair        := key '=' value
key         := UPPER_SNAKE_TOKEN
value       := <rest of line, verbatim>
comment     := ';' <anything>        (inferred — never seen in samples)
```

Rules *(observed)*:

* **Encoding** is ASCII / single-byte. No BOM, no UTF-8 multibyte sequences in
  the samples.
* **Line endings are CRLF** in every shipped sample. A parser must accept both
  CRLF and LF. Go's `bufio.Scanner` with the default `ScanLines` split already
  strips a trailing `\r`, so this is handled transparently.
* **`=` is the only key/value separator.** Split on the *first* `=`. Values may
  legally contain `=` and spaces (`IDENT_STRING=123-…-PS>UP.`).
* **Values are not quoted and not escaped.** Leading/trailing spaces inside a
  value are significant only in that AVEVA pads some fields with spaces
  (e.g. `PRODUCTION_DATE=          `, `PLATE_ALIGNMENT=  `). A robust reader
  should treat an all-spaces value as empty *(inferred)*.
* **No trailing whitespace** is emitted after a value in the samples.
* **Order is significant only inside contour/segment runs.** Field order within
  a section is conventional but the reader dispatches on the key, so a tolerant
  reader may treat field order as free. The segment run *is* order-dependent in
  the sense that the last segment's `U`/`V` closes a segment.
* **Section markers carry no `=`.** `NO_OF_SEG=9` is a pair; `START_OF_CONTOUR`
  is a marker. A reader distinguishes them by the presence of `=`.

### 2.1 The `-` sentinel *(important)*

Numeric fields may contain a **single hyphen `-`** meaning *"not applicable /
undefined for this end cut"*. Observed in `LISTED_PROFILE` end-cut data:

```
A=-
B=-
C=-
R2=-
V3=-
V4=-
```

`strconv.ParseFloat("-", 64)` **fails**. Any parser that hard-fails on `-` will
reject every real `LISTED_PROFILE` file at the first `LEFT_END` block.
*(observed: `testdata/profiles.gen`, keys `A,B,C,R2,V3,V4`)*

Recommended treatment: map `-` (and the empty string) to a `*float64` nil, or to
`0` with a separate "defined" flag.

---

## 3. Complete key vocabulary

Recovered from the format's token mapper — the function that turns a key line
into an integer token id, then dispatches on it. The token ranges are grouped by
section, which is how the reader knows the current context.

### 3.1 Section markers

| Token | Marker | Payload |
| --- | --- | --- |
| `0x65` | `GENERAL_DATA` | `CGeneralData` |
| `0x66` | `PART_DATA` | `CPartData` |
| `0x67` | `MARKING_DATA` | `CMarkingData` |
| `0x68` | `GSD_DATA` | `CGsdData` |
| `0x69` | `MARKING_SHAPE_DATA` | `CMarkingShapeData` |
| `0x6a` | `LABELTEXT_DATA` | `CLabelText` |
| `0x6b` | `LABELSYMBOL_DATA` | `CLabelSymbol` |
| `0x6c` | `BURNING_DATA` | `CBurningData` |
| `0x6d` | `IDLE_DATA` | `CIdleData` |
| `0x6e` | `BLASTING_DATA` | `CBlastingData` |
| `0x6f` | `AUXILIARY_DATA` | `CAuxiliaryData` |
| `0x70` | `PART_INFORMATION` | `CPartInformation` |
| `0x71` | `GEOMETRY_DATA` | `CGeometryData` |
| `0x72` | `STRING_DATA` | `CStringData` |
| `0x78` | `BLASTING_MARKING_DATA` | — |
| `0x79` | `END_OF_BLASTING_MARKING_DATA` | — |

Note `0x73`/`0x74` are the literal tokens `2AXIS` / `3AXIS`, recognised as bare
lines because they appear in `TYPE_OF_GENERIC_FILE=PLATE_PART_CHECK` template
text.

### 3.2 `GENERAL_DATA` — document header

Terminator: `END_OF_GENERAL_DATA`.

| Token | Key | Type | Canonical write order |
| --- | --- | --- | --- |
| `0xe7` | `TYPE_OF_GENERIC_FILE` | string | 1 |
| `0xcb` | `TYPE_OF_MANUFACT` | string | 2 |
| `0xc9` | `BURNER_ID` | string | 3 |
| `0xca` | `NEST_NAME` | string | 4 |
| `0xcc` | `RAW_NAME` | string | 5 |
| `0xcd` | `RAW_LENGTH` | float | 6 |
| `0xce` | `RAW_WIDTH` | float | 7 |
| `0xcf` | `RAW_THICKNESS` | float | 8 |
| `0xd0` | `RAW_AREA` | float | 9 |
| `0xd1` | `QUALITY` | string | 10 |
| `0xd2` | `DENSITY` | float | 11 |
| `0xd3` | `SURF_TREAT_TS` | string | 12 |
| `0xd4` | `SURF_TREAT_OS` | string | 13 |
| `0xd5` | `REST_LENGTH` | float | 14 |
| `0xd6` | `PRODUCTION_DATE` | string (`YYYY-MM-DD` or spaces) | 15 |
| `0xd7` | `NO_OF_PARTS` | int | 16 |
| `0xd8` | `TOTAL_IDLE` | float | 17 |
| `0xd9` | `TOTAL_MARKING` | float | 18 |
| `0xda` | `TOTAL_BURNING` | float | 19 |
| `0xdb` | `TOTAL_BLASTING` | float | 20 |
| `0xdc` | `TOTAL_NO_OF_STARTS` | int | 21 |
| `0xdd` | `NO_OF_BURNING_STARTS` | int | 22 |
| `0xde` | `NO_OF_BLASTING_STARTS` | int | 23 |
| `0xdf` | `NO_OF_MARKING_STARTS` | int | 24 |
| `0xe0` | `STARTING_POS_X` | float | 25 |
| `0xe1` | `STARTING_POS_Y` | float | 26 |
| `0xe3` | `QUANTITY_NORMAL` | int | 27 |
| `0xe4` | `QUANTITY_MIRRORED` | int | 28 |
| `0xe5` | `TOOL_ID` | int | 29 |
| `0xe6` | `PROCESS_ID` | string | 30 |
| `0xe8` | `SEQUENCE` | string | 31 |
| `0xe2` | `DISTANCE_Y1_Y2` | float | 32 |
| `0xea` | `CUTTING_GAS` | int | 33 |
| `0xeb` | `MATERIAL_CODE` | string | 34 |
| `0xec` | `PURCHASE_INFO` | string | 35 |
| `0xed` | `PLATE_ALIGNMENT` | string | 36 |
| `0xee` | `SPLIT_DATE` | string | 37 |
| `0xef` | `TOTAL_MARK_BLAST` | float | 38 |
| `0xf0` | `NO_OF_MARK_BLAST_STARTS` | int | 39 |
| `0xf1` | `VERIFICATION_DATE` | string | 40 |
| `0xe9` | `END_OF_GENERAL_DATA` | marker | 41 |

Write order is the order the shipped `CGeneralData` serialiser emits
*(observed: `FUN_1015a4b0` in `TbGenPostProcessor.dll`)*.

Sample header (`testdata/nest.gen`):

```
GENERAL_DATA
TYPE_OF_GENERIC_FILE=PLATE_PART
TYPE_OF_MANUFACT=2AXIS
BURNER_ID=C
NEST_NAME=45-169816-03250-M72
RAW_NAME=
RAW_LENGTH=19500.00
RAW_WIDTH=3200.00
RAW_THICKNESS=11.00
RAW_AREA=62.40
QUALITY=A40
DENSITY=7850.00
...
END_OF_GENERAL_DATA
```

### 3.3 `PART_INFORMATION` — part→section association

Terminator: `END_OF_PART_INFORMATION`.

| Token | Key | Type |
| --- | --- | --- |
| `0x12d` | `PART_NAME` **or** `NAME` | string |
| `0x12e` | `TYPE_OF_WORK` | string (`BURNING`, …) |
| `0x12f` | `END_OF_PART_INFORMATION` | marker |

`PART_INFORMATION` is the **scope switch**. Everything emitted after it
(`BURNING_DATA`, `MARKING_DATA`, `GEOMETRY_DATA`, `STRING_DATA`, `IDLE_DATA`,
…) belongs to the part named by its `PART_NAME`/`NAME`. In a nest file the
`PART_DATA` blocks all come first, then the per-part payload blocks reference
them by name.

**The relation is not 1:1.** Measured on `nest.gen`: 394 `PART_DATA`, 395
`PART_INFORMATION`, 392 distinct names. Two independent multiplicities occur:

* a part may be **placed** several times — one `PART_DATA` per nest position
  (`"54949-DA4137_36X-3ZK"` appears 3× in both blocks);
* a part may be **opened** by several scopes — more `PART_INFORMATION` than
  `PART_DATA` (`"54949-QW197_3X-21ZK"`: 1 `PART_DATA`, 2 `PART_INFORMATION`).

A reader must therefore associate scopes with placements positionally and let
excess scopes accumulate onto the last placement of that name. Keying parts by
name alone loses geometry (see findings Bug 5).

Sample:

```
PART_INFORMATION
PART_NAME=54949-QW183_4X-6ZK
TYPE_OF_WORK=BURNING
END_OF_PART_INFORMATION
```

### 3.4 `PART_DATA` — part attributes

Terminator: `END_OF_PART_DATA`.

| Token | Key | Type | Canonical write order |
| --- | --- | --- | --- |
| — | `NAME` | string | 1 |
| `0x12e` | `TYPE_OF_WORK` | string | 2 |
| `0x191` | `POSNO` | string | 3 |
| `0x192` | `PARTNAME_LONG` | string | 4 |
| `0x1a5` | `PARTNAME_LONG_SB` | string | (variant) |
| `0x193` | `SHIP_NO` | string | 5 |
| `0x194` | `ASSEMBLY_LOW` | string | 6 |
| `0x1a6` | `ASSEMBLY_LOW_SB` | string | (variant) |
| `0x195` | `ASSEMBLY_HIGH` | string | 7 |
| `0x1a7` | `ASSEMBLY_HIGH_SB` | string | (variant) |
| `0x196` | `BLOCK_NO` | string | 8 |
| `0x197` | `PART_AREA` | float `%.2f` | 9 |
| `0x198` | `MIRRORED` | int `0`/`1` | 10 |
| `0x1be` | `CLIP_GRINDING_SIDE` | string | 11 |
| `0x199..0x19c` | `CUST_DATA_1..4` | string | 12–15 |
| `0x1b4..0x1bd` | `CUST_TEXT_1..10` | string | 16–25 |
| `0x19d` | `WORKING_LOCATION` | string | 26 |
| `0x19e` | `BUILDING_STRATEGY` | string | 27 |
| `0x19f` | `PLANNING_UNIT` | string | 28 |
| `0x1a0` | `EXTENSION_U` | float `%.5f` | 29 |
| `0x1a1` | `EXTENSION_V` | float `%.5f` | 30 |
| `0x1a2` | `PLATE_SIDE` | string | 31 |
| `0x1a3` | `PANEL_NAME` | string | 32 |
| `0x1a8` | `PART_COG_U` | float `%.5f` | 33 |
| `0x1a9` | `PART_COG_V` | float `%.5f` | 34 |
| `0x1ab` | `NESTED_ON` | string | 35 |
| `0x1aa` | `KNUCKLED` | string `NO`/… | 36 |
| `0xcc` | `RAW_NAME` | string | 37 |
| `0x1ad..0x1af` | `NO_INTERVALS_EXCESS_1..3` | int | 38–40 |
| `0x1b0..0x1b2` | `LENGTH_EXCESS_1..3` | float | 41–43 |
| `0x1bf` | `COMMENT` | string | — |
| `0x1c0` | `FUNCTIONAL_DESCR_CODE` | string | — |
| `0x1c1` | `FUNCTIONAL_DESCR_TEXT` | string | — |
| `0x1b3` | `DESTINATION` | string | — |
| `0x1fe` | `PART_ID` | int | — |
| `0x1a4` | `END_OF_PART_DATA` | marker | — |

Order in the table reflects the *interleaved* real writer order observed in the
samples (which differs slightly from the reduced `CPartData::Write` in the post
processor). The post-processor's own serialiser
*(observed: `FUN_1015fbc0`)* emits this reduced sequence, confirming the field
set and types but omitting the `_SB` variants, `PART_COG_*`, `COMMENT`,
`DESTINATION`, `PART_ID` and `CUST_TEXT_*`:

```
PART_DATA / NAME= / TYPE_OF_WORK= / POSNO= / PARTNAME_LONG= / SHIP_NO= /
ASSEMBLY_LOW= / ASSEMBLY_HIGH= / BLOCK_NO= / PART_AREA= / MIRRORED= /
CUST_DATA_1..4= / WORKING_LOCATION= / BUILDING_STRATEGY= / PLANNING_UNIT= /
EXTENSION_U= / EXTENSION_V= / PLATE_SIDE= / PANEL_NAME= / KNUCKLED= /
NESTED_ON= / RAW_NAME= / NO_INTERVALS_EXCESS_1..3= / LENGTH_EXCESS_1..3= /
END_OF_PART_DATA
```

### 3.5 `BURNING_DATA` — one cut path

Terminator: `END_OF_BURNING_DATA`.

| Token | Key | Type |
| --- | --- | --- |
| `0x1f5` | `SHAPE` | string (`OUTER_CONTOUR`, `HOLE`, …) |
| `0x1f6` | `DIRECTION` | int |
| `0x1f7` | `GEOMETRY_VALID_FOR` | int |
| `0x1f9` | `NUMBER_OF_HEADS` | int |
| `0x1fa` | `HOLE_SHIPSIDE` | string |
| `0x1fb` | `BEVEL_DEFINED` | string |
| `0x1fc` | `DESIGNATION` | string |
| `0x1fd` | `START_END_IN_GAP` | string |
| `0x1fe` | `PART_ID` | int |
| `0xe2` | `DISTANCE_Y1_Y2` | float `%.2f` |
| — | `START_OF_CONTOUR` … `END_OF_CONTOUR` | contour |
| — | `BEVEL_DATA` … `END_OF_BEVEL_DATA` | bevel, nested inside the contour |
| `0x1f8` | `END_OF_BURNING_DATA` | marker |

`BEVEL_DATA` appears **inside** `START_OF_CONTOUR … END_OF_CONTOUR`, immediately
after `START_OF_CONTOUR` and before the first segment *(observed in
`testdata/nest.gen`)*:

```
BURNING_DATA
SHAPE=OUTER_CONTOUR
DIRECTION=1
GEOMETRY_VALID_FOR=1
NUMBER_OF_HEADS=1
DISTANCE_Y1_Y2=0.00
START_OF_CONTOUR
BEVEL_DATA
BEVEL=NONE
...
END_OF_BEVEL_DATA
NO_OF_SEG=...
START_U=...
...
END_OF_CONTOUR
END_OF_BURNING_DATA
```

### 3.6 `MARKING_DATA` — one marking path

Terminator: `END_OF_MARKING_DATA`. Real writer order *(observed in
`testdata/part.gen`)*:

```
MARKING_DATA
MARKING_SIDE=TS
MARKING_TYPE=PANEL
MARKING_NAME=27456-P1QW128.5_S
MARKING_USER_TYPE=0
MARKING_ATTRI=
MARKING_ATTRI_SB=27456-V27456-PS1-405S
MARKING_SHIPSIDE=SB
MARKING_ASSEMBLY_LOW=
MARKING_PROFILE_THICKNESS=0.00
MARKING_WELD=0.00
MARKING_DIRECTION=1
NUMBER_OF_HEADS=1
GEOMETRY_VALID_FOR=1
DISTANCE_Y1_Y2=0.00
MTRL_SIDE_U=0.00000
MTRL_SIDE_V=0.00000
COMMENT=
INCLINATION_ANGLE=0.00
START_OF_CONTOUR
...
END_OF_CONTOUR
END_OF_MARKING_DATA
```

| Token | Key | Type |
| --- | --- | --- |
| `0x44d` | `MARKING_SIDE` | string (`TS`, `OS`) |
| `0x44e` | `MARKING_TYPE` | string |
| `0x44f` | `MARKING_NAME` | string |
| `0x45c` | `MARKING_USER_TYPE` | int |
| `0x450` | `MARKING_ATTRI` | string |
| `0x456` | `MARKING_ATTRI_SB` | string |
| `0x457` | `MARKING_SHIPSIDE` | string (`BOTH`, `SB`, `TS`) |
| `0x458` | `MARKING_ASSEMBLY_LOW` | string |
| `0x459` | `MARKING_PROFILE_THICKNESS` | float `%.2f` |
| `0x45b` | `MARKING_WELD` | float `%.2f` |
| `0x45a` | `MARKING_DIRECTION` | int |
| `0x1f9` | `NUMBER_OF_HEADS` | int |
| `0x1f7` | `GEOMETRY_VALID_FOR` | int |
| `0xe2` | `DISTANCE_Y1_Y2` | float `%.2f` |
| `0x453` | `MTRL_SIDE_U` | float `%.5f` |
| `0x454` | `MTRL_SIDE_V` | float `%.5f` |
| `0x1bf` | `COMMENT` | string |
| — | `INCLINATION_ANGLE` | float `%.2f` |
| `0x455` | `END_OF_MARKING_DATA` | marker |

The post-processor's reduced `CMarkingData` serialiser *(observed:
`FUN_1015e4b0`)* emits `MARKING_SIDE, MARKING_TYPE, MARKING_NAME,
MARKING_ATTRI, NUMBER_OF_HEADS, GEOMETRY_VALID_FOR, DISTANCE_Y1_Y2,
MTRL_SIDE_U, MTRL_SIDE_V` then the contour — i.e. the mature superset above is
what the brokerage writer produces.

### 3.7 Per-part payload sections

These are **siblings of `PART_DATA`**, scoped to the preceding
`PART_INFORMATION`. Field sets below are measured over the 45,750-file corpus,
not inferred.

| Section | Terminator | Fields | Corpus |
| --- | --- | --- | ---: |
| `GEOMETRY_DATA` | `END_OF_GEOMETRY_DATA` | `GEOMETRY_TYPE` (`OUTER_CONTOUR`, `HOLE`, `ROLL_AXIS`, `ROLL_AXIS_CROSS_SIGN`, `BEVEL_SYMBOL`, or empty), `ROLL_AXIS_NO`, vernacular alias `TYPE`, one `START_OF_CONTOUR…END_OF_CONTOUR` | — |
| `STRING_DATA` | `END_OF_STRING_DATA` | `STRING_TYPE`, `STRING_POSITION_U/V`, `STRING_ANGLE`, `STRING_HEIGHT`, `STRING` | — |
| `IDLE_DATA` | `END_OF_IDLE_DATA` | a contour (rapid/traverse path, no cut data) | 187 in one file |
| `LABELTEXT_DATA` | `END_OF_LABELTEXT_DATA` | `TEXT_TYPE` (space-padded), `TEXT_POSITION_U/V`, `TEXT_ANGLE`, `TEXT_HEIGHT`, `TEXT_DIRECTION`, `TEXT_SIDE` (`BOTH`/`TS`/`OS`), `TEXT`, `TEXT_ELEMENTS`, then `TEXT_ELEMENTS` × nested `GEOMETRY_DATA` glyph strokes | **4,362 files** |
| `LABELSYMBOL_DATA` | `END_OF_LABELSYMBOL_DATA` | `SYMBOL_TYPE` (space-padded), `SYMBOL_POSITION_U/V`, `SYMBOL_ANGLE`, `SYMBOL_HEIGHT`, `SYMBOL_FONT`, `SYMBOL_NUMBER` | **4,356 files** |
| `EDGE_DATA` | `END_OF_EDGE_DATA` | `NAME`, `START_U`, `START_V`, `END_U`, `END_V`, `ALENGTH`, `MLENGTH` | 124 files |
| `AUXILIARY_DATA` | `END_OF_AUXILIARY_DATA` | `AUXILIARY_FUNCTION`, `AUXILIARY_PARAMETER`. Emitted **inside a contour**, after the segments | 201 files, 16,122 blocks |
| `BLASTING_DATA` | `END_OF_BLASTING_DATA` | `SHAPE`, `DIRECTION`, `GEOMETRY_VALID_FOR`, `NUMBER_OF_HEADS`, `DISTANCE_Y1_Y2`, contour | **0 files** |
| `GSD_DATA` | `END_OF_GSD_DATA` | `VECT_CB_Y1_U/V`, `VECT_BA_Y1_U/V`, `VECT_CB_Y2_U/V`, `VECT_BA_Y2_U/V`, `GSD`, `GSD_DIST` | **0 files** |
| `MARKING_SHAPE_DATA` | — | marking stroke glyph data | **0 files** |
| `BLASTING_MARKING_DATA` | `END_OF_BLASTING_MARKING_DATA` | — | **0 files** |

`AUXILIARY_DATA` inside a contour is easy to miss:

```
START_OF_CONTOUR
NO_OF_SEG=1
... segments ...
AUXILIARY_DATA
AUXILIARY_FUNCTION=99
END_OF_AUXILIARY_DATA
END_OF_CONTOUR
```

**Space padding is real.** `TEXT_TYPE` and `SYMBOL_TYPE` are right-padded to 26
characters (`TEXT_TYPE=PANEL_NAME                `); trim before use.

### 3.8 Sections the shipped reader names but no sample exercises

`MARKING_LEFT_END`, `MARKING_RIGHT_END`, `REF_LINE_MARKING`, `BENDING_TABLE`,
`GENERAL_MARKING_DATA`, `CURVE_DATA`, `CURVE_3D_DATA`, `PENETRATING_PROFILE`,
`DRILLING`/`THREADING`, `SLOTTED` occur in **zero** of 45,750 corpus files.
Their field names are known only from the `mar.dll` emission literals
(§6.6–§6.9). Treat their schemas as **unvalidated**.

## 4. Contour and segment encoding

A contour is a **polyline of arc-or-line segments in the plate's (U, V) plane**.
It is the only place where the file encodes real geometry.

```
START_OF_CONTOUR
NO_OF_SEG=<int>            # advisory count; do not rely on it
START_U=<float>            # start point of the whole contour
START_V=<float>
[BEVEL_DATA ... END_OF_BEVEL_DATA]     # optional, burning contours only
[EXCESS_DATA  EXCESS_VALUE=<float>]    # optional
[AUXILIARY_DATA ... END_OF_AUXILIARY_DATA]  # optional
  AMP_U=<float> AMP_V=<float> AMP=<float>
  RADIUS=<float> SWEEP=<float>
  ORIGIN_U=<float> ORIGIN_V=<float>
  U=<float> V=<float>                  # segment END point; closes the segment
  ... one 9-key group per segment ...
END_OF_CONTOUR
```

Segment keys and semantics:

| Key | Meaning |
| --- | --- |
| `U`, `V` | **End point** of the segment in plate coordinates. Closing `V` commits the segment. |
| `RADIUS` | Arc radius. `0` ⇒ straight line. |
| `SWEEP` | Arc sweep angle in **radians**, signed (sign gives direction). |
| `ORIGIN_U`, `ORIGIN_V` | Arc centre. `0,0` when the segment is a line. |
| `AMP` | Signed amplitude/offset (bulge magnitude). |
| `AMP_U`, `AMP_V` | Amplitude resolved on each axis (bevel/contour normal). |
| `NO_OF_SEG` | Number of segments that follow *(advisory)*. |

**Segment start point.** Only the *first* segment's start is given explicitly
(`START_U`/`START_V`). Each subsequent segment starts at the previous segment's
end `(U, V)`. This is why `U`/`V` "closes" a segment.

**`ORIGIN_U`/`ORIGIN_V` are emitted** by the brokerage writer (present in every
segment of `part.gen`/`nest.gen`) but **omitted** by the `LISTED_PROFILE`
writer (`profiles.gen`), which emits only
`AMP_U, AMP_V, AMP, SWEEP, U, V`. A reader must therefore not require
`RADIUS`/`ORIGIN_U`/`ORIGIN_V`.

**Arc-vs-line rule**: a segment is an arc iff `RADIUS != 0`; `SWEEP` is in
radians and its sign determines traversal direction. Arc midpoint and bulge are
derived, not stored.

Measured:

| Corpus | Segments | `RADIUS != 0` | `RADIUS == 0` |
| --- | ---: | ---: | ---: |
| `part.gen` burning contours | 2764 | 1504 | 1260 |
| `profiles.gen` cross-sections | 14025 (`Σ NO_OF_SEG`) | 0 | 440 emitted, all `0` |

Note the asymmetry: `profiles.gen` emits `RADIUS` only 440 times out of 14025
segments, so **`RADIUS` is optional** — absent means straight. That is the
strongest argument for not requiring `RADIUS` in a reader.

**Contour closure.** `START_OF_CONTOUR` may carry a numeric suffix
(`START_OF_CONTOUR1`, `START_OF_CONTOUR2`), with matching
`END_OF_CONTOUR1`/`2`. These appear in profile end-cut data where two contours
(left and right flank of an end cut) coexist.

---

## 5. Bevel encoding

```
BEVEL_DATA
BEVEL=<NONE|TS|OS|BOTH|CVBA>
BEVEL_CVBA=<...>            # only when BEVEL=CVBA
PLATE_THICKNESS=<float>
BEVEL_CODE=<float>          # signed; sign flips with plate side
BEVEL_TYPE=<int>
BEVEL_VARIANT=<int>
BEVEL_NAME=<string>
MANUAL=<...>
E=<float>  GAP=<float>  NOSE=<float>  H=<float>
H_FACT=<float>  H_FACT_ADJUST=<float>
ANGLE_TS=<float>  ANGLE2_TS=<float>  ANGLE_OS=<float>  ANGLE2_OS=<float>
DEPTH_TS=<float>  DEPTH_OS=<float>
CHAMFER_WIDTH_TS=<float>  CHAMFER_WIDTH_OS=<float>
CHAMFER_HEIGHT_TS=<float> CHAMFER_HEIGHT_OS=<float>
ANGLE2_WTS=<float>  ANGLE2_WOS=<float>
WELD_TS=<float>  WELD_OS=<float>
END_OF_BEVEL_DATA
```

`BEVEL_CODE` sign encodes which material side carries the bevel. Observed values
in `nest.gen`: `0.00`, `±116.00`, `±145.00`, `±273.00`. `BEVEL_TYPE` observed:
`0` (none), `20`, `30`, `35`, `40`.

`ENDCUT_*` (profile end cuts) uses a parallel but distinct code space:
`ENDCUT_TYPE=21|27`, `ENDCUT_MASK` a bit-string such as `00010110000`,
`ENDCUT_CODE` an integer such as `2131`.

---

## 6. `LISTED_PROFILE` documents

> **Evidence upgrade.** This section is no longer sample-derived only. `mar.dll`
> stores its emission literals **in first-reference order**, and each
> `KEY=` literal was recovered in place (`.rdata` 0x3250b80–0x3253ff0, image base
> 0x10000000). The sequence below is that literal order, which is the writer's
> emission order. Sample cross-check: all keys below that appear in
> `testdata/profiles.gen` appear there in the same relative order.

### 6.1 Document header

```
TYPE_OF_GENERIC_FILE=LISTED_PROFILE
VERSION=1.0
USAGE=LIST
COMMON_DATA
  NEST_NAME=
  SHAPE=PP
  START_OF_CONTOUR ... END_OF_CONTOUR     # profile cross-section
  DIMENSION=, WEB_HEIGHT=, WEB_THICKNESS=, MAX_WEB_HEIGHT=,
  FLANGE_WIDTH=, FLANGE_THICKNESS=,
  RAW_LENGTH=, REST_LENGTH=, USED_LENGTH=, LAST_SCRAP_LENGTH=,
  QUALITY=, NO_OF_PROFS=, REST=, GSD=, GSD_DIST=, GSD_ONE_LIM=
END_OF_COMMON_DATA
```

There is **no `GENERAL_DATA` marker** in this family — the document begins
directly with `TYPE_OF_GENERIC_FILE=`.

### 6.2 Record structure — a sequence of siblings

This is the single most important structural fact, and the one the Go package
gets wrong (see findings Bug 2):

```
PROFILE_DATA
  ...attribute block...
END_OF_PROFILE_DATA          <-- closes ONLY the attribute block
LEFT_END
  ...end-cut + end contour...
END_OF_LEFT_END
MARKING_LEFT_END
  ...SEGMENT, START_OF_FCONTOUR/END_OF_FCONTOUR,
         START_OF_FCONTOUR2/END_OF_FCONTOUR2...
END_OF_MARKING_LEFT_END
RIGHT_END
  ...
END_OF_RIGHT_END
MARKING_RIGHT_END
  ...
END_OF_MARKING_RIGHT_END
[REF_LINE_MARKING
   REF_NO=, ORDER_NO=, DIST_LEFT=
 END_OF_REF_LINE_MARKING]
[GEOMETRY_DATA ... END_OF_GEOMETRY_DATA]        # 0..N
[STRING_DATA   ... END_OF_STRING_DATA]          # 0..N
[HOLES_NOTCHES_CUTOUTS ... END_OF_HOLES_NOTCHES_CUTOUTS]
[BENDING_TABLE ... END_OF_BENDING_TABLE]
[GENERAL_MARKING_DATA ... END_OF_GENERAL_MARKING_DATA]
[CONNECTION_TRACE ... END_OF_CONNECTION_TRACE]
[CURVE_DATA ... END_OF_CURVE_DATA]
[CURVE_3D_DATA ...]
... next PROFILE_DATA ...            <-- a record ends here
REST_DATA
  REST_NAME=
END_OF_REST_DATA
```

Confirmed by literal order in `mar.dll`:
`... CASE=`, `END_OF_PROFILE_DATA`, `LEFT_END`, `LEFT_ORIGIN=`, …, `END_OF_LEFT_END`,
`MARKING_LEFT_END`, …, `END_OF_MARKING_LEFT_END`, `RIGHT_END`, `END_OF_RIGHT_END`,
`MARKING_RIGHT_END`, `END_OF_MARKING_RIGHT_END`, `REF_LINE_MARKING`, …,
`REST_DATA`, `REST_NAME=`, `END_OF_REST_DATA`.

`END_OF_PROFILE_DATA` is therefore **not** the end of the profile record. A
reader must keep the current profile open until the next `PROFILE_DATA`,
`REST_DATA` or EOF.

### 6.3 `PROFILE_DATA` attribute block — authoritative order

Emission order per `mar.dll` literals (each literal already contains the `=`):

```
PROFILE_DATA
GENERATED_AS=          SYMMETRICAL=
MLENGTH=               MLENGTH_MANUAL=
TLENGTH=               TLENGTH_MANUAL=
END_POINT_DIST=        END_POINT_DIST_PROD=
TRACE_LENGTH=          TRACE_LENGTH_PROD=
TWIST_ANGLE=           APPROX_PART_WEIGHT=
IDENT_STRING=          PARTNAME_SHORT=
SHIP_NAME=             PROF_SIDE=          SIDE_UP=
DIRECTION_SIGN=        DIRECTION_TRACE=    DIRECTION_FACE=
BEVEL_TRACE=           CHAMFER=            H_FACT=        H_FACT_ADJUST=
GRINDING=
ASSEMBLY=              ASSEMBLY_SB=
ASSEMBLY_TOT=          ASSEMBLY_TOT_SB=
SURFACE_TREATMENT=     ASSEMBLY_SEQUENCE=
BAUGRUPPE=             ZEICHNUNGSNO=       ZAEHLNO=
BLOCK_NAME=            NO_OF_REF_LINES=    MIRROR=
FORM=                  NO_OF_BENDING_TABLES=  TABLE_TYPE=   BEND_DIRECTION=
ADDITIONAL_INFO1=  ... ADDITIONAL_INFO12=
MARKING_LEFT_EXCIST=   MARKING_RIGHT_EXCIST=
NO_OF_MACS=            NO_OF_MARKS=
KNUCKLE_U_START=       KNUCKLE_V_START=    KNUCKLE_U_END=   KNUCKLE_V_END=
KNUCKLE_INCLINATION=
GAP_ADJUST_LEFT=       GAP_ADJUST_RIGHT=
CASE=
END_OF_PROFILE_DATA
```

Keys the post-processor style (`=`-separated) writer adds around the block:
`NAME=`, `POSNO=`, `PARTNAME_LONG=`, `SHIP_NO=`, `DIRECTION=`, `BLOCK_NO=`,
`CUST_DATA_1..4=`, `CUST_TEXT_1..10=`, `PLANNING_UNIT=`, `WORKING_LOCATION=`,
`BUILDING_STRATEGY=` — present in `profiles.gen` though not in the recovered
literal run above (they are shared with the plate writer and therefore
deduplicated elsewhere in `.rdata`).

### 6.4 `LEFT_END` / `RIGHT_END` — end cut

Both ends share one schema. Emission order:

```
LEFT_END
LEFT_ORIGIN=          LEFT_CLOSEST_POINT=     LEFT_FARTHEST_POINT=   LEFT_V0=
LEFT_CLOSEST_POINT_WEB=                       LEFT_FARTHEST_POINT_WEB=
BURN_ECUT=
ENDCUT_TYPE=          ENDCUT_MASK=            ENDCUT_CODE=           ENDCUT_CODEO=
ENDCUT_NAME=
R1=  R2=
V1=  V2=  V3=  V4=
AMPL=  KS=
ENDCUT_USER_DEFINED_NAME=   ENDCUT_USER_DEFINED_PARAM<n>=   ENDCUT_USER_DEFINED_DATA=
EXCESS=
CONNECTION_ANGLE=     WEB_SEG=
# flange bevel 1
BEVEL_DEFINED_FLANGE=  BEVEL_CODE_FLANGE=  BEVEL_TYPE_FLANGE=  BEVEL_VARIANT_FLANGE=
E_FLANGE=  GAP_FLANGE=  CHAMFER_FLANGE=  ALPHA_FLANGE=  ALPHA2_FLANGE=
BETA_FLANGE=  BETA2_FLANGE=  NOSE_FLANGE=  H_FLANGE=
H_FACT_FLANGE=  H_FACT_ADJUST_FLANGE=  RAD_FLANGE=  RADB_FLANGE=
NOSE_ANGLE_FLANGE=  NOSE_LEN_A_FLANGE=  NOSE_LEN_B_FLANGE=
ANGLE_TS_FLANGE=  ANGLE2_TS_FLANGE=  ANGLE_OS_FLANGE=  ANGLE2_OS_FLANGE=
DEPTH_TS_FLANGE=  DEPTH_OS_FLANGE=
CHAMFER_WIDTH_TS_FLANGE=  CHAMFER_WIDTH_OS_FLANGE=
ANGLE2_WTS_FLANGE=  ANGLE2_WOS_FLANGE=
CHAMFER_HEIGHT_TS_FLANGE=  CHAMFER_HEIGHT_OS_FLANGE=
CONNECTION_ANGLE_FLANGE=  FLA_SEG=
# flange bevel 2 (identical set, suffix _FLANGE2)
BEVEL_DEFINED_FLANGE2= ... FLA2_SEG=
GSD=  GSD_DIST=
KNUCKLE_HEIGHT=  KNUCKLE_RATIO=  LENGTH_FROM_EDGE=  EXTRA_LENGTH=
SNIP_LENGTH=  KNUCKLE_GAP=  FLANGE_BEVEL_ANGLE=  KNUCKLE_ANGLE=
FLANGE_EXTENSION=
START_OF_CONTOUR ... END_OF_CONTOUR                 # end profile
START_OF_FCONTOUR ... END_OF_FCONTOUR               # flange end profile
START_OF_FCONTOUR2 ... END_OF_FCONTOUR2             # flange 2 end profile
END_OF_LEFT_END
```

Measured in `profiles.gen`: all 880 ends carry both a `Contour` (2 segments) and a
`FlangeContour`. `ENDCUT_TYPE` ∈ {21 (760 ends), 27 (120 ends)}.

### 6.5 The `-` sentinel

`A`, `B`, `C`, `R2`, `V3`, `V4` may be a bare `-`, meaning *parameter not
applicable to this end-cut type*. This is systematic, not corruption — measured
across all 880 end blocks in `profiles.gen`:

| `ENDCUT_TYPE` | `A` | `R2` | `B`, `C`, `V3`, `V4` | blocks |
| --- | --- | --- | --- | ---: |
| 21 | `-` | `-` | `-` | 760 |
| 27 | `42` | `42` | `-` | 120 |

100 % correlation with no exceptions. `B`, `C`, `V3`, `V4` are `-` on **every**
block, i.e. reserved/unused in this configuration.

A parser must accept `-` wherever a number may appear. Mapping it to `0` is
acceptable only if the field is separately marked "undefined", because `0` is a
legal radius.

### 6.6 `MARKING_LEFT_END` / `MARKING_RIGHT_END`

Per end, an optional marking block carrying `SEGMENT` plus
`START_OF_FCONTOUR`/`END_OF_FCONTOUR` and
`START_OF_FCONTOUR2`/`END_OF_FCONTOUR2` contours. Not read by the Go package.

### 6.7 `REF_LINE_MARKING`

```
REF_LINE_MARKING
REF_NO=   ORDER_NO=   DIST_LEFT=
END_OF_REF_LINE_MARKING
```

Not read by the Go package.

### 6.8 `HOLES_NOTCHES_CUTOUTS` — sub-structured

Richer than the plate `Hole`. Attributes `OPEN=`, `DIST_ORIGIN=`,
`DIST_ORIGIN_TO_ZEROP=`, `DIST_ORIGIN_V=`, `ROTATION=`, then one of:

| Sub-block | Terminator | Fields |
| --- | --- | --- |
| `PENETRATING_PROFILE` | `END_OF_PENETRATING_PROFILE` | `DIM=`, `INTERSECTING_ANGLE=`, `INCLINATION_ANGLE=`, `MTRL_PENETRATING=` (`RIGHT`/…) |
| `DRILLING` | `END_OF_DRILLING` | contains `THREADING` → `THREAD_START=` → `END_OF_THREADING` |
| `SLOTTED` | `END_OF_SLOTTED` | `LENGTH=`, `WIDTH=` |

Terminator: `END_OF_HOLES_NOTCHES_CUTOUTS`.

### 6.9 Other profile-record sections (all unread by the Go package)

| Section | Terminator | Fields |
| --- | --- | --- |
| `BENDING_TABLE` | `END_OF_BENDING_TABLE` | `NO_OF_POINTS=` (documented limit 100 points) |
| `GENERAL_MARKING_DATA` | `END_OF_GENERAL_MARKING_DATA` | `ATTR=` |
| `CONNECTION_TRACE` | `END_OF_CONNECTION_TRACE` | contour |
| `CURVE_DATA` | `END_OF_CURVE_DATA` | `START_OF_MCONTOUR`/`END_OF_MCONTOUR`, `START_OF_NCONTOUR`/`END_OF_NCONTOUR` |
| `CURVE_3D_DATA` | — | 3D curve data |
| `REST_DATA` | `END_OF_REST_DATA` | `REST_NAME=` (tail of document) |

### 6.10 A second, unrelated serialisation lives in the same binary

`mar.dll` also contains a **comma/semicolon** dialect of the same schema:

```
BEVEL_DATA;
ANGLE_TS,%.2lf;
END_OF_BEVEL_DATA;
LEFT_END;
ENDCUT_CODE,'%s';
```

That is a *different* serialisation (`,` separates key from value, `;`
terminates). Its sibling order differs (there `LEFT_END` precedes the block that
ends with `END_OF_PROFILE_DATA`). Do **not** apply the ordering rules of §6.2 to
that dialect, and do not implement it by extending the `=` reader. It is out of
scope for the `.gen` text format but is worth knowing about because grepping
`mar.dll` for these key names returns both dialects interleaved.


## 7. Observed enumeration values

Values actually present in shipped samples. Treat these as an open set — the
format has no schema and unknown values must pass through, not error.

| Key | Observed values |
| --- | --- |
| `TYPE_OF_GENERIC_FILE` | `PLATE_PART`, `LISTED_PROFILE`, `PLATE_PART_CHECK` (post-processor template), `ASSEMBLY_PART` |
| `TYPE_OF_MANUFACT` | `2AXIS`, `3AXIS` |
| `TYPE_OF_WORK` | `BURNING` |
| `PLATE_SIDE` | `SB`, `PS`, `TS` |
| `PROF_SIDE` | `PS`, `SB` |
| `MIRRORED` | `0`, `1` |
| `MIRROR` | `NO`, `YES` |
| `MARKING_SIDE` | `TS`, `OS` |
| `MARKING_SHIPSIDE` | `BOTH`, `SB`, `TS` |
| `MARKING_TYPE` | `PANEL`, … |
| `SHAPE` | `OUTER_CONTOUR`, `HOLE`, `PP` (profile cross-section) |
| `GEOMETRY_TYPE` | `OUTER_CONTOUR`, `HOLE`, `ROLL_AXIS`, `ROLL_AXIS_CROSS_SIGN`, `BEVEL_SYMBOL` |
| `STRING_TYPE` | `PART_NAME`, `THICKNESS_&_QUALITY`, `BEVEL_ANGLE`, `EXCESS_LIMIT`, `EXCESS_GEOMETRY`, `DIR_TOP`, `DIR_RIGHT`, `1`, `2`, `3`, `4`, `5` |
| `BEVEL` | `NONE`, `TS`, `OS`, `BOTH`, `CVBA` |
| `KNUCKLED` | `NO`, `YES` |
| `DIRECTION` (profile) | `UP`, `DOWN`, `CL`, `OUTB` |
| `DIRECTION` (burning) | `0`, `1` |
| `SIDE_UP` | `MOULD` |
| `FORM` | `STRAIGHT`, … |
| `GENERATED_AS` | `STIFFENER` |
| `DIRECTION_TRACE` | `FOR`, `AFT` |
| `ENDCUT_TYPE` | `21`, `27` |

---

## 8. Numeric formatting conventions

**Formatting is per writer, not per key.** The two writers observed here use
opposite conventions, so a spec must not state a single format per key.

### 8.1 Plate family (`part.gen`, `nest.gen`) — fixed-width

Every numeric field is padded to a fixed number of decimals. Measured: **100 %**
of occurrences in both files use the stated width.

| Field group | Format | Examples |
| --- | --- | --- |
| Areas / totals / raw plate | `%.2f` | `PART_AREA=19750083.01`, `RAW_LENGTH=19500.00`, `RAW_WIDTH=3200.00`, `RAW_THICKNESS=11.00`, `TOTAL_BURNING=313330.41`, `DENSITY=7850.00` |
| Plate extension / COG | `%.5f` | `EXTENSION_U=3062.51539`, `PART_COG_U=1539.85835` |
| Contour `U`/`V`/`ORIGIN_*` | `%.5f` | `U=3064.05509`, `ORIGIN_U=0.00000` |
| `AMP`,`AMP_U`,`AMP_V`,`RADIUS`,`SWEEP` | `%.5f` | `AMP=-2.05000`, `SWEEP=1.57080` |
| `MTRL_SIDE_U/V` | `%.5f` | `MTRL_SIDE_U=0.00000` |
| `DISTANCE_Y1_Y2` | `%.2f` | `DISTANCE_Y1_Y2=0.00` |
| `BEVEL_CODE` | `%.2f` | `BEVEL_CODE=0.00`, `BEVEL_CODE=-273.00` |
| Counters (`NO_OF_*`, `NUMBER_OF_HEADS`, `NO_OF_SEG`, `MIRRORED`) | integer | `NO_OF_SEG=9`, `MIRRORED=0` |

### 8.2 `LISTED_PROFILE` (`profiles.gen`) — minimal digits

The profile writer emits numbers with **no fixed padding**. Measured
distributions:

| Field | Decimal counts observed |
| --- | --- |
| `AMP`, `AMP_U`, `AMP_V` | `0` ×8480, `2` ×875 |
| `SWEEP` | `0` ×8480, `3` ×1, `4` ×874 |
| `U` | `0` ×6533, `1` ×2202, `2` ×620 |
| `V` | `0` ×4919, `1` ×1908, `2` ×2528 |
| `RADIUS` | `0` ×440 (always literally `0`) |
| `RAW_LENGTH` | `0` ×1 (`853726`) |
| `DENSITY` | `8` ×1 (`7850.00000000`) |

So the same semantic field can appear as `U=0`, `U=7`, `U=38.5` or
`U=3064.05509` depending on which writer produced the file.

### 8.3 Consequences for a reader

* **Never assume a fixed width**, never `%0.Nf`-style parse, and never
  round-trip by re-formatting with the plate widths for a profile file.
* Numbers may be written with **no decimal point at all** (`MLENGTH=2430`,
  `LEFT_ORIGIN=0`, `BEVEL_CODE=145`).
* Precision is lossy in the plate family: `AMP` is stored to 5 decimals, so
  values round-trip to ~1e-5 in plate coordinates. Do not treat parsed floats as
  exact.
* `ENDCUT_MASK` is the one non-numeric scalar: a zero-padded bit string such as
  `00010110000`. Keep it as a **string**, never as an integer.
* `PRODUCTION_DATE` / `SPLIT_DATE` are `YYYY-MM-DD`; `PRODUCTION_DATE` may be
  all-spaces when unset (10 spaces observed).

---

## 9. Two-axis vs three-axis

`TYPE_OF_MANUFACT` selects the manufacturing mode:

* `2AXIS` — planar cutting. Contours are plain `(U, V)` polylines. `nest.gen`
  is `2AXIS`.
* `3AXIS` — bevel/3D cutting. Contours carry per-segment `AMP`, `AMP_U`,
  `AMP_V` and `BEVEL_DATA`. `part.gen` is `3AXIS`.

Both use the identical contour grammar; `3AXIS` simply populates more segment
keys. A reader should not branch on the mode.

---

## 10. Parse algorithm (reference)

The shipped reader does this *(observed: `FUN_10035750` = the generic-file
driver, `FUN_1014d210` = the `KEY=VALUE` splitter, `FUN_1014b0e0` = the key→token
mapper)*:

1. Open the file as a line stream.
2. For each line, split on the **first** `=`.
3. Trim leading and trailing spaces from the key and from the value.
4. Map the key to an integer token id; store the token id on the current
   section object.
5. Dispatch to the current section object's `Process(token, value)` virtual
   method, which returns:
   * `1` — key not recognised in the current section → log
     `Unrecognised Generic data at line N, (key)` and continue,
   * `2` — ignore (blank/comment),
   * `3` — section complete.
6. On a section-open token (`GENERAL_DATA`, `PART_DATA`, `MARKING_DATA`,
   `GSD_DATA`, `MARKING_SHAPE_DATA`, `LABELTEXT_DATA`, `LABELSYMBOL_DATA`,
   `BURNING_DATA`, `IDLE_DATA`, `BLASTING_DATA`, `AUXILIARY_DATA`,
   `PART_INFORMATION`, `GEOMETRY_DATA`, `STRING_DATA`) allocate a new object of
   the corresponding class and make it current.
7. On a section-close token hand the object to the document.
8. Unknown keys are **tolerated and logged**, never fatal.

Key robustness properties to preserve:

* **Unknown sections and unknown keys must be skipped, not fatal.** The format
  is version-extended additively (e.g. `_SB` variants, `FUNCTIONAL_DESCR_*`,
  `ADDITIONAL_INFO1..12` are late additions).
* **Section-close is authoritative, not the count fields.** `NO_OF_PARTS` and
  `NO_OF_SEG` are advisory; a truncated or over-long stream must still parse
  structurally.
* **`PART_INFORMATION` is the scope switch** for everything that follows until
  the next `PART_INFORMATION`.

---

## 11. Unknowns

Everything the earlier drafts listed as unknown is now either resolved or
explicitly bounded as unobservable with the available artifacts.

### Resolved

| Former unknown | Answer | Evidence |
| --- | --- | --- |
| `AMP` vs `AMP_U`/`AMP_V` | `\|AMP\| == hypot(AMP_U, AMP_V)` holds for **arc segments only** (1501/1504 arcs in `part.gen`); it does not hold for lines | measured |
| Is a single-part `PLATE_PART` with empty `TYPE_OF_GENERIC_FILE` legal? | Yes — `testdata/part.gen` and corpus files | measured |
| Might `END_OF_IDLE_DATA` be omitted? | Not observed; every `IDLE_DATA` in the corpus is terminated | corpus |
| Is a `;` comment form legal? | No — no corpus file contains one, and the reader has no comment handling | corpus |
| Can `PART_DATA` be absent? | Yes — 6 corpus files declare the part only via `PART_INFORMATION` | corpus |
| Can `PART_NAME` be empty? | Yes — a nameless scope denotes a nest-wide pass (e.g. `TYPE_OF_WORK=LABELLING`) | corpus |
| Is `DIRECTION` flipped by mirroring? | No — `MIRRORED=0` and `=1` are both dominantly `DIRECTION=1` | corpus |
| Are numeric values always present? | No — empty values and `-` both occur | corpus |
| Field set of `LABELTEXT_DATA`…`EDGE_DATA` | Enumerated in §3.7 | corpus |

### Bounded as unobservable

These cannot be resolved without artifacts not present here. They are recorded,
not guessed:

* Field **semantics** (not just names) of `MARKING_LEFT_END`,
  `MARKING_RIGHT_END`, `REF_LINE_MARKING`, `BENDING_TABLE`,
  `GENERAL_MARKING_DATA`, `CURVE_DATA`, `CURVE_3D_DATA`, `BLASTING_DATA`,
  `GSD_DATA`, `MARKING_SHAPE_DATA`, `BLASTING_MARKING_DATA` and the hole
  sub-blocks — zero corpus occurrences, so nothing can validate an
  interpretation. The readers capture these verbatim as `RawSection` (or model
  them from binary literals, marked unvalidated) so no data is lost meanwhile.
* Exact **types** of the profile-side scalars beyond what the `%`-formatted
  literals and samples show.
* Whether `END_OF_PROFILE_DATA` may ever be omitted by an older writer.
* The `,`/`;` dialect in `mar.dll` (§6.10): a separate serialiser, deliberately
  not modelled.
* Runtime behaviour: everything here is static artifact + sample evidence; no
  AVEVA component was executed.

To close the unobservable items, run the corresponding writer (or collect
`.gen` output from a site that uses those machine configurations) and the
`RawSection` captures will show the real field sets without any code change.
