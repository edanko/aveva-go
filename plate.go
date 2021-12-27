package gen

import (
	"bufio"
	"io"
	"strings"
)

func ReadPlate(r io.Reader) map[string]*PartData {
	s := bufio.NewScanner(r)

	var generalData *GeneralData
	s.Scan()
	if s.Text() == "GENERAL_DATA" {
		generalData = readGeneralData(s)
	} else {
		return nil
	}

	parts := make(map[string]*PartData)
	var lastPartName string

	for s.Scan() {
		switch s.Text() {
		case "PART_DATA":
			pd := readPartData(s)
			pd.Thickness = generalData.Thickness
			pd.Quality = generalData.Quality

			if p, ok := parts[pd.Name]; ok {
				p.Quantity++
			} else {
				parts[pd.Name] = pd
			}

			if generalData.NoOfParts == 0 {
				lastPartName = pd.Name
			}

		case "PART_INFORMATION":
			for s.Scan() {
				k, v, ok := strings.Cut(s.Text(), "=")

				if !ok {
					break
				}

				switch k {
				case "PART_NAME":
					lastPartName = v
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
			parts[lastPartName].MarkingData = append(parts[lastPartName].MarkingData, md)

		case "GEOMETRY_DATA":
			gd := readGeometryData(s)
			parts[lastPartName].GeometryData = append(parts[lastPartName].GeometryData, gd)

		case "STRING_DATA":
			sd := readStringData(s)
			parts[lastPartName].StringData = append(parts[lastPartName].StringData, sd)

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
			parts[lastPartName].BurningData = append(parts[lastPartName].BurningData, b)

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
