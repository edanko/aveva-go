package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type ProfileData struct {
	Shape            string
	Dimension        string
	Quality          string
	Name             string
	PosNo            string
	Mlength          float64
	MlengthManual    float64
	Tlength          float64
	TlengthManual    float64
	EndPointDist     float64
	EndPointDistProd float64
	TraceLength      float64
	TraceLengthProd  float64
	TwistAngle       float64
	ShipNo           string
	ShipName         string
	ProfSide         string
	SideUp           string
	Direction        string
	DirectionSign    string
	DirectionTrace   string
	BevelTrace       int
	BevelCode        int
	BevelName        string
	BevelType        int
	BevelVariant     int
	E                float64
	Gap              float64
	ChamferTs        float64
	ChamferOs        float64
	Alpha            float64
	Beta             float64
	Nose             float64
	H                float64
	HFact            float64
	HFactAdjust      float64
	AngleTs          float64
	AngleOs          float64
	DepthTs          float64
	DepthOs          float64
	ChamferWidthTs   float64
	ChamferWidthOs   float64
	ChamferHeightTs  float64
	ChamferHeightOs  float64
	Grinding         int
	Milling          int
	Assembly         string
	AssemblyHigh     string
	AssemblyTot      string
	SurfaceTreatment string
	BlockNo          string
	BlockName        string
	Mirror           string
	Form             string
	BendDirection    string
	LeftEnd          End
	RightEnd         End
	Holes            []Hole
	Quantity         int
}

func readProfileData(s *bufio.Scanner) ProfileData {
	var p ProfileData

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		switch k {
		case "LEFT_END":
			p.LeftEnd = readEnd(s)
			continue next

		case "RIGHT_END":
			p.RightEnd = readEnd(s)
			continue next

		case "HOLES_NOTCHES_CUTOUTS":
			p.Holes = append(p.Holes, readHolesNotchesCutouts(s))
			continue next

		case "CONNECTION_TRACE":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_CONNECTION_TRACE":
					continue next
				}
			}

		case "GEOMETRY_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_GEOMETRY_DATA":
					continue next
				}
			}

		case "STRING_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_STRING_DATA":
					continue next
				}
			}
		}

		if !ok {
			break
		}

		switch k {
		case "NAME":
			p.Name = v

		case "POSNO":
			p.PosNo = v

		case "MLENGTH":
			p.Mlength, _ = strconv.ParseFloat(v, 64)
		case "MLENGTH_MANUAL":
			p.MlengthManual, _ = strconv.ParseFloat(v, 64)
		case "TLENGTH":
			p.Tlength, _ = strconv.ParseFloat(v, 64)
		case "TLENGTH_MANUAL":
			p.TlengthManual, _ = strconv.ParseFloat(v, 64)

		case "END_POINT_DIST":
			p.EndPointDist, _ = strconv.ParseFloat(v, 64)

		case "END_POINT_DIST_PROD":
			p.EndPointDistProd, _ = strconv.ParseFloat(v, 64)

		case "TRACE_LENGTH":
			p.TraceLength, _ = strconv.ParseFloat(v, 64)

		case "TRACE_LENGTH_PROD":
			p.TraceLengthProd, _ = strconv.ParseFloat(v, 64)

		case "TWIST_ANGLE":
			p.TwistAngle, _ = strconv.ParseFloat(v, 64)

		case "SHIP_NO":
			p.ShipNo = v
		case "SHIP_NAME":
			p.ShipName = v

		case "PROF_SIDE":
			p.ProfSide = v

		case "SIDE_UP":
			p.SideUp = v

		case "DIRECTION":
			p.Direction = v
		case "DIRECTION_SIGN":
			p.DirectionSign = v
		case "DIRECTION_TRACE":
			p.DirectionTrace = v

		case "BEVEL_TRACE":
			p.BevelTrace, _ = strconv.Atoi(v)
		case "BEVEL_CODE":
			p.BevelCode, _ = strconv.Atoi(v)
		case "BEVEL_NAME":
			p.BevelName = v
		case "BEVEL_TYPE":
			p.BevelType, _ = strconv.Atoi(v)
		case "BEVEL_VARIANT":
			p.BevelVariant, _ = strconv.Atoi(v)
		case "E":
			p.E, _ = strconv.ParseFloat(v, 64)
		case "GAP":
			p.Gap, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_TS":
			p.ChamferTs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_OS":
			p.ChamferOs, _ = strconv.ParseFloat(v, 64)
		case "ALPHA":
			p.Alpha, _ = strconv.ParseFloat(v, 64)
		case "BETA":
			p.Beta, _ = strconv.ParseFloat(v, 64)
		case "NOSE":
			p.Nose, _ = strconv.ParseFloat(v, 64)
		case "H":
			p.H, _ = strconv.ParseFloat(v, 64)
		case "H_FACT":
			p.HFact, _ = strconv.ParseFloat(v, 64)
		case "H_FACT_ADJUST":
			p.HFactAdjust, _ = strconv.ParseFloat(v, 64)
		case "ANGLE_TS":
			p.AngleTs, _ = strconv.ParseFloat(v, 64)
		case "ANGLE_OS":
			p.AngleOs, _ = strconv.ParseFloat(v, 64)
		case "DEPTH_TS":
			p.DepthTs, _ = strconv.ParseFloat(v, 64)
		case "DEPTH_OS":
			p.DepthOs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_WIDTH_TS":
			p.ChamferWidthTs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_WIDTH_OS":
			p.ChamferWidthOs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_HEIGHT_TS":
			p.ChamferHeightTs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_HEIGHT_OS":
			p.ChamferHeightOs, _ = strconv.ParseFloat(v, 64)

		case "GRINDING":
			p.Grinding, _ = strconv.Atoi(v)
		case "MILLING":
			p.Milling, _ = strconv.Atoi(v)

		case "ASSEMBLY":
			p.Assembly = v
		case "ASSEMBLY_HIGH":
			p.AssemblyHigh = v
		case "ASSEMBLY_TOT":
			p.AssemblyTot = v

		case "SURFACE_TREATMENT":
			p.BlockNo = v

		case "BLOCK_NO":
			p.BlockNo = v
		case "BLOCK_NAME":
			p.BlockName = v

		case "MIRROR":
			p.Mirror = v

		case "FORM":
			p.Form = v

		case "BEND_DIRECTION":
			p.BendDirection = v
		}
	}

	p.Quantity = 1
	return p
}
