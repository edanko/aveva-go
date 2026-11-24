package gen

import (
	"bufio"
)

type Bevel struct {
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

func readBevel(s *bufio.Scanner) (*Bevel, error) {
	bevel := &Bevel{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)
		if !found {
			break
		}

		switch key {
		case "BEVEL":
			bevel.Bevel = string(value)

		case "BEVEL_CVBA":
			bevel.BevelCVBA = string(value)

		case "BEVEL_CODE":
			bevelCode, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.BevelCode = bevelCode

		case "MANUAL":
			bevel.Manual = string(value)

		case "BEVEL_NAME":
			bevel.BevelName = string(value)

		case "BEVEL_TYPE":
			bevelType, err := atoi(value)
			if err != nil {
				return nil, err
			}
			bevel.BevelType = bevelType

		case "BEVEL_VARIANT":
			bevelVariant, err := atoi(value)
			if err != nil {
				return nil, err
			}
			bevel.BevelVariant = bevelVariant

		case "PLATE_THICKNESS":
			plateThickness, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.PlateThickness = plateThickness

		case "E":
			e, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.E = e

		case "ANGLE_TS":
			angleTS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.AngleTS = angleTS

		case "ANGLE2_TS":
			angle2TS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.Angle2TS = angle2TS

		case "ANGLE_OS":
			angleOS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.AngleOS = angleOS

		case "ANGLE2_OS":
			angle2OS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.Angle2OS = angle2OS

		case "DEPTH_TS":
			depthTS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.DepthTS = depthTS

		case "DEPTH_OS":
			depthOS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.DepthOS = depthOS

		case "CHAMFER_WIDTH_TS":
			chamferWidthTS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.ChamferWidthTS = chamferWidthTS

		case "CHAMFER_WIDTH_OS":
			chamferWidthOS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.ChamferWidthOS = chamferWidthOS

		case "ANGLE2_WTS":
			angle2Wts, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.Angle2Wts = angle2Wts

		case "ANGLE2_WOS":
			angle2Wos, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.Angle2Wos = angle2Wos

		case "CHAMFER_HEIGHT_TS":
			chamferHeightTS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.ChamferHeightTS = chamferHeightTS

		case "CHAMFER_HEIGHT_OS":
			chamferHeightOS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.ChamferHeightOS = chamferHeightOS

		case "WELD_TS":
			weldTS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.WeldTS = weldTS

		case "WELD_OS":
			weldOS, err := atof(value)
			if err != nil {
				return nil, err
			}
			bevel.WeldOS = weldOS
		}
	}

	return bevel, nil
}
