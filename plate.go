package gen

import (
	"bufio"
	"io"
	"strings"
)

// placementsOf records the PART_DATA positions of one part name and how many of
// them have been bound to a PART_INFORMATION scope.
type placementsOf struct {
	indices []int
	used    int
}

// ReadParts parses a plate-part generic file and returns its parts.
//
// Blocks that carry no owning part (see Document) are not included; use
// ReadDocument when they matter.
func ReadParts(r io.Reader) ([]*Part, error) {
	doc, err := ReadDocument(r)
	if err != nil {
		return nil, err
	}

	return doc.Parts, nil
}

// ReadDocument parses a plate-part generic file, retaining both the parts and the
// document-level blocks that are not tied to one part.
func ReadDocument(r io.Reader) (*Document, error) {
	s := bufio.NewScanner(r)

	generalData, err := readGeneralData(s)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		Type:     generalData.TypeOfGenericFile,
		Manufact: generalData.TypeOfManufact,
		General:  generalData,
	}

	parts := make([]*Part, 0, generalData.NoOfParts)

	// A nest may place the same part more than once, and a part may be opened by
	// more than one PART_INFORMATION scope, so the relation is neither 1:1 nor a
	// function of the name alone. Track every placement and bind scopes to them
	// in order; once a name's placements are used up, further scopes accumulate
	// onto its last placement.
	placements := make(map[string]*placementsOf, generalData.NoOfParts)

	var lastPartIdx int

	for s.Scan() {
		switch s.Text() {
		case "PART_DATA":
			part, err := readPart(s)
			if err != nil {
				return nil, err
			}
			part.Thickness = generalData.Thickness
			part.Quality = generalData.Quality

			parts = append(parts, part)
			pl := placements[part.Name]
			if pl == nil {
				pl = &placementsOf{}
				placements[part.Name] = pl
			}
			pl.indices = append(pl.indices, len(parts)-1)

		case "PART_INFORMATION":
			info := PartInformation{}

			for s.Scan() {
				if s.Text() == "END_OF_PART_INFORMATION" {
					break
				}

				key, value, found := strings.Cut(s.Text(), "=")
				if !found {
					continue
				}

				if key == "TYPE_OF_WORK" {
					info.TypeOfWork = cleanStr(value)
					continue
				}

				// The format spells the owning part either way; the reference
				// reader maps both to the same key id.
				switch key {
				case "PART_NAME", "NAME":
					if value == "" {
						// A scope with no owning part, e.g. a TYPE_OF_WORK=LABELLING
						// run. Its payload is not part data; drop the association.
						lastPartIdx = -1
						continue
					}

					pl := placements[value]
					switch {
					case pl != nil && pl.used < len(pl.indices):
						lastPartIdx = pl.indices[pl.used]
						pl.used++
					case pl != nil:
						// More scopes than placements: accumulate onto the last
						// placement of this name.
						lastPartIdx = pl.indices[len(pl.indices)-1]
					default:
						// A part referenced with no PART_DATA. Single-part
						// documents carry the part attribute in GENERAL_DATA and
						// declare it only here, so synthesize it rather than
						// failing or dropping the payload.
						parts = append(parts, &Part{
							Name:      value,
							Thickness: generalData.Thickness,
							Quality:   generalData.Quality,
							PartCog:   &Point{},
							Quantity:  1,
						})
						lastPartIdx = len(parts) - 1
						placements[value] = &placementsOf{
							indices: []int{lastPartIdx},
							used:    1,
						}
					}
				}
			}

			if lastPartIdx >= 0 {
				info.PartName = parts[lastPartIdx].Name
				parts[lastPartIdx].Information = append(parts[lastPartIdx].Information, info)
			}

		case "MARKING_DATA":
			md, err := readMarking(s)
			if err != nil {
				return nil, err
			}

			if lastPartIdx < 0 {
				continue
			}
			parts[lastPartIdx].MarkingData = append(parts[lastPartIdx].MarkingData, md)

		case "GEOMETRY_DATA":
			gd, err := readGeometry(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx < 0 {
				continue
			}
			parts[lastPartIdx].GeometryData = append(parts[lastPartIdx].GeometryData, gd)

		case "STRING_DATA":
			sd, err := readStringData(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx < 0 {
				continue
			}
			parts[lastPartIdx].StringData = append(parts[lastPartIdx].StringData, sd)

		case "BURNING_DATA":
			b, err := readBurningData(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx < 0 {
				continue
			}
			parts[lastPartIdx].BurningData = append(parts[lastPartIdx].BurningData, b)

		case "LABELTEXT_DATA":
			lt, err := readLabelText(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].LabelTexts = append(parts[lastPartIdx].LabelTexts, lt)
			} else {
				doc.LabelTexts = append(doc.LabelTexts, lt)
			}

		case "LABELSYMBOL_DATA":
			ls, err := readLabelSymbol(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].LabelSymbols = append(parts[lastPartIdx].LabelSymbols, ls)
			} else {
				doc.LabelSymbols = append(doc.LabelSymbols, ls)
			}

		case "EDGE_DATA":
			e, err := readEdge(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].Edges = append(parts[lastPartIdx].Edges, e)
			}

		case "BLASTING_DATA", "GSD_DATA", "MARKING_SHAPE_DATA", "BLASTING_MARKING_DATA":
			// Documented in the shipped reader but absent from every available
			// sample; capture verbatim instead of dropping.
			name := s.Text()
			rs, err := readRawSection(s, name)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].RawSections = append(parts[lastPartIdx].RawSections, rs)
			}

		case "AUXILIARY_DATA":
			a, err := readAuxiliary(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].Auxiliaries = append(parts[lastPartIdx].Auxiliaries, a)
			}

		case "IDLE_DATA":
			idle, err := readIdleData(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].IdleData = append(parts[lastPartIdx].IdleData, idle)
			}

		case "BUMP_DATA":
			bump, err := readBumpData(s)
			if err != nil {
				return nil, err
			}
			if lastPartIdx >= 0 {
				parts[lastPartIdx].BumpData = append(parts[lastPartIdx].BumpData, bump)
			} else if len(parts) > 0 {
				// BUMP_DATA describes the plate edge and can appear after the
				// last scope has closed; attach it to the last part.
				parts[len(parts)-1].BumpData = append(parts[len(parts)-1].BumpData, bump)
			}
		default:
			// Blank lines carry nothing and are ignorable.
			text := strings.TrimSpace(s.Text())
			if text == "" {
				continue
			}

			// A bare marker that is neither known here nor a KEY=VALUE pair is
			// the end of the block we are inside. Do not scan on: doing so
			// discards every later part in the file.
			if !strings.Contains(text, "=") {
				doc.Parts = parts
				return doc, nil
			}
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	doc.Parts = parts

	return doc, nil
}
