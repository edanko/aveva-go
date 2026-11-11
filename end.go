package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type End struct {
	LeftOrigin              float64
	LeftClosestPoint        float64
	LeftFarthestPoint       float64
	LeftV0                  float64
	LeftClosestPointWeb     float64
	LeftFarthestPointWeb    float64
	BurnEcut                int
	Stoss                   int
	FabricationExcess       float64
	EndcutType              int
	EndcutMask              string
	EndcutCode              int
	A                       float64
	B                       float64
	C                       float64
	R1                      float64
	R2                      float64
	V1                      float64
	V2                      float64
	V3                      float64
	V4                      float64
	Ks                      float64
	Excess                  float64
	BevelDefined            int
	BevelCode               int
	BevelName               string
	BevelType               int
	BevelVariant            int
	E                       float64
	Gap                     float64
	Chamfer                 float64
	ChamferTS               float64
	ChamferOS               float64
	Alpha                   float64
	Beta                    float64
	Nose                    float64
	H                       float64
	HFact                   float64
	HFactAdjust             float64
	AngleTS                 float64
	AngleOS                 float64
	DepthTS                 float64
	DepthOS                 float64
	ChamferWidthTS          float64
	ChamferWidthOS          float64
	ChamferHeightTS         float64
	ChamferHeightOS         float64
	ConnectionAngle         float64
	WebSeg                  float64
	BevelDefinedFlange      float64
	BevelCodeFlange         float64
	BevelNameFlange         string
	BevelTypeFlange         float64
	BevelVariantFlange      float64
	EFlange                 float64
	GapFlange               float64
	ChamferFlange           float64
	ChamferTSFlange         float64
	ChamferOSFlange         float64
	AlphaFlange             float64
	BetaFlange              float64
	NoseFlange              float64
	HFlange                 float64
	HFactFlange             float64
	HFactAdjustFlange       float64
	AngleTSFlange           float64
	AngleOSFlange           float64
	DepthTSFlange           float64
	DepthOSFlange           float64
	ChamferWidthTSFlange    float64
	ChamferWidthOSFlange    float64
	ChamferHeightTSFlange   float64
	ChamferHeightOSFlange   float64
	ConnectionAngleFlange   float64
	FlaSeg                  float64
	BevelDefinedFlange2     float64
	BevelCodeFlange2        float64
	BevelNameFlange2        string
	BevelTypeFlange2        float64
	BevelVariantFlange2     float64
	EFlange2                float64
	GapFlange2              float64
	ChamferFlange2          float64
	ChamferTSFlange2        float64
	ChamferOSFlange2        float64
	AlphaFlange2            float64
	BetaFlange2             float64
	NoseFlange2             float64
	HFlange2                float64
	HFactFlange2            float64
	HFactAdjustFlange2      float64
	AngleTSFlange2          float64
	AngleOSFlange2          float64
	DepthTSFlange2          float64
	DepthOSFlange2          float64
	ChamferWidthTSFlange2   float64
	ChamferWidthOSFlange2   float64
	ChamferHeightTSFlange2  float64
	ChamferHeightOSFlange2  float64
	Fla2Seg                 float64
	Gsd                     float64
	GsdDist                 float64
	Contour                 *Contour
	FlangeContour           *Contour
	EndcutCodeO             string
	EndcutName              string
	Ampl                    float64
	EndcutUserDefinedName   string
	RadFlange               float64
	RadBFlange              float64
	NoseAngleFlange         float64
	NoseLenAFlange          float64
	NoseLenBFlange          float64
	RadFlange2              float64
	RadBFlange2             float64
	NoseAngleFlange2        float64
	NoseLenAFlange2         float64
	NoseLenBFlange2         float64
	KnuckleHeight           float64
	KnuckleRatio            float64
	LengthFromEdge          float64
	ExtraLength             float64
	SnipLength              float64
	KnuckleGap              float64
	FlangeBevelAngle        float64
	KnuckleAngle            float64
	FlangeExtension         float64
	EndcutUserDefinedParams [20]float64 // index 0 == PARAM1
	EndcutUserDefinedData   [20]string  // index 0 == DATA1
}

