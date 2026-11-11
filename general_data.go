package gen

import (
	"bufio"
	"strconv"
	"strings"
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
		key, value, found := strings.Cut(s.Text(), "=")

		if key == "GENERAL_DATA" {
			generalDataFound = true
			continue
		}

		if !found {
			break
		}

		switch key {
		case "RAW_THICKNESS":
			thickness, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}
			g.Thickness = thickness

		case "QUALITY":
			g.Quality = value

		case "NO_OF_PARTS":
			noOfParts, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			g.NoOfParts = noOfParts

		case "TYPE_OF_GENERIC_FILE":
			g.TypeOfGenericFile = value

		case "TYPE_OF_MANUFACT":
			g.TypeOfManufact = value

		case "BURNER_ID":
			g.BurnerID = value

		case "NEST_NAME":
			g.NestName = value

		case "RAW_NAME":
			g.RawName = value

		case "RAW_LENGTH":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.RawLength = v

		case "RAW_WIDTH":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.RawWidth = v

		case "RAW_AREA":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.RawArea = v

		case "DENSITY":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.Density = v

		case "SURF_TREAT_TS":
			g.SurfTreatTS = value

		case "SURF_TREAT_OS":
			g.SurfTreatOS = value

		case "REST_LENGTH":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.RestLength = v

		case "SPLIT_DATE":
			g.SplitDate = value

		case "PRODUCTION_DATE":
			g.ProductionDate = value

		case "TOTAL_IDLE":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.TotalIdle = v

		case "TOTAL_MARKING":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.TotalMarking = v

		case "TOTAL_BURNING":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.TotalBurning = v

		case "TOTAL_BLASTING":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.TotalBlasting = v

		case "TOTAL_MARK_BLAST":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.TotalMarkBlast = v

		case "TOTAL_NO_OF_STARTS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.TotalNoOfStarts = v

		case "NO_OF_BURNING_STARTS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.NoOfBurningStarts = v

		case "NO_OF_MARKING_STARTS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.NoOfMarkingStarts = v

		case "NO_OF_BLASTING_STARTS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.NoOfBlastingStarts = v

		case "NO_OF_MARK_BLAST_STARTS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.NoOfMarkBlastStarts = v

		case "STARTING_POS_X":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.StartingPosX = v

		case "STARTING_POS_Y":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.StartingPosY = v

		case "QUANTITY_NORMAL":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.QuantityNormal = v

		case "QUANTITY_MIRRORED":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.QuantityMirrored = v

		case "TOOL_ID":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.ToolID = v

		case "PROCESS_ID":
			g.ProcessID = value

		case "SEQUENCE":
			g.Sequence = value

		case "DISTANCE_Y1_Y2":
			v, err := parseOptFloat(value)
			if err != nil {
				return nil, err
			}
			g.DistanceY1Y2 = v

		case "CUTTING_GAS":
			v, err := parseOptInt(value)
			if err != nil {
				return nil, err
			}
			g.CuttingGas = v

		case "MATERIAL_CODE":
			g.MaterialCode = value

		case "PURCHASE_INFO":
			g.PurchaseInfo = value

		case "PLATE_ALIGNMENT":
			g.PlateAlignment = value

		case "VERIFICATION_DATE":
			g.VerificationDate = value
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
