package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type End struct {
	LeftOrigin             float64
	LeftClosestPoint       float64
	LeftFarthestPoint      float64
	LeftV0                 float64
	LeftClosestPointWeb    float64
	LeftFarthestPointWeb   float64
	BurnEcut               int
	Stoss                  int
	FabricationExcess      float64
	EndcutType             int
	EndcutMask             string
	EndcutCode             int
	A                      float64
	B                      float64
	C                      float64
	R1                     float64
	R2                     float64
	V1                     float64
	V2                     float64
	V3                     float64
	V4                     float64
	Ks                     float64
	Excess                 float64
	BevelDefined           int
	BevelCode              int
	BevelName              string
	BevelType              int
	BevelVariant           int
	E                      float64
	Gap                    float64
	Chamfer                float64
	ChamferTs              float64
	ChamferOs              float64
	Alpha                  float64
	Beta                   float64
	Nose                   float64
	H                      float64
	HFact                  float64
	HFactAdjust            float64
	AngleTs                float64
	AngleOs                float64
	DepthTs                float64
	DepthOs                float64
	ChamferWidthTs         float64
	ChamferWidthOs         float64
	ChamferHeightTs        float64
	ChamferHeightOs        float64
	ConnectionAngle        float64
	WebSeg                 float64
	BevelDefinedFlange     float64
	BevelCodeFlange        float64
	BevelNameFlange        string
	BevelTypeFlange        float64
	BevelVariantFlange     float64
	EFlange                float64
	GapFlange              float64
	ChamferFlange          float64
	ChamferTsFlange        float64
	ChamferOsFlange        float64
	AlphaFlange            float64
	BetaFlange             float64
	NoseFlange             float64
	HFlange                float64
	HFactFlange            float64
	HFactAdjustFlange      float64
	AngleTsFlange          float64
	AngleOsFlange          float64
	DepthTsFlange          float64
	DepthOsFlange          float64
	ChamferWidthTsFlange   float64
	ChamferWidthOsFlange   float64
	ChamferHeightTsFlange  float64
	ChamferHeightOsFlange  float64
	ConnectionAngleFlange  float64
	FlaSeg                 float64
	BevelDefinedFlange2    float64
	BevelCodeFlange2       float64
	BevelNameFlange2       string
	BevelTypeFlange2       float64
	BevelVariantFlange2    float64
	EFlange2               float64
	GapFlange2             float64
	ChamferFlange2         float64
	ChamferTsFlange2       float64
	ChamferOsFlange2       float64
	AlphaFlange2           float64
	BetaFlange2            float64
	NoseFlange2            float64
	HFlange2               float64
	HFactFlange2           float64
	HFactAdjustFlange2     float64
	AngleTsFlange2         float64
	AngleOsFlange2         float64
	DepthTsFlange2         float64
	DepthOsFlange2         float64
	ChamferWidthTsFlange2  float64
	ChamferWidthOsFlange2  float64
	ChamferHeightTsFlange2 float64
	ChamferHeightOsFlange2 float64
	Fla2Seg                float64
	Gsd                    float64
	GsdDist                float64
	Contour                Contour
	FContour               Contour
}

