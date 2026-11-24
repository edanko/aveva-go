package gen

import (
	"bufio"
)

type Marking struct {
	MarkingSide             string
	MarkingType             string
	MarkingName             string
	MarkingUserType         int
	MarkingAttri            string
	MarkingAttriSb          string
	MarkingShipside         string
	MarkingAssemblyLow      string
	MarkingProfileThickness float64
	MarkingWeld             float64
	MarkingDirection        int
	NumberOfHeads           int
	GeometryValidFor        int
	DistanceY1Y2            float64
	MtrlSideU               float64
	MtrlSideV               float64
	InclinationAngle        float64
	Contour                 *Contour
	Comment                 string
}

func readMarking(s *bufio.Scanner) (Marking, error) {
	marking := Marking{}

	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "START_OF_CONTOUR" {
			contour, err := readContour(s)
			if err != nil {
				return Marking{}, err
			}
			marking.Contour = contour
			continue
		}

		if !found {
			break
		}

		switch key {
		case "MARKING_SIDE":
			marking.MarkingSide = string(value)

		case "MARKING_TYPE":
			marking.MarkingType = string(value)

		case "MARKING_NAME":
			marking.MarkingName = string(value)

		case "MARKING_USER_TYPE":
			markingType, err := atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingUserType = markingType

		case "MARKING_ATTRI":
			marking.MarkingAttri = string(value)

		case "MARKING_ATTRI_SB":
			marking.MarkingAttriSb = string(value)

		case "MARKING_SHIPSIDE":
			marking.MarkingShipside = string(value)

		case "MARKING_ASSEMBLY_LOW":
			marking.MarkingAssemblyLow = string(value)

		case "MARKING_PROFILE_THICKNESS":
			markingProfileThickness, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingProfileThickness = markingProfileThickness

		case "MARKING_WELD":
			markingWeld, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingWeld = markingWeld

		case "MARKING_DIRECTION":
			markingDirection, err := atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MarkingDirection = markingDirection

		case "NUMBER_OF_HEADS":
			numberOfHeads, err := atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.NumberOfHeads = numberOfHeads

		case "GEOMETRY_VALID_FOR":
			geometryValidFor, err := atoi(value)
			if err != nil {
				return Marking{}, err
			}
			marking.GeometryValidFor = geometryValidFor

		case "DISTANCE_Y1_Y2":
			distanceY1Y2, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.DistanceY1Y2 = distanceY1Y2

		case "MTRL_SIDE_U":
			mtrlSideU, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MtrlSideU = mtrlSideU

		case "MTRL_SIDE_V":
			mtrlSideV, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.MtrlSideV = mtrlSideV

		case "INCLINATION_ANGLE":
			inclinationAngle, err := atof(value)
			if err != nil {
				return Marking{}, err
			}
			marking.InclinationAngle = inclinationAngle

		case "COMMENT":
			marking.Comment = string(value)
		}
	}
	return marking, nil
}
