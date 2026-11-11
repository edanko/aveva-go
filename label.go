package gen

import (
	"bufio"
	"strings"
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
		key, value, found := strings.Cut(s.Text(), "=")

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
			l.Type = cleanStr(value)
		case "TEXT_POSITION_U":
			l.PositionU, err = parseOptFloat(value)
		case "TEXT_POSITION_V":
			l.PositionV, err = parseOptFloat(value)
		case "TEXT_ANGLE":
			l.Angle, err = parseOptFloat(value)
		case "TEXT_HEIGHT":
			l.Height, err = parseOptFloat(value)
		case "TEXT_DIRECTION":
			l.Direction, err = parseOptInt(value)
		case "TEXT_SIDE":
			l.Side = cleanStr(value)
		case "TEXT":
			l.Text = value
		case "TEXT_ELEMENTS":
			l.Elements, err = parseOptInt(value)
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
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "END_OF_LABELSYMBOL_DATA" {
			return sym, nil
		}

		if !found {
			return sym, nil
		}

		var err error

		switch key {
		case "SYMBOL_TYPE":
			sym.Type = cleanStr(value)
		case "SYMBOL_POSITION_U":
			sym.PositionU, err = parseOptFloat(value)
		case "SYMBOL_POSITION_V":
			sym.PositionV, err = parseOptFloat(value)
		case "SYMBOL_ANGLE":
			sym.Angle, err = parseOptFloat(value)
		case "SYMBOL_HEIGHT":
			sym.Height, err = parseOptFloat(value)
		case "SYMBOL_FONT":
			sym.Font, err = parseOptInt(value)
		case "SYMBOL_NUMBER":
			sym.Number, err = parseOptInt(value)
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
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "END_OF_AUXILIARY_DATA" {
			return a, nil
		}

		if !found {
			return a, nil
		}

		switch key {
		case "AUXILIARY_FUNCTION":
			a.Function = cleanStr(value)
		case "AUXILIARY_PARAMETER":
			a.Parameter = value
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
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "END_OF_EDGE_DATA" {
			return e, nil
		}

		if !found {
			return e, nil
		}

		var err error

		switch key {
		case "NAME":
			e.Name = value
		case "START_U":
			e.StartU, err = parseOptFloat(value)
		case "START_V":
			e.StartV, err = parseOptFloat(value)
		case "END_U":
			e.EndU, err = parseOptFloat(value)
		case "END_V":
			e.EndV, err = parseOptFloat(value)
		case "ALENGTH":
			e.ALength, err = parseOptFloat(value)
		case "MLENGTH":
			e.MLength, err = parseOptFloat(value)
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
	terminator := "END_OF_" + name

	for s.Scan() {
		text := s.Text()
		if text == terminator {
			return rs, nil
		}

		if key, value, found := strings.Cut(text, "="); found {
			rs.Fields = append(rs.Fields, Field{Key: key, Value: value})
			continue
		}

		rs.Fields = append(rs.Fields, Field{Key: strings.TrimSpace(text)})
	}

	return rs, s.Err()
}
