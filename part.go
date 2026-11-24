package gen

import (
	"bufio"
	"bytes"
)

type Part struct {
	Thickness           float64
	Quality             string
	Name                string
	PosNo               string
	PartnameLong        string
	ShipNo              string
	AssemblyLow         string
	AssemblyHigh        string
	BlockNo             string
	PartArea            float64
	Mirrored            bool
	ExtensionU          float64
	ExtensionV          float64
	PlateSide           string
	PartCog             *Point
	NoIntervalsExcess1  int
	LengthExcess1       float64
	NoIntervalsExcess2  int
	LengthExcess2       float64
	NoIntervalsExcess3  int
	LengthExcess3       float64
	StringData          []StringData
	BurningData         []BurningData
	MarkingData         []Marking
	GeometryData        []Geometry
	Quantity            int
	PartnameLongSB      string
	AssemblyLowSB       string
	AssemblyHighSB      string
	CustData1           string
	CustData2           string
	CustData3           string
	CustData4           string
	CustText1           string
	CustText2           string
	CustText3           string
	CustText4           string
	CustText5           string
	CustText6           string
	CustText7           string
	CustText8           string
	CustText9           string
	CustText10          string
	WorkingLocation     string
	BuildingStrategy    string
	PlanningUnit        string
	PanelName           string
	ClipGrindingSide    string
	Comment             string
	Destination         string
	PartID              int
	FunctionalDescrCode string
	FunctionalDescrText string
	TypeOfWork          string
	Knuckled            string
	NestedOn            string
	RawName             string
	Information         []PartInformation
	LabelTexts          []LabelText
	LabelSymbols        []LabelSymbol
	Auxiliaries         []Auxiliary
	Edges               []Edge
	RawSections         []RawSection
	IdleData            []IdleData
	BumpData            []BumpData
	Transformation      *TransformationData
}

func readPart(s *bufio.Scanner) (*Part, error) {
	part := &Part{}
	var partCog Point

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "TRANSFORMATION_DATA" {
			t, err := readTransformationData(s)
			if err != nil {
				return nil, err
			}
			part.Transformation = t

			continue
		}

		if !found {
			break
		}

		switch key {
		case "NAME":
			part.Name = string(value)

		case "POSNO":
			part.PosNo = string(value)

		case "PARTNAME_LONG":
			part.PartnameLong = string(value)

		case "PARTNAME_LONG_SB":
			part.PartnameLongSB = string(value)

		case "SHIP_NO":
			part.ShipNo = string(value)

		case "ASSEMBLY_LOW":
			part.AssemblyLow = string(value)

		case "ASSEMBLY_LOW_SB":
			part.AssemblyLowSB = string(value)

		case "ASSEMBLY_HIGH":
			part.AssemblyHigh = string(value)

		case "ASSEMBLY_HIGH_SB":
			part.AssemblyHighSB = string(value)

		case "BLOCK_NO":
			part.BlockNo = string(value)

		case "PART_AREA":
			partArea, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.PartArea = partArea

		case "MIRRORED":
			mirrored, err := atob(value)
			if err != nil {
				return nil, err
			}
			part.Mirrored = mirrored

		case "EXTENSION_U":
			extensionU, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.ExtensionU = extensionU

		case "EXTENSION_V":
			extensionV, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.ExtensionV = extensionV

		case "PLATE_SIDE":
			part.PlateSide = string(value)

		case "PART_COG_U":
			x, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			partCog.X = x

		case "PART_COG_V":
			y, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			partCog.Y = y

		case "NO_INTERVALS_EXCESS_1":
			n, err := atoi(value)
			if err != nil {
				return nil, err
			}
			part.NoIntervalsExcess1 = n

		case "LENGTH_EXCESS_1":
			length, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.LengthExcess1 = length

		case "NO_INTERVALS_EXCESS_2":
			n, err := atoi(value)
			if err != nil {
				return nil, err
			}
			part.NoIntervalsExcess2 = n

		case "LENGTH_EXCESS_2":
			length, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.LengthExcess2 = length

		case "NO_INTERVALS_EXCESS_3":
			n, err := atoi(value)
			if err != nil {
				return nil, err
			}
			part.NoIntervalsExcess3 = n

		case "LENGTH_EXCESS_3":
			length, err := atof(value)
			if err != nil {
				return nil, err
			}
			part.LengthExcess3 = length

		case "CUST_DATA_1":
			part.CustData1 = string(value)

		case "CUST_DATA_2":
			part.CustData2 = string(value)

		case "CUST_DATA_3":
			part.CustData3 = string(value)

		case "CUST_DATA_4":
			part.CustData4 = string(value)

		case "CUST_TEXT_1":
			part.CustText1 = string(value)

		case "CUST_TEXT_2":
			part.CustText2 = string(value)

		case "CUST_TEXT_3":
			part.CustText3 = string(value)

		case "CUST_TEXT_4":
			part.CustText4 = string(value)

		case "CUST_TEXT_5":
			part.CustText5 = string(value)

		case "CUST_TEXT_6":
			part.CustText6 = string(value)

		case "CUST_TEXT_7":
			part.CustText7 = string(value)

		case "CUST_TEXT_8":
			part.CustText8 = string(value)

		case "CUST_TEXT_9":
			part.CustText9 = string(value)

		case "CUST_TEXT_10":
			part.CustText10 = string(value)

		case "WORKING_LOCATION":
			part.WorkingLocation = string(value)

		case "BUILDING_STRATEGY":
			part.BuildingStrategy = string(value)

		case "PLANNING_UNIT":
			part.PlanningUnit = string(value)

		case "PANEL_NAME":
			part.PanelName = string(value)

		case "CLIP_GRINDING_SIDE":
			part.ClipGrindingSide = string(value)

		case "COMMENT":
			part.Comment = string(value)

		case "DESTINATION":
			part.Destination = string(value)

		case "PART_ID":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			part.PartID = v

		case "FUNCTIONAL_DESCR_CODE":
			part.FunctionalDescrCode = string(value)

		case "FUNCTIONAL_DESCR_TEXT":
			part.FunctionalDescrText = string(value)

		case "TYPE_OF_WORK":
			part.TypeOfWork = string(value)

		case "KNUCKLED":
			part.Knuckled = string(value)

		case "NESTED_ON":
			part.NestedOn = string(value)

		case "RAW_NAME":
			part.RawName = string(value)
		}
	}

	part.PartCog = &partCog
	part.Quantity = 1
	return part, nil
}

