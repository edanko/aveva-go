package gen

import (
	"bufio"
	"strconv"
	"strings"
)

type BevelData struct {
	Bevel           string
	BevelCVBA       string
	BevelCode       float64
	Manual          string
	BevelName       string
	BevelType       int
	BevelVariant    int
	PlateThickness  float64
	E               float64
	AngleTS         float64
	Angle2TS        float64
	AngleOS         float64
	Angle2OS        float64
	DepthTS         float64
	DepthOS         float64
	ChamferWidthTS  float64
	ChamferWidthOS  float64
	Angle2Wts       float64
	Angle2Wos       float64
	ChamferHeightTS float64
	ChamferHeightOS float64
	WeldTS          float64
	WeldOS          float64
}

func readBevelData(s *bufio.Scanner) *BevelData {
	bd := new(BevelData)

	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")

		if !ok {
			break
		}

		switch k {
		case "BEVEL":
			bd.Bevel = v

		case "BEVEL_CVBA":
			bd.BevelCVBA = v

		case "BEVEL_CODE":
			bd.BevelCode, _ = strconv.ParseFloat(v, 64)

		case "MANUAL":
			bd.Manual = v

		case "BEVEL_NAME":
			bd.BevelName = v

		case "BEVEL_TYPE":
			bd.BevelType, _ = strconv.Atoi(v)

		case "BEVEL_VARIANT":
			bd.BevelVariant, _ = strconv.Atoi(v)

		case "PLATE_THICKNESS":
			bd.PlateThickness, _ = strconv.ParseFloat(v, 64)

		case "E":
			bd.E, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_TS":
			bd.AngleTS, _ = strconv.ParseFloat(v, 64)

		case "ANGLE2_TS":
			bd.Angle2TS, _ = strconv.ParseFloat(v, 64)

		case "ANGLE_OS":
			bd.AngleOS, _ = strconv.ParseFloat(v, 64)

		case "ANGLE2_OS":
			bd.Angle2OS, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_TS":
			bd.DepthTS, _ = strconv.ParseFloat(v, 64)

		case "DEPTH_OS":
			bd.DepthOS, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_TS":
			bd.ChamferWidthTS, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_WIDTH_OS":
			bd.ChamferWidthOS, _ = strconv.ParseFloat(v, 64)

		case "ANGLE2_WTS":
			bd.Angle2Wts, _ = strconv.ParseFloat(v, 64)

		case "ANGLE2_WOS":
			bd.Angle2Wos, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_TS":
			bd.ChamferHeightTS, _ = strconv.ParseFloat(v, 64)

		case "CHAMFER_HEIGHT_OS":
			bd.ChamferHeightOS, _ = strconv.ParseFloat(v, 64)

		case "WELD_TS":
			bd.WeldTS, _ = strconv.ParseFloat(v, 64)

		case "WELD_OS":
			bd.WeldOS, _ = strconv.ParseFloat(v, 64)
		}
	}

	return bd
}
