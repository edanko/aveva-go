package gen

import (
	"bufio"
	"io"
)

func ReadListedProfile(r io.Reader) map[string]ProfileData {
	s := bufio.NewScanner(r)

	// type
	s.Scan()
	if s.Text() != "TYPE_OF_GENERIC_FILE=LISTED_PROFILE" {
		return nil
	}
	// version
	s.Scan()
	// usage
	s.Scan()

	var commonData CommonData
	profs := make(map[string]ProfileData)

	for s.Scan() {
		t := s.Text()
		switch t {
		case "COMMON_DATA":
			commonData = readCommonData(s)

		case "PROFILE_DATA":
			pd := readProfileData(s)
			pd.Shape = commonData.Shape
			pd.Dimension = commonData.Dimension
			pd.Quality = commonData.Quality

			// TODO: think about map key, may be better use pd.BlockNo+pd.PosNo
			if p, ok := profs[pd.Name]; ok {
				p.Quantity++
			} else {
				profs[pd.Name] = pd
			}

		case "REST_DATA":
			for s.Scan() {
				switch s.Text() {
				case "END_OF_REST_DATA":
					break
				}
			}
		}
	}

	return profs
}
