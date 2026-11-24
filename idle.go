package gen

import (
	"bufio"
	"bytes"
)

// IdleData is an IDLE_DATA block: the rapid/traverse path the machine takes
// between cuts. It carries a contour and no cut or bevel data, so it is dropped
// by consumers that only draw geometry.
//
// Ordering below is taken from the corpus (93,571 blocks across 769 files), which
// is direct evidence of what shipped files contain. The decompiled writer was not
// reachable when this was written: REA had no usable provider (Ghidra 12.2
// installed, REA requires 12.1.x; Hopper/IDA/Kuna unconfigured). Layout:
//
//	IDLE_DATA
//	START_OF_CONTOUR
//	NO_OF_SEG / START_U / START_V
//	<segments>
//	END_OF_CONTOUR
//	END_OF_IDLE_DATA
type IdleData struct {
	Contour *Contour
}

func readIdleData(s *bufio.Scanner) (IdleData, error) {
	idle := IdleData{}

	for s.Scan() {
		// Decide everything about the line before readContour can refill the
		// scanner buffer that line aliases.
		line := s.Bytes()
		startOfContour := bytes.Equal(line, startOfContourMarker)
		endOfIdleData := bytes.Equal(line, endOfIdleDataMarker)
		hasField := bytes.Contains(line, eqSep)

		switch {
		case startOfContour:
			contour, err := readContour(s)
			if err != nil {
				return IdleData{}, err
			}
			idle.Contour = contour

		case endOfIdleData:
			return idle, nil
		}

		if !hasField && !startOfContour {
			return idle, s.Err()
		}
	}

	return idle, s.Err()
}

// BumpData is a BUMP_DATA block: the plate edge's bump/cam profile. One per
// document in all 2,974 corpus files that have it.
//
// BUMP_DATA appears in no artifact in this installation — the string is absent
// from TbGenPostProcessor.dll, mar.dll and mar_p.dll, so it is emitted by a
// component not present here and its schema is known only from samples:
//
//	BUMP_DATA
//	SHAPE=BUMP_ALL
//	START_OF_CONTOUR
//	NO_OF_SEG / START_U / START_V
//	<segments>
//	END_OF_CONTOUR
//	END_OF_BUMP_DATA
type BumpData struct {
	Shape   string // SHAPE, observed value BUMP_ALL
	Contour *Contour
}

func readBumpData(s *bufio.Scanner) (BumpData, error) {
	bump := BumpData{}

	for s.Scan() {
		line := s.Bytes()
		startOfContour := bytes.Equal(line, startOfContourMarker)
		endOfBumpData := bytes.Equal(line, endOfBumpDataMarker)
		key, value, found := bytes.Cut(line, eqSep)

		switch {
		case startOfContour:
			contour, err := readContour(s)
			if err != nil {
				return BumpData{}, err
			}
			bump.Contour = contour

		case endOfBumpData:
			return bump, nil
		}

		if !found {
			return bump, s.Err()
		}

		if bytes.Equal(key, shapeKey) {
			bump.Shape = string(value)
		}
	}

	return bump, s.Err()
}