// Mirror flips the part about its U axis: contour winding is reversed, marking
// sides swap TS/OS, and every absolute U coordinate is negated.
//
// Deliberately NOT flipped:
//
//   - BurningData.Direction / Marking.MarkingDirection. Measured over the
//     45,750-file corpus, MIRRORED=0 and MIRRORED=1 documents are both
//     dominantly DIRECTION=1 (and both contain -1), so DIRECTION is not a
//     function of mirroring and must not be inferred from it.
//   - Marking.MarkingAttri / assembly references: they are text, not geometry.
func (p *Part) Mirror() {
	p.Mirrored = !p.Mirrored

	if p.PartCog != nil {
		p.PartCog.X = -p.PartCog.X
	}

	// Every loop below indexes the slice rather than ranging over it: Part's
	// payload slices hold values, so ranging would mutate a copy and silently
	// drop the change.
	for i := range p.StringData {
		if p.StringData[i].Pos != nil {
			p.StringData[i].Pos.X = -p.StringData[i].Pos.X
		}
	}

	for i := range p.LabelTexts {
		p.LabelTexts[i].PositionU = -p.LabelTexts[i].PositionU
		for j := range p.LabelTexts[i].Geometry {
			p.LabelTexts[i].Geometry[j].Contour.invert()
		}
	}

	for i := range p.LabelSymbols {
		p.LabelSymbols[i].PositionU = -p.LabelSymbols[i].PositionU
	}

	for i := range p.Edges {
		p.Edges[i].StartU = -p.Edges[i].StartU
		p.Edges[i].EndU = -p.Edges[i].EndU
	}

	for i := range p.BurningData {
		p.BurningData[i].Contour.invert()
		if p.BurningData[i].GeometryData != nil {
			p.BurningData[i].GeometryData.Contour.invert()
		}
	}

	for i := range p.MarkingData {
		if p.MarkingData[i].MarkingSide == "TS" {
			p.MarkingData[i].MarkingSide = "OS"
		} else {
			p.MarkingData[i].MarkingSide = "TS"
		}
		p.MarkingData[i].Contour.invert()
	}

	for i := range p.GeometryData {
		p.GeometryData[i].Contour.invert()
	}

	for i := range p.IdleData {
		p.IdleData[i].Contour.invert()
	}

	for i := range p.BumpData {
		p.BumpData[i].Contour.invert()
	}
}

// TransformationData is the TRANSFORMATION_DATA block inside PART_DATA: the
// placement transform applied to the part.
type TransformationData struct {
	Type            string // TRANSFORMATION_TYPE, observed value REDUCED
	Transformation1 float64
	Transformation2 float64
	Transformation3 float64
	Transformation4 float64
	Transformation5 float64
	Transformation6 float64
}

func readTransformationData(s *bufio.Scanner) (*TransformationData, error) {
	t := &TransformationData{}

	for s.Scan() {
		if bytes.Equal(s.Bytes(), endOfTransformationData) {
			return t, nil
		}

		key, value, found := nextKeyValue(s)
		if !found {
			return t, nil
		}

		var err error

		switch key {
		case "TRANSFORMATION_TYPE":
			t.Type = string(value)
		case "TRANSFORMATION_1":
			t.Transformation1, err = parseOptBytes(value)
		case "TRANSFORMATION_2":
			t.Transformation2, err = parseOptBytes(value)
		case "TRANSFORMATION_3":
			t.Transformation3, err = parseOptBytes(value)
		case "TRANSFORMATION_4":
			t.Transformation4, err = parseOptBytes(value)
		case "TRANSFORMATION_5":
			t.Transformation5, err = parseOptBytes(value)
		case "TRANSFORMATION_6":
			t.Transformation6, err = parseOptBytes(value)
		}

		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