func readEnd(s *bufio.Scanner) End {
	var e End

	bevelNamesFound := 0

next:
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		switch k {
		case "START_OF_CONTOUR":
			e.Contour = readContour(s)
			continue next

		case "START_OF_FCONTOUR":
			e.FContour = readContour(s)
			continue next
		}

		if !ok {
			break
		}

		switch k {
		case "LEFT_ORIGIN":
			e.LeftOrigin, _ = strconv.ParseFloat(v, 64)

		case "LEFT_CLOSEST_POINT":
			e.LeftClosestPoint, _ = strconv.ParseFloat(v, 64)

		case "LEFT_FARTHEST_POINT":
			e.LeftFarthestPoint, _ = strconv.ParseFloat(v, 64)

		case "LEFT_V0":
			e.LeftV0, _ = strconv.ParseFloat(v, 64)

		case "LEFT_CLOSEST_POINT_WEB":
			e.LeftClosestPointWeb, _ = strconv.ParseFloat(v, 64)

		case "LEFT_FARTHEST_POINT_WEB":
			e.LeftFarthestPointWeb, _ = strconv.ParseFloat(v, 64)

		case "BURN_ECUT":
			e.BurnEcut, _ = strconv.Atoi(v)

		case "STOSS":
			e.Stoss, _ = strconv.Atoi(v)

		case "FABRICATION_EXCESS":
			e.FabricationExcess, _ = strconv.ParseFloat(v, 64)

		case "ENDCUT_TYPE":
			e.EndcutType, _ = strconv.Atoi(v)

		case "ENDCUT_MASK":
			e.EndcutMask = v

		case "ENDCUT_CODE":
			e.EndcutCode, _ = strconv.Atoi(v)

		case "A":
			e.A, _ = strconv.ParseFloat(v, 64)
		case "B":
			e.B, _ = strconv.ParseFloat(v, 64)
		case "C":
			e.C, _ = strconv.ParseFloat(v, 64)
		case "R1":
			e.R1, _ = strconv.ParseFloat(v, 64)
		case "R2":
			e.R2, _ = strconv.ParseFloat(v, 64)
		case "V1":
			e.V1, _ = strconv.ParseFloat(v, 64)
		case "V2":
			e.V2, _ = strconv.ParseFloat(v, 64)
		case "V3":
			e.V3, _ = strconv.ParseFloat(v, 64)
		case "V4":
			e.V4, _ = strconv.ParseFloat(v, 64)
		case "KS":
			e.Ks, _ = strconv.ParseFloat(v, 64)

		case "EXCESS":
			e.Excess, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_DEFINED":
			e.BevelDefined, _ = strconv.Atoi(v)

		case "BEVEL_CODE":
			e.BevelCode, _ = strconv.Atoi(v)

		case "BEVEL_TYPE":
			e.BevelType, _ = strconv.Atoi(v)

		case "BEVEL_VARIANT":
			e.BevelVariant, _ = strconv.Atoi(v)

		case "E":
			e.E, _ = strconv.ParseFloat(v, 64)

		case "GAP":
			e.Gap, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER":
			e.Chamfer, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_TS":
			e.ChamferTs, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_OS":
			e.ChamferOs, _ = strconv.ParseFloat(v, 64)

		case "ALPHA":
			e.Alpha, _ = strconv.ParseFloat(v, 64)

		case "BETA":
			e.Beta, _ = strconv.ParseFloat(v, 64)

		case "NOSE":
			e.Nose, _ = strconv.ParseFloat(v, 64)

		case "H":
			e.H, _ = strconv.ParseFloat(v, 64)

		case "H_FACT":
			e.HFact, _ = strconv.ParseFloat(v, 64)

		case "H_FACT_ADJUST":
			e.HFactAdjust, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_TS":
			e.AngleTs, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_OS":
			e.AngleOs, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_TS":
			e.DepthTs, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_OS":
			e.DepthOs, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_TS":
			e.ChamferWidthTs, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_OS":
			e.ChamferWidthOs, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_TS":
			e.ChamferHeightTs, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_OS":
			e.ChamferHeightOs, _ = strconv.ParseFloat(v, 64)

		case "CONNECTION_ANGLE":
			e.ConnectionAngle, _ = strconv.ParseFloat(v, 64)

		case "WEB_SEG":
			e.WebSeg, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_DEFINED_FLANGE":
			e.BevelDefinedFlange, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_CODE_FLANGE":
			e.BevelCodeFlange, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_NAME":
			switch bevelNamesFound {
			case 0:
				e.BevelName = v
			case 1:
				e.BevelNameFlange = v
			case 2:
				e.BevelNameFlange2 = v
			}

			bevelNamesFound++

		case "BEVEL_TYPE_FLANGE":
			e.BevelTypeFlange, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_VARIANT_FLANGE":
			e.BevelVariantFlange, _ = strconv.ParseFloat(v, 64)

		case "E_FLANGE":
			e.EFlange, _ = strconv.ParseFloat(v, 64)

		case "GAP_FLANGE":
			e.GapFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_FLANGE":
			e.ChamferFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_TS_FLANGE":
			e.ChamferTsFlange, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_OS_FLANGE":
			e.ChamferOsFlange, _ = strconv.ParseFloat(v, 64)

		case "ALPHA_FLANGE":
			e.AlphaFlange, _ = strconv.ParseFloat(v, 64)

		case "BETA_FLANGE":
			e.BetaFlange, _ = strconv.ParseFloat(v, 64)

		case "NOSE_FLANGE":
			e.NoseFlange, _ = strconv.ParseFloat(v, 64)

		case "H_FLANGE":
			e.HFlange, _ = strconv.ParseFloat(v, 64)

		case "H_FACT_FLANGE":
			e.HFactFlange, _ = strconv.ParseFloat(v, 64)

		case "H_FACT_ADJUST_FLANGE":
			e.HFactAdjustFlange, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_TS_FLANGE":
			e.AngleTsFlange, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_OS_FLANGE":
			e.AngleOsFlange, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_TS_FLANGE":
			e.DepthTsFlange, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_OS_FLANGE":
			e.DepthOsFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_TS_FLANGE":
			e.ChamferWidthTsFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_OS_FLANGE":
			e.ChamferWidthOsFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_TS_FLANGE":
			e.ChamferHeightTsFlange, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_OS_FLANGE":
			e.ChamferHeightOsFlange, _ = strconv.ParseFloat(v, 64)

		case "CONNECTION_ANGLE_FLANGE":
			e.ConnectionAngleFlange, _ = strconv.ParseFloat(v, 64)

		case "FLA_SEG":
			e.FlaSeg, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_DEFINED_FLANGE2":
			e.BevelDefinedFlange2, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_CODE_FLANGE2":
			e.BevelCodeFlange2, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_TYPE_FLANGE2":
			e.BevelTypeFlange2, _ = strconv.ParseFloat(v, 64)

		case "BEVEL_VARIANT_FLANGE2":
			e.BevelVariantFlange2, _ = strconv.ParseFloat(v, 64)

		case "E_FLANGE2":
			e.EFlange2, _ = strconv.ParseFloat(v, 64)

		case "GAP_FLANGE2":
			e.GapFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_FLANGE2":
			e.ChamferFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_TS_FLANGE2":
			e.ChamferTsFlange2, _ = strconv.ParseFloat(v, 64)
		case "CHAMFER_OS_FLANGE2":
			e.ChamferOsFlange2, _ = strconv.ParseFloat(v, 64)

		case "ALPHA_FLANGE2":
			e.AlphaFlange2, _ = strconv.ParseFloat(v, 64)

		case "BETA_FLANGE2":
			e.BetaFlange2, _ = strconv.ParseFloat(v, 64)

		case "NOSE_FLANGE2":
			e.NoseFlange2, _ = strconv.ParseFloat(v, 64)

		case "H_FLANGE2":
			e.HFlange2, _ = strconv.ParseFloat(v, 64)

		case "H_FACT_FLANGE2":
			e.HFactFlange2, _ = strconv.ParseFloat(v, 64)

		case "H_FACT_ADJUST_FLANGE2":
			e.HFactAdjustFlange2, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_TS_FLANGE2":
			e.AngleTsFlange2, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_OS_FLANGE2":
			e.AngleOsFlange2, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_TS_FLANGE2":
			e.DepthTsFlange2, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_OS_FLANGE2":
			e.DepthOsFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_TS_FLANGE2":
			e.ChamferWidthTsFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_OS_FLANGE2":
			e.ChamferWidthOsFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_TS_FLANGE2":
			e.ChamferHeightTsFlange2, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_OS_FLANGE2":
			e.ChamferHeightOsFlange2, _ = strconv.ParseFloat(v, 64)

		case "FLA2_SEG":
			e.Fla2Seg, _ = strconv.ParseFloat(v, 64)

		case "GSD":
			e.Gsd, _ = strconv.ParseFloat(v, 64)

		case "GSD_DIST":
			e.GsdDist, _ = strconv.ParseFloat(v, 64)
		}
	}
	return e
}
