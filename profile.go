package gen

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// ReadProfiles parses a LISTED_PROFILE document and returns its profiles.
func ReadProfiles(r io.Reader) ([]*Profile, error) {
	doc, err := ReadProfileDocument(r)
	if err != nil {
		return nil, err
	}

	return doc.Profiles, nil
}

// ReadProfileDocument parses a LISTED_PROFILE document, keeping the header and
// COMMON_DATA so the document can be written back.
func ReadProfileDocument(r io.Reader) (*ProfileDocument, error) {
	s := bufio.NewScanner(r)

	doc := &ProfileDocument{}

	// type
	s.Scan()
	if s.Text() != "TYPE_OF_GENERIC_FILE=LISTED_PROFILE" {
		return nil, ErrNotListedProfile
	}
	// version
	if s.Scan() {
		doc.Version, _ = strings.CutPrefix(s.Text(), "VERSION=")
	}
	// usage
	if s.Scan() {
		doc.Usage, _ = strings.CutPrefix(s.Text(), "USAGE=")
	}

	// common data
	s.Scan()
	if s.Text() != "COMMON_DATA" {
		return nil, ErrNotListedProfile
	}
	commonData, err := readCommonData(s)
	if err != nil {
		return nil, err
	}

	doc.Common = commonData
	profiles := make([]*Profile, 0, commonData.NoOfProfs)

	// A profile record is a sequence of sibling blocks: the attribute block is
	// closed by END_OF_PROFILE_DATA and LEFT_END / RIGHT_END follow it as
	// siblings, not children. Keep the record open until the next PROFILE_DATA.
	var cur *Profile

	for s.Scan() {
		t := s.Text()
		switch t {
		case "PROFILE_DATA":
			profile, err := readProfile(s)
			if err != nil {
				return nil, err
			}
			profile.Shape = commonData.Shape
			profile.Dimension = commonData.Dimension
			profile.Quality = commonData.Quality

			cur = profile
			profiles = append(profiles, profile)

		case "LEFT_END":
			end, err := readEnd(s)
			if err != nil {
				return nil, err
			}
			if cur != nil {
				cur.LeftEnd = end
			}

		case "RIGHT_END":
			end, err := readEnd(s)
			if err != nil {
				return nil, err
			}
			if cur != nil {
				cur.RightEnd = end
			}

		case "REST_DATA":
			for s.Scan() {
				if bytes.Equal(s.Bytes(), []byte("END_OF_REST_DATA")) {
					break
				}
			}

		case "MARKING_LEFT_END", "MARKING_RIGHT_END", "REF_LINE_MARKING",
			"BENDING_TABLE", "GENERAL_MARKING_DATA", "CURVE_DATA", "CURVE_3D_DATA":
			// Present in the mar.dll writer but absent from every sample;
			// capture verbatim rather than dropping.
			name := string(s.Bytes())
			rs, err := readRawSection(s, name)
			if err != nil {
				return nil, err
			}
			if cur != nil {
				cur.RawSections = append(cur.RawSections, rs)
			}
		}
	}

	doc.Profiles = profiles

	return doc, nil
}

