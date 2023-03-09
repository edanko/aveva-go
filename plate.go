package gen

import (
	"bufio"
	"io"
	"strings"
)

func ReadPlate(r io.Reader) []*PartData {
	s := bufio.NewScanner(r)

	var generalData *GeneralData
	s.Scan()
	if s.Text() == "GENERAL_DATA" {
		generalData = readGeneralData(s)
	} else {
		return nil
	}

	space := 1
	if generalData.NoOfParts > 0 {
		space = generalData.NoOfParts
	}

	parts := make([]*PartData, 0, space)
	nameIndex := make(map[string]int, space)

	var lastPartIdx int

	for s.Scan() {
		switch s.Text() {
		case "PART_DATA":
			pd := readPartData(s)
			pd.Thickness = generalData.Thickness
			pd.Quality = generalData.Quality

			parts = append(parts, pd)
			nameIndex[pd.Name] = len(parts) - 1

		case "PART_INFORMATION":
			for s.Scan() {
				k, v, ok := strings.Cut(s.Text(), "=")

				if !ok {
					break
				}

				switch k {
				case "PART_NAME":
					lastPartIdx = nameIndex[v]
				}
			}

		case "IDLE_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_IDLE_DATA":
					break
				}
			}

		case "MARKING_DATA":
			md := readMarkingData(s)
			parts[lastPartIdx].MarkingData = append(parts[lastPartIdx].MarkingData, md)

		case "GEOMETRY_DATA":
			gd := readGeometryData(s)
			parts[lastPartIdx].GeometryData = append(parts[lastPartIdx].GeometryData, gd)

		case "STRING_DATA":
			sd := readStringData(s)
			parts[lastPartIdx].StringData = append(parts[lastPartIdx].StringData, sd)

		case "BUMP_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_BUMP_DATA":
					break
				}
			}

		case "LABELTEXT_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_LABELTEXT_DATA":
					break
				}
			}

		case "LABELSYMBOL_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_LABELSYMBOL_DATA":
					break
				}
			}

		case "BURNING_DATA":
			b := readBurningData(s)
			parts[lastPartIdx].BurningData = append(parts[lastPartIdx].BurningData, b)

		case "EDGE_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_EDGE_DATA":
					break
				}
			}
		}
	}

	return parts
}