func readEnd(s *bufio.Scanner) (*End, error) {
	end := &End{}

	bevelNamesFound := 0

next:
	for s.Scan() {
		key, value, found := strings.Cut(s.Text(), "=")

		switch key {
		case "START_OF_CONTOUR":
			contour, err := readContour(s)
			if err != nil {
				return nil, err
			}
			end.Contour = contour
			continue next

		case "START_OF_FCONTOUR":
			contour, err := readContour(s)
			if err != nil {
				return nil, err
			}
			end.FlangeContour = contour
			continue next
		}

		if !found {
			break
		}

		switch key {
		case "LEFT_ORIGIN":
			leftOrigin, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftOrigin = leftOrigin

		case "LEFT_CLOSEST_POINT":
			leftClosestPoint, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftClosestPoint = leftClosestPoint

		case "LEFT_FARTHEST_POINT":
			leftFarthestPoint, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftFarthestPoint = leftFarthestPoint

		case "LEFT_V0":
			leftV0, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftV0 = leftV0

		case "LEFT_CLOSEST_POINT_WEB":
			leftClosestPointWeb, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftClosestPointWeb = leftClosestPointWeb

		case "LEFT_FARTHEST_POINT_WEB":
			leftFarthestPointWeb, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LeftFarthestPointWeb = leftFarthestPointWeb

		case "BURN_ECUT":
			burnEcut, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.BurnEcut = burnEcut

		case "STOSS":
			stoss, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.Stoss = stoss

		case "FABRICATION_EXCESS":
			fabricationExcess, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.FabricationExcess = fabricationExcess

		case "ENDCUT_TYPE":
			endcutType, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.EndcutType = endcutType

		case "ENDCUT_MASK":
			end.EndcutMask = value

		case "ENDCUT_CODE":
			endcutCode, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.EndcutCode = endcutCode

		case "A":
			a, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.A = a
		case "B":
			b, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.B = b
		case "C":
			c, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.C = c
		case "R1":
			r1, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.R1 = r1
		case "R2":
			r2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.R2 = r2
		case "V1":
			v1, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.V1 = v1
		case "V2":
			v2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.V2 = v2
		case "V3":
			v3, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.V3 = v3
		case "V4":
			v4, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.V4 = v4
		case "KS":
			ks, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Ks = ks

		case "EXCESS":
			excess, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Excess = excess

		case "BEVEL_DEFINED":
			bevelDefined, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.BevelDefined = bevelDefined

		case "BEVEL_CODE":
			bevelCode, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.BevelCode = bevelCode

		case "BEVEL_TYPE":
			bevelType, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.BevelType = bevelType

		case "BEVEL_VARIANT":
			bevelVariant, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			end.BevelVariant = bevelVariant

		case "E":
			eValue, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.E = eValue

		case "GAP":
			gap, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Gap = gap

		case "CHAMFER":
			chamfer, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Chamfer = chamfer

		case "CHAMFER_TS":
			chamferTs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferTS = chamferTs
		case "CHAMFER_OS":
			chamferOs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferOS = chamferOs

		case "ALPHA":
			alpha, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Alpha = alpha

		case "BETA":
			beta, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Beta = beta

		case "NOSE":
			nose, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Nose = nose

		case "H":
			h, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.H = h

		case "H_FACT":
			hFact, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFact = hFact

		case "H_FACT_ADJUST":
			hFactAdjust, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFactAdjust = hFactAdjust

		case "ANGLE_TS":
			angleTs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleTS = angleTs

		case "ANGLE_OS":
			angleOs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleOS = angleOs

		case "DEPTH_TS":
			depthTs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthTS = depthTs

		case "DEPTH_OS":
			depthOs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthOS = depthOs

		case "CHAMFER_WIDTH_TS":
			chamferWidthTs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthTS = chamferWidthTs

		case "CHAMFER_WIDTH_OS":
			chamferWidthOs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthOS = chamferWidthOs

		case "CHAMFER_HEIGHT_TS":
			chamferHeightTs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightTS = chamferHeightTs

		case "CHAMFER_HEIGHT_OS":
			chamferHeightOs, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightOS = chamferHeightOs

		case "CONNECTION_ANGLE":
			connectionAngle, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ConnectionAngle = connectionAngle

		case "WEB_SEG":
			webSeg, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.WebSeg = webSeg

		case "BEVEL_DEFINED_FLANGE":
			bevelDefinedFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelDefinedFlange = bevelDefinedFlange

		case "BEVEL_CODE_FLANGE":
			bevelCodeFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelCodeFlange = bevelCodeFlange

		case "BEVEL_NAME":
			switch bevelNamesFound {
			case 0:
				end.BevelName = value
			case 1:
				end.BevelNameFlange = value
			case 2:
				end.BevelNameFlange2 = value
			}

			bevelNamesFound++

		case "BEVEL_TYPE_FLANGE":
			bevelTypeFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelTypeFlange = bevelTypeFlange

		case "BEVEL_VARIANT_FLANGE":
			bevelVariantFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelVariantFlange = bevelVariantFlange

		case "E_FLANGE":
			eFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EFlange = eFlange

		case "GAP_FLANGE":
			gapFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.GapFlange = gapFlange

		case "CHAMFER_FLANGE":
			chamferFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferFlange = chamferFlange

		case "CHAMFER_TS_FLANGE":
			chamferTsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferTSFlange = chamferTsFlange
		case "CHAMFER_OS_FLANGE":
			chamferOsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferOSFlange = chamferOsFlange

		case "ALPHA_FLANGE":
			alphaFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AlphaFlange = alphaFlange

		case "BETA_FLANGE":
			betaFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BetaFlange = betaFlange

		case "NOSE_FLANGE":
			noseFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseFlange = noseFlange

		case "H_FLANGE":
			hFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFlange = hFlange

		case "H_FACT_FLANGE":
			hFactFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFactFlange = hFactFlange

		case "H_FACT_ADJUST_FLANGE":
			hFactAdjustFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFactAdjustFlange = hFactAdjustFlange

		case "ANGLE_TS_FLANGE":
			angleTsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleTSFlange = angleTsFlange

		case "ANGLE_OS_FLANGE":
			angleOsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleOSFlange = angleOsFlange

		case "DEPTH_TS_FLANGE":
			depthTsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthTSFlange = depthTsFlange

		case "DEPTH_OS_FLANGE":
			depthOsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthOSFlange = depthOsFlange

		case "CHAMFER_WIDTH_TS_FLANGE":
			chamferWidthTsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthTSFlange = chamferWidthTsFlange

		case "CHAMFER_WIDTH_OS_FLANGE":
			chamferWidthOsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthOSFlange = chamferWidthOsFlange

		case "CHAMFER_HEIGHT_TS_FLANGE":
			chamferHeightTsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightTSFlange = chamferHeightTsFlange

		case "CHAMFER_HEIGHT_OS_FLANGE":
			chamferHeightOsFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightOSFlange = chamferHeightOsFlange

		case "CONNECTION_ANGLE_FLANGE":
			connectionAngleFlange, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ConnectionAngleFlange = connectionAngleFlange

		case "FLA_SEG":
			flaSeg, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.FlaSeg = flaSeg

		case "BEVEL_DEFINED_FLANGE2":
			bevelDefinedFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelDefinedFlange2 = bevelDefinedFlange2

		case "BEVEL_CODE_FLANGE2":
			bevelCodeFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelCodeFlange2 = bevelCodeFlange2

		case "BEVEL_TYPE_FLANGE2":
			bevelTypeFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelTypeFlange2 = bevelTypeFlange2

		case "BEVEL_VARIANT_FLANGE2":
			bevelVariantFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BevelVariantFlange2 = bevelVariantFlange2

		case "E_FLANGE2":
			eFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EFlange2 = eFlange2

		case "GAP_FLANGE2":
			gapFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.GapFlange2 = gapFlange2

		case "CHAMFER_FLANGE2":
			chamferFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferFlange2 = chamferFlange2

		case "CHAMFER_TS_FLANGE2":
			chamferTsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferTSFlange2 = chamferTsFlange2

		case "CHAMFER_OS_FLANGE2":
			chamferOsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferOSFlange2 = chamferOsFlange2

		case "ALPHA_FLANGE2":
			alphaFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AlphaFlange2 = alphaFlange2

		case "BETA_FLANGE2":
			betaFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.BetaFlange2 = betaFlange2

		case "NOSE_FLANGE2":
			noseFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseFlange2 = noseFlange2

		case "H_FLANGE2":
			hFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFlange2 = hFlange2

		case "H_FACT_FLANGE2":
			hFactFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFactFlange2 = hFactFlange2

		case "H_FACT_ADJUST_FLANGE2":
			hFactAdjustFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.HFactAdjustFlange2 = hFactAdjustFlange2

		case "ANGLE_TS_FLANGE2":
			angleTsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleTSFlange2 = angleTsFlange2

		case "ANGLE_OS_FLANGE2":
			angleOsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.AngleOSFlange2 = angleOsFlange2

		case "DEPTH_TS_FLANGE2":
			depthTsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthTSFlange2 = depthTsFlange2

		case "DEPTH_OS_FLANGE2":
			depthOsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.DepthOSFlange2 = depthOsFlange2

		case "CHAMFER_WIDTH_TS_FLANGE2":
			chamferWidthTsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthTSFlange2 = chamferWidthTsFlange2

		case "CHAMFER_WIDTH_OS_FLANGE2":
			chamferWidthOsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferWidthOSFlange2 = chamferWidthOsFlange2

		case "CHAMFER_HEIGHT_TS_FLANGE2":
			chamferHeightTsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightTSFlange2 = chamferHeightTsFlange2

		case "CHAMFER_HEIGHT_OS_FLANGE2":
			chamferHeightOsFlange2, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ChamferHeightOSFlange2 = chamferHeightOsFlange2

		case "FLA2_SEG":
			fla2Seg, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Fla2Seg = fla2Seg

		case "GSD":
			gsd, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Gsd = gsd

		case "GSD_DIST":
			gsdDist, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.GsdDist = gsdDist

		case "ENDCUT_CODEO":
			end.EndcutCodeO = value

		case "ENDCUT_NAME":
			end.EndcutName = value

		case "AMPL":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.Ampl = v

		case "ENDCUT_USER_DEFINED_NAME":
			end.EndcutUserDefinedName = value

		case "RAD_FLANGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.RadFlange = v

		case "RADB_FLANGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.RadBFlange = v

		case "NOSE_ANGLE_FLANGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseAngleFlange = v

		case "NOSE_LEN_A_FLANGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseLenAFlange = v

		case "NOSE_LEN_B_FLANGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseLenBFlange = v

		case "RAD_FLANGE2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.RadFlange2 = v

		case "RADB_FLANGE2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.RadBFlange2 = v

		case "NOSE_ANGLE_FLANGE2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseAngleFlange2 = v

		case "NOSE_LEN_A_FLANGE2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseLenAFlange2 = v

		case "NOSE_LEN_B_FLANGE2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.NoseLenBFlange2 = v

		case "KNUCKLE_HEIGHT":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.KnuckleHeight = v

		case "KNUCKLE_RATIO":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.KnuckleRatio = v

		case "LENGTH_FROM_EDGE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.LengthFromEdge = v

		case "EXTRA_LENGTH":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.ExtraLength = v

		case "SNIP_LENGTH":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.SnipLength = v

		case "KNUCKLE_GAP":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.KnuckleGap = v

		case "FLANGE_BEVEL_ANGLE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.FlangeBevelAngle = v

		case "KNUCKLE_ANGLE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.KnuckleAngle = v

		case "FLANGE_EXTENSION":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.FlangeExtension = v

		case "ENDCUT_USER_DEFINED_PARAM1":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[0] = v

		case "ENDCUT_USER_DEFINED_PARAM2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[1] = v

		case "ENDCUT_USER_DEFINED_PARAM3":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[2] = v

		case "ENDCUT_USER_DEFINED_PARAM4":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[3] = v

		case "ENDCUT_USER_DEFINED_PARAM5":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[4] = v

		case "ENDCUT_USER_DEFINED_PARAM6":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[5] = v

		case "ENDCUT_USER_DEFINED_PARAM7":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[6] = v

		case "ENDCUT_USER_DEFINED_PARAM8":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[7] = v

		case "ENDCUT_USER_DEFINED_PARAM9":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[8] = v

		case "ENDCUT_USER_DEFINED_PARAM10":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[9] = v

		case "ENDCUT_USER_DEFINED_PARAM11":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[10] = v

		case "ENDCUT_USER_DEFINED_PARAM12":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[11] = v

		case "ENDCUT_USER_DEFINED_PARAM13":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[12] = v

		case "ENDCUT_USER_DEFINED_PARAM14":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[13] = v

		case "ENDCUT_USER_DEFINED_PARAM15":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[14] = v

		case "ENDCUT_USER_DEFINED_PARAM16":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[15] = v

		case "ENDCUT_USER_DEFINED_PARAM17":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[16] = v

		case "ENDCUT_USER_DEFINED_PARAM18":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[17] = v

		case "ENDCUT_USER_DEFINED_PARAM19":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[18] = v

		case "ENDCUT_USER_DEFINED_PARAM20":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			end.EndcutUserDefinedParams[19] = v

		case "ENDCUT_USER_DEFINED_DATA1":
			end.EndcutUserDefinedData[0] = value

		case "ENDCUT_USER_DEFINED_DATA2":
			end.EndcutUserDefinedData[1] = value

		case "ENDCUT_USER_DEFINED_DATA3":
			end.EndcutUserDefinedData[2] = value

		case "ENDCUT_USER_DEFINED_DATA4":
			end.EndcutUserDefinedData[3] = value

		case "ENDCUT_USER_DEFINED_DATA5":
			end.EndcutUserDefinedData[4] = value

		case "ENDCUT_USER_DEFINED_DATA6":
			end.EndcutUserDefinedData[5] = value

		case "ENDCUT_USER_DEFINED_DATA7":
			end.EndcutUserDefinedData[6] = value

		case "ENDCUT_USER_DEFINED_DATA8":
			end.EndcutUserDefinedData[7] = value

		case "ENDCUT_USER_DEFINED_DATA9":
			end.EndcutUserDefinedData[8] = value

		case "ENDCUT_USER_DEFINED_DATA10":
			end.EndcutUserDefinedData[9] = value

		case "ENDCUT_USER_DEFINED_DATA11":
			end.EndcutUserDefinedData[10] = value

		case "ENDCUT_USER_DEFINED_DATA12":
			end.EndcutUserDefinedData[11] = value

		case "ENDCUT_USER_DEFINED_DATA13":
			end.EndcutUserDefinedData[12] = value

		case "ENDCUT_USER_DEFINED_DATA14":
			end.EndcutUserDefinedData[13] = value

		case "ENDCUT_USER_DEFINED_DATA15":
			end.EndcutUserDefinedData[14] = value

		case "ENDCUT_USER_DEFINED_DATA16":
			end.EndcutUserDefinedData[15] = value

		case "ENDCUT_USER_DEFINED_DATA17":
			end.EndcutUserDefinedData[16] = value

		case "ENDCUT_USER_DEFINED_DATA18":
			end.EndcutUserDefinedData[17] = value

		case "ENDCUT_USER_DEFINED_DATA19":
			end.EndcutUserDefinedData[18] = value

		case "ENDCUT_USER_DEFINED_DATA20":
			end.EndcutUserDefinedData[19] = value
		}
	}
	return end, nil
}