type Profile struct {
	Shape              string
	Dimension          string
	Quality            string
	Name               string
	PosNo              string
	MLength            float64
	MLengthManual      float64
	TLength            float64
	TLengthManual      float64
	EndPointDist       float64
	EndPointDistProd   float64
	TraceLength        float64
	TraceLengthProd    float64
	TwistAngle         float64
	ShipNo             string
	ShipName           string
	ProfSide           string
	SideUp             string
	Direction          string
	DirectionSign      string
	DirectionTrace     string
	BevelTrace         int
	BevelCode          int
	BevelName          string
	BevelType          int
	BevelVariant       int
	E                  float64
	Gap                float64
	ChamferTS          float64
	ChamferOS          float64
	Alpha              float64
	Beta               float64
	Nose               float64
	H                  float64
	HFact              float64
	HFactAdjust        float64
	AngleTS            float64
	AngleOS            float64
	DepthTS            float64
	DepthOS            float64
	ChamferWidthTS     float64
	ChamferWidthOS     float64
	ChamferHeightTS    float64
	ChamferHeightOS    float64
	Grinding           int
	Assembly           string
	AssemblyHigh       string
	AssemblyTotal      string
	SurfaceTreatment   string
	BlockNo            string
	BlockName          string
	Mirror             string
	Form               string
	BendDirection      string
	LeftEnd            *End
	RightEnd           *End
	Holes              []*Hole
	Quantity           int
	GeneratedAs        string
	Symmetrical        string
	ApproxPartWeight   float64
	IdentString        string
	PartnameLong       string
	PartnameShort      string
	DirectionFace      string
	Chamfer            float64
	AssemblySB         string
	AssemblyTotalSB    string
	AssemblySequence   string
	Baugruppe          string
	Zeichnungsno       string
	Zaehlno            string
	NoOfRefLines       int
	NoOfBendingTables  int
	TableType          string
	AdditionalInfo1    string
	AdditionalInfo2    string
	AdditionalInfo3    string
	AdditionalInfo4    string
	AdditionalInfo5    string
	AdditionalInfo6    string
	AdditionalInfo7    string
	AdditionalInfo8    string
	AdditionalInfo9    string
	AdditionalInfo10   string
	AdditionalInfo11   string
	AdditionalInfo12   string
	MarkingLeftExcist  int
	MarkingRightExcist int
	NoOfMacs           int
	NoOfMarks          int
	KnuckleUStart      float64
	KnuckleVStart      float64
	KnuckleUEnd        float64
	KnuckleVEnd        float64
	KnuckleInclination float64
	GapAdjustLeft      float64
	GapAdjustRight     float64
	Case               string
	RawSections        []RawSection
	CustData1          string
	CustData2          string
	CustData3          string
	CustData4          string
	CustText1          string
	CustText2          string
	CustText3          string
	CustText4          string
	CustText5          string
	CustText6          string
	CustText7          string
	CustText8          string
	CustText9          string
	CustText10         string
	PlanningUnit       string
	WorkingLocation    string
	BuildingStrategy   string
}

