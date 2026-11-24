package gen

import (
	"bufio"
	"strconv"
)

type CommonData struct {
	Shape           string
	Dimension       string
	Quality         string
	NoOfProfs       uint64
	NestName        string
	Contour         *Contour
	WebHeight       int
	WebThickness    int
	MaxWebHeight    int
	FlangeWidth     float64
	FlangeThickness float64
	RawLength       float64
	RestLength      float64
	UsedLength      float64
	LastScrapLength float64
	Density         float64
	Area            float64
	TotalWeight     float64
	RawBar          string
	BuyingMark      string
	CuttingStn      int
	TextHeight      int
	TextWidth       int
	TextPlacing     int
	TextDistTop     int
	TextMax1        int
	TextMax2        int
	TextHor1        int
	TextVer1        int
	TextHor2        int
	TextVer2        int
	TextHor3        int
	TextVer3        int
	Tscrap          int
	Rest            int
	GsdOneLim       int
	Gsd             float64
	GsdDist         float64
}

func readCommonData(s *bufio.Scanner) (*CommonData, error) {
	cd := &CommonData{}

	for s.Scan() {
		k, raw, ok := nextKeyValue(s)

		if k == "START_OF_CONTOUR" {
			contour, err := readContour(s)
			if err != nil {
				return nil, err
			}
			cd.Contour = contour

			continue
		}

		if !ok {
			return cd, nil
		}

		switch k {
		case "SHAPE":
			cd.Shape = string(raw)

		case "DIMENSION":
			cd.Dimension = string(raw)

		case "QUALITY":
			cd.Quality = string(raw)

		case "NO_OF_PROFS":
			noOfProfs, err := strconv.ParseUint(unsafeString(raw), 10, 64)
			if err != nil {
				return nil, err
			}
			cd.NoOfProfs = noOfProfs

		case "NEST_NAME":
			cd.NestName = string(raw)

		case "RAW_BAR":
			cd.RawBar = string(raw)

		case "BUYING_MARK":
			cd.BuyingMark = string(raw)

		case "WEB_HEIGHT":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.WebHeight = v

		case "WEB_THICKNESS":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.WebThickness = v

		case "MAX_WEB_HEIGHT":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.MaxWebHeight = v

		case "CUTTING_STN":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.CuttingStn = v

		case "TEXT_HEIGHT":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextHeight = v

		case "TEXT_WIDTH":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextWidth = v

		case "TEXT_PLACING":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextPlacing = v

		case "TEXT_DIST_TOP":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextDistTop = v

		case "TEXT_MAX1":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextMax1 = v

		case "TEXT_MAX2":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextMax2 = v

		case "TEXT_HOR1":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextHor1 = v

		case "TEXT_VER1":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextVer1 = v

		case "TEXT_HOR2":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextHor2 = v

		case "TEXT_VER2":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextVer2 = v

		case "TEXT_HOR3":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextHor3 = v

		case "TEXT_VER3":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TextVer3 = v

		case "TSCRAP":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.Tscrap = v

		case "REST":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.Rest = v

		case "GSD_ONE_LIM":
			v, err := parseOptIntBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.GsdOneLim = v

		case "FLANGE_WIDTH":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.FlangeWidth = v

		case "FLANGE_THICKNESS":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.FlangeThickness = v

		case "RAW_LENGTH":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.RawLength = v

		case "REST_LENGTH":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.RestLength = v

		case "USED_LENGTH":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.UsedLength = v

		case "LAST_SCRAP_LENGTH":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.LastScrapLength = v

		case "DENSITY":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.Density = v

		case "AREA":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.Area = v

		case "TOTAL_WEIGHT":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.TotalWeight = v

		case "GSD":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.Gsd = v

		case "GSD_DIST":
			v, err := parseOptBytes(raw)
			if err != nil {
				return nil, err
			}
			cd.GsdDist = v
		}
	}
	return cd, nil
}
