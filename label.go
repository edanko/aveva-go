package gen

import (
	"bufio"
	"bytes"
)

// PartInformation is one PART_INFORMATION scope. It names the part a following
// payload block belongs to and classifies the work. A part may be opened by more
// than one scope, so Part.Information is a list, in document order.
type PartInformation struct {
	PartName   string
	TypeOfWork string
}

// LabelText is a LABELTEXT_DATA block: a marking text plus the vector strokes
// that draw it. TEXT_ELEMENTS is the glyph stroke count; each stroke is one
// nested GEOMETRY_DATA block.
type LabelText struct {
	Type      string // TEXT_TYPE, space padded in the file
	PositionU float64
	PositionV float64
	Angle     float64
	Height    float64
	Direction int
	Side      string // TEXT_SIDE: BOTH, TS, OS
	Text      string
	Elements  int        // TEXT_ELEMENTS: number of glyph strokes
	Geometry  []Geometry // the strokes
}

func readLabelText(s *bufio.Scanner) (LabelText, error) {
	l := LabelText{}

next:
	for s.Scan() {
		key, value, found := nextKeyValue(s)

		switch key {
		case "GEOMETRY_DATA":
			g, err := readGeometry(s)
			if err != nil {
				return LabelText{}, err
			}
			l.Geometry = append(l.Geometry, g)
			continue next

		case "END_OF_LABELTEXT_DATA":
			return l, nil
		}

		if !found {
			return l, nil
		}

		var err error

		switch key {
		case "TEXT_TYPE":
			l.Type = cleanBytes(value)
		case "TEXT_POSITION_U":
			l.PositionU, err = parseOptBytes(value)
		case "TEXT_POSITION_V":
			l.PositionV, err = parseOptBytes(value)
		case "TEXT_ANGLE":
			l.Angle, err = parseOptBytes(value)
		case "TEXT_HEIGHT":
			l.Height, err = parseOptBytes(value)
		case "TEXT_DIRECTION":
			l.Direction, err = parseOptIntBytes(value)
		case "TEXT_SIDE":
			l.Side = cleanBytes(value)
		case "TEXT":
			l.Text = string(value)
		case "TEXT_ELEMENTS":
			l.Elements, err = parseOptIntBytes(value)
		}

		if err != nil {
			return LabelText{}, err
		}
	}

	return l, nil
}

// LabelSymbol is a LABELSYMBOL_DATA block: a symbol reference placed on the part
// (position marks, margin marks, ...).
type LabelSymbol struct {
	Type      string // SYMBOL_TYPE, space padded in the file
	PositionU float64
	PositionV float64
	Angle     float64
	Height    float64
	Font      int
	Number    int
}

func readLabelSymbol(s *bufio.Scanner) (LabelSymbol, error) {
	sym := LabelSymbol{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "END_OF_LABELSYMBOL_DATA" {
			return sym, nil
		}

		if !found {
			return sym, nil
		}

		var err error

		switch key {
		case "SYMBOL_TYPE":
			sym.Type = cleanBytes(value)
		case "SYMBOL_POSITION_U":
			sym.PositionU, err = parseOptBytes(value)
		case "SYMBOL_POSITION_V":
			sym.PositionV, err = parseOptBytes(value)
		case "SYMBOL_ANGLE":
			sym.Angle, err = parseOptBytes(value)
		case "SYMBOL_HEIGHT":
			sym.Height, err = parseOptBytes(value)
		case "SYMBOL_FONT":
			sym.Font, err = parseOptIntBytes(value)
		case "SYMBOL_NUMBER":
			sym.Number, err = parseOptIntBytes(value)
		}

		if err != nil {
			return LabelSymbol{}, err
		}
	}

	return sym, nil
}

// Auxiliary is an AUXILIARY_DATA block: an opaque machine/macro directive that
// the generic file passes through unchanged.
type Auxiliary struct {
	Function  string // AUXILIARY_FUNCTION
	Parameter string // AUXILIARY_PARAMETER
}

// readAuxiliary keeps the (value, error) shape shared by the other block
// readers, so the section dispatch stays uniform.
func readAuxiliary(s *bufio.Scanner) (Auxiliary, error) { //nolint:unparam // uniform reader signature
	a := Auxiliary{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "END_OF_AUXILIARY_DATA" {
			return a, nil
		}

		if !found {
			return a, nil
		}

		switch key {
		case "AUXILIARY_FUNCTION":
			a.Function = cleanBytes(value)
		case "AUXILIARY_PARAMETER":
			a.Parameter = string(value)
		}
	}

	return a, nil
}

// Edge is an EDGE_DATA block: a measured edge of the part with its along-edge
// and material lengths.
type Edge struct {
	Name    string
	StartU  float64
	StartV  float64
	EndU    float64
	EndV    float64
	ALength float64 // ALENGTH
	MLength float64 // MLENGTH
}

func readEdge(s *bufio.Scanner) (Edge, error) {
	e := Edge{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "END_OF_EDGE_DATA" {
			return e, nil
		}

		if !found {
			return e, nil
		}

		var err error

		switch key {
		case "NAME":
			e.Name = string(value)
		case "START_U":
			e.StartU, err = parseOptBytes(value)
		case "START_V":
			e.StartV, err = parseOptBytes(value)
		case "END_U":
			e.EndU, err = parseOptBytes(value)
		case "END_V":
			e.EndV, err = parseOptBytes(value)
		case "ALENGTH":
			e.ALength, err = parseOptBytes(value)
		case "MLENGTH":
			e.MLength, err = parseOptBytes(value)
		}

		if err != nil {
			return Edge{}, err
		}
	}

	return e, nil
}

// Field is one KEY=VALUE pair captured verbatim.
type Field struct {
	Key   string
	Value string
}

// RawSection captures a section whose field set is not modelled, so that no data
// is silently discarded. The generic file is extended additively and several
// documented sections (MARKING_LEFT_END, REF_LINE_MARKING, BENDING_TABLE,
// CURVE_DATA, PENETRATING_PROFILE, DRILLING, SLOTTED, BLASTING_DATA, GSD_DATA,
// MARKING_SHAPE_DATA, BLASTING_MARKING_DATA) have no sample available; they are
// preserved here rather than dropped.
type RawSection struct {
	Name   string
	Fields []Field
}

// readRawSection consumes lines through END_OF_<name>. Nested blocks are not
// interpreted; their markers are recorded as bare fields with an empty value.
func readRawSection(s *bufio.Scanner, name string) (RawSection, error) {
	rs := RawSection{Name: name}
	terminator := []byte("END_OF_" + name)

	for s.Scan() {
		line := s.Bytes()
		if bytes.Equal(line, terminator) {
			return rs, nil
		}

		if key, value, found := bytes.Cut(line, eqSep); found {
			rs.Fields = append(rs.Fields, Field{Key: string(key), Value: string(value)})
			continue
		}

		rs.Fields = append(rs.Fields, Field{Key: string(bytes.TrimSpace(line))})
	}

	return rs, s.Err()
}