func readProfile(s *bufio.Scanner) (*Profile, error) {
	profile := &Profile{}

next:
	for s.Scan() {
		key, value, found := nextKeyValue(s)

		switch key {
		case "LEFT_END":
			end, err := readEnd(s)
			if err != nil {
				return nil, err
			}
			profile.LeftEnd = end
			continue next

		case "RIGHT_END":
			end, err := readEnd(s)
			if err != nil {
				return nil, err
			}
			profile.RightEnd = end
			continue next

		case "HOLES_NOTCHES_CUTOUTS":
			hole, err := parseHole(s)
			if err != nil {
				return nil, err
			}
			profile.Holes = append(profile.Holes, hole)
			continue next

		case "CONNECTION_TRACE":
			for s.Scan() {
				if bytes.Equal(s.Bytes(), []byte("END_OF_CONNECTION_TRACE")) {
					continue next
				}
			}

		case "GEOMETRY_DATA":
			for s.Scan() {
				if bytes.Equal(s.Bytes(), []byte("END_OF_GEOMETRY_DATA")) {
					continue next
				}
			}

		case "STRING_DATA":
			for s.Scan() {
				if bytes.Equal(s.Bytes(), []byte("END_OF_STRING_DATA")) {
					continue next
				}
			}
		}

		if !found {
			break
		}

		switch key {
		case "NAME":
			profile.Name = string(value)

		case "POSNO":
			profile.PosNo = string(value)

		case "MLENGTH":
			mLength, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.MLength = mLength

		case "MLENGTH_MANUAL":
			mLengthManual, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.MLengthManual = mLengthManual

		case "TLENGTH":
			tLength, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.TLength = tLength

		case "TLENGTH_MANUAL":
			tLengthManual, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.TLengthManual = tLengthManual

		case "END_POINT_DIST":
			endPointDist, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.EndPointDist = endPointDist

		case "END_POINT_DIST_PROD":
			endPointDistProd, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.EndPointDistProd = endPointDistProd

		case "TRACE_LENGTH":
			traceLength, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.TraceLength = traceLength

		case "TRACE_LENGTH_PROD":
			traceLengthProd, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.TraceLengthProd = traceLengthProd

		case "TWIST_ANGLE":
			twistAngle, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.TwistAngle = twistAngle

		case "SHIP_NO":
			profile.ShipNo = string(value)

		case "SHIP_NAME":
			profile.ShipName = string(value)

		case "PROF_SIDE":
			profile.ProfSide = string(value)

		case "SIDE_UP":
			profile.SideUp = string(value)

		case "DIRECTION":
			profile.Direction = string(value)

		case "DIRECTION_SIGN":
			profile.DirectionSign = string(value)

		case "DIRECTION_TRACE":
			profile.DirectionTrace = string(value)

		case "BEVEL_TRACE":
			bevelTrace, err := atoi(value)
			if err != nil {
				return nil, err
			}
			profile.BevelTrace = bevelTrace

		case "BEVEL_CODE":
			bevelCode, err := atoi(value)
			if err != nil {
				return nil, err
			}
			profile.BevelCode = bevelCode

		case "BEVEL_NAME":
			profile.BevelName = string(value)

		case "BEVEL_TYPE":
			bevelType, err := atoi(value)
			if err != nil {
				return nil, err
			}
			profile.BevelType = bevelType

		case "BEVEL_VARIANT":
			bevelVariant, err := atoi(value)
			if err != nil {
				return nil, err
			}
			profile.BevelVariant = bevelVariant

		case "E":
			e, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.E = e

		case "GAP":
			gap, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.Gap = gap

		case "CHAMFER_TS":
			chamferTs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferTS = chamferTs

		case "CHAMFER_OS":
			chamferOs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferOS = chamferOs

		case "ALPHA":
			alpha, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.Alpha = alpha

		case "BETA":
			beta, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.Beta = beta

		case "NOSE":
			nose, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.Nose = nose

		case "H":
			h, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.H = h

		case "H_FACT":
			hFact, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.HFact = hFact

		case "H_FACT_ADJUST":
			hFactAdjust, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.HFactAdjust = hFactAdjust

		case "ANGLE_TS":
			angleTs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.AngleTS = angleTs

		case "ANGLE_OS":
			angleOs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.AngleOS = angleOs

		case "DEPTH_TS":
			depthTs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.DepthTS = depthTs

		case "DEPTH_OS":
			depthOs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.DepthOS = depthOs

		case "CHAMFER_WIDTH_TS":
			chamferWidthTs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferWidthTS = chamferWidthTs

		case "CHAMFER_WIDTH_OS":
			chamferWidthOs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferWidthOS = chamferWidthOs

		case "CHAMFER_HEIGHT_TS":
			chamferHeightTs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferHeightTS = chamferHeightTs

		case "CHAMFER_HEIGHT_OS":
			chamferHeightOs, err := atof(value)
			if err != nil {
				return nil, err
			}
			profile.ChamferHeightOS = chamferHeightOs

		case "GRINDING":
			grinding, err := atoi(value)
			if err != nil {
				return nil, err
			}
			profile.Grinding = grinding

		case "ASSEMBLY":
			profile.Assembly = string(value)

		case "ASSEMBLY_HIGH":
			profile.AssemblyHigh = string(value)

		case "ASSEMBLY_TOT":
			profile.AssemblyTotal = string(value)

		case "SURFACE_TREATMENT":
			profile.SurfaceTreatment = string(value)

		case "BLOCK_NO":
			profile.BlockNo = string(value)

		case "BLOCK_NAME":
			profile.BlockName = string(value)

		case "MIRROR":
			profile.Mirror = string(value)

		case "FORM":
			profile.Form = string(value)

		case "BEND_DIRECTION":
			profile.BendDirection = string(value)

		case "GENERATED_AS":
			profile.GeneratedAs = string(value)

		case "SYMMETRICAL":
			profile.Symmetrical = string(value)

		case "APPROX_PART_WEIGHT":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.ApproxPartWeight = v

		case "IDENT_STRING":
			profile.IdentString = string(value)

		case "PARTNAME_LONG":
			profile.PartnameLong = string(value)

		case "PARTNAME_SHORT":
			profile.PartnameShort = string(value)

		case "DIRECTION_FACE":
			profile.DirectionFace = string(value)

		case "CHAMFER":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.Chamfer = v

		case "ASSEMBLY_SB":
			profile.AssemblySB = string(value)

		case "ASSEMBLY_TOT_SB":
			profile.AssemblyTotalSB = string(value)

		case "ASSEMBLY_SEQUENCE":
			profile.AssemblySequence = string(value)

		case "BAUGRUPPE":
			profile.Baugruppe = string(value)

		case "ZEICHNUNGSNO":
			profile.Zeichnungsno = string(value)

		case "ZAEHLNO":
			profile.Zaehlno = string(value)

		case "NO_OF_REF_LINES":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.NoOfRefLines = v

		case "NO_OF_BENDING_TABLES":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.NoOfBendingTables = v

		case "TABLE_TYPE":
			profile.TableType = string(value)

		case "ADDITIONAL_INFO1":
			profile.AdditionalInfo1 = string(value)

		case "ADDITIONAL_INFO2":
			profile.AdditionalInfo2 = string(value)

		case "ADDITIONAL_INFO3":
			profile.AdditionalInfo3 = string(value)

		case "ADDITIONAL_INFO4":
			profile.AdditionalInfo4 = string(value)

		case "ADDITIONAL_INFO5":
			profile.AdditionalInfo5 = string(value)

		case "ADDITIONAL_INFO6":
			profile.AdditionalInfo6 = string(value)

		case "ADDITIONAL_INFO7":
			profile.AdditionalInfo7 = string(value)

		case "ADDITIONAL_INFO8":
			profile.AdditionalInfo8 = string(value)

		case "ADDITIONAL_INFO9":
			profile.AdditionalInfo9 = string(value)

		case "ADDITIONAL_INFO10":
			profile.AdditionalInfo10 = string(value)

		case "ADDITIONAL_INFO11":
			profile.AdditionalInfo11 = string(value)

		case "ADDITIONAL_INFO12":
			profile.AdditionalInfo12 = string(value)

		case "MARKING_LEFT_EXCIST":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.MarkingLeftExcist = v

		case "MARKING_RIGHT_EXCIST":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.MarkingRightExcist = v

		case "NO_OF_MACS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.NoOfMacs = v

		case "NO_OF_MARKS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			profile.NoOfMarks = v

		case "KNUCKLE_U_START":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.KnuckleUStart = v

		case "KNUCKLE_V_START":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.KnuckleVStart = v

		case "KNUCKLE_U_END":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.KnuckleUEnd = v

		case "KNUCKLE_V_END":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.KnuckleVEnd = v

		case "KNUCKLE_INCLINATION":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.KnuckleInclination = v

		case "GAP_ADJUST_LEFT":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.GapAdjustLeft = v

		case "GAP_ADJUST_RIGHT":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			profile.GapAdjustRight = v

		case "CASE":
			profile.Case = string(value)

		case "CUST_DATA_1":
			profile.CustData1 = string(value)

		case "CUST_DATA_2":
			profile.CustData2 = string(value)

		case "CUST_DATA_3":
			profile.CustData3 = string(value)

		case "CUST_DATA_4":
			profile.CustData4 = string(value)

		case "CUST_TEXT_1":
			profile.CustText1 = string(value)

		case "CUST_TEXT_2":
			profile.CustText2 = string(value)

		case "CUST_TEXT_3":
			profile.CustText3 = string(value)

		case "CUST_TEXT_4":
			profile.CustText4 = string(value)

		case "CUST_TEXT_5":
			profile.CustText5 = string(value)

		case "CUST_TEXT_6":
			profile.CustText6 = string(value)

		case "CUST_TEXT_7":
			profile.CustText7 = string(value)

		case "CUST_TEXT_8":
			profile.CustText8 = string(value)

		case "CUST_TEXT_9":
			profile.CustText9 = string(value)

		case "CUST_TEXT_10":
			profile.CustText10 = string(value)

		case "PLANNING_UNIT":
			profile.PlanningUnit = string(value)

		case "WORKING_LOCATION":
			profile.WorkingLocation = string(value)

		case "BUILDING_STRATEGY":
			profile.BuildingStrategy = string(value)
		}
	}

	profile.Quantity = 1
	return profile, nil
}
