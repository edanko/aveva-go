package gen

import (
	"bufio"
	"io"
)

func ReadProfile(r io.Reader) []*ProfileData {
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

	// common data
	s.Scan()
	if s.Text() != "COMMON_DATA" {
		return nil
	}
	commonData := readCommonData(s)
	profs := make([]*ProfileData, 0, commonData.NoOfProfs)

	for s.Scan() {
		t := s.Text()
		switch t {
		case "PROFILE_DATA":
			pd := readProfileData(s)
			pd.Shape = commonData.Shape
			pd.Dimension = commonData.Dimension
			pd.Quality = commonData.Quality

			profs = append(profs, pd)

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
