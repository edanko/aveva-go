package gen

import (
	"bufio"
)

type GeneralData struct {
	Thickness           float64
	Quality             string
	NoOfParts           int
	TypeOfGenericFile   string
	TypeOfManufact      string
	BurnerID            string
	NestName            string
	RawName             string
	RawLength           float64
	RawWidth            float64
	RawArea             float64
	Density             float64
	SurfTreatTS         string
	SurfTreatOS         string
	RestLength          float64
	SplitDate           string
	ProductionDate      string
	TotalIdle           float64
	TotalMarking        float64
	TotalBurning        float64
	TotalBlasting       float64
	TotalMarkBlast      float64
	TotalNoOfStarts     int
	NoOfBurningStarts   int
	NoOfMarkingStarts   int
	NoOfBlastingStarts  int
	NoOfMarkBlastStarts int
	StartingPosX        float64
	StartingPosY        float64
	QuantityNormal      int
	QuantityMirrored    int
	ToolID              int
	ProcessID           string
	Sequence            string
	DistanceY1Y2        float64
	CuttingGas          int
	MaterialCode        string
	PurchaseInfo        string
	PlateAlignment      string
	VerificationDate    string
}

func readGeneralData(s *bufio.Scanner) (*GeneralData, error) {
	g := &GeneralData{}
	generalDataFound := false
	for s.Scan() {
		key, value, found := nextKeyValue(s)

		if key == "GENERAL_DATA" {
			generalDataFound = true
			continue
		}

		if !found {
			break
		}

		switch key {
		case "RAW_THICKNESS":
			thickness, err := atof(value)
			if err != nil {
				return nil, err
			}
			g.Thickness = thickness

		case "QUALITY":
			g.Quality = string(value)

		case "NO_OF_PARTS":
			noOfParts, err := atoi(value)
			if err != nil {
				return nil, err
			}
			g.NoOfParts = noOfParts

		case "TYPE_OF_GENERIC_FILE":
			g.TypeOfGenericFile = string(value)

		case "TYPE_OF_MANUFACT":
			g.TypeOfManufact = string(value)

		case "BURNER_ID":
			g.BurnerID = string(value)

		case "NEST_NAME":
			g.NestName = string(value)

		case "RAW_NAME":
			g.RawName = string(value)

		case "RAW_LENGTH":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.RawLength = v

		case "RAW_WIDTH":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.RawWidth = v

		case "RAW_AREA":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.RawArea = v

		case "DENSITY":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.Density = v

		case "SURF_TREAT_TS":
			g.SurfTreatTS = string(value)

		case "SURF_TREAT_OS":
			g.SurfTreatOS = string(value)

		case "REST_LENGTH":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.RestLength = v

		case "SPLIT_DATE":
			g.SplitDate = string(value)

		case "PRODUCTION_DATE":
			g.ProductionDate = string(value)

		case "TOTAL_IDLE":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalIdle = v

		case "TOTAL_MARKING":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalMarking = v

		case "TOTAL_BURNING":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalBurning = v

		case "TOTAL_BLASTING":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalBlasting = v

		case "TOTAL_MARK_BLAST":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalMarkBlast = v

		case "TOTAL_NO_OF_STARTS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.TotalNoOfStarts = v

		case "NO_OF_BURNING_STARTS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.NoOfBurningStarts = v

		case "NO_OF_MARKING_STARTS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.NoOfMarkingStarts = v

		case "NO_OF_BLASTING_STARTS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.NoOfBlastingStarts = v

		case "NO_OF_MARK_BLAST_STARTS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.NoOfMarkBlastStarts = v

		case "STARTING_POS_X":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.StartingPosX = v

		case "STARTING_POS_Y":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.StartingPosY = v

		case "QUANTITY_NORMAL":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.QuantityNormal = v

		case "QUANTITY_MIRRORED":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.QuantityMirrored = v

		case "TOOL_ID":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.ToolID = v

		case "PROCESS_ID":
			g.ProcessID = string(value)

		case "SEQUENCE":
			g.Sequence = string(value)

		case "DISTANCE_Y1_Y2":
			v, err := parseOptBytes(value)
			if err != nil {
				return nil, err
			}
			g.DistanceY1Y2 = v

		case "CUTTING_GAS":
			v, err := parseOptIntBytes(value)
			if err != nil {
				return nil, err
			}
			g.CuttingGas = v

		case "MATERIAL_CODE":
			g.MaterialCode = string(value)

		case "PURCHASE_INFO":
			g.PurchaseInfo = string(value)

		case "PLATE_ALIGNMENT":
			g.PlateAlignment = string(value)

		case "VERIFICATION_DATE":
			g.VerificationDate = string(value)
		}
	}

	if !generalDataFound {
		// The family marker is absent: this is a different kind of document.
		// A parse failure inside GENERAL_DATA, by contrast, is a malformed
		// plate document and must surface as itself so callers do not skip it
		// as if it were another family.
		return nil, ErrNotPlatePart
	}
	return g, nil
}
