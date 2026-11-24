package gen

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"unsafe"
)

// The readers walk a bufio.Scanner line by line and split each line at the first
// '='. Scanner.Text() copies every line into a fresh string, and the readers
// then keep only the string-valued fields: profiling the readers showed those
// copies were their largest single source of garbage, because the numeric
// majority is parsed and thrown away. The helpers below work on Scanner.Bytes()
// so only the values a caller actually retains are ever copied.

// eqSep is the separator the generic format uses between key and value.
var eqSep = []byte("=")

// Markers are compared as bytes so a scan loop needs no string copy for them.
var (
	endOfTransformationData = []byte("END_OF_TRANSFORMATION_DATA")
	endOfPartInformation    = []byte("END_OF_PART_INFORMATION")
	endOfIdleDataMarker     = []byte("END_OF_IDLE_DATA")
	endOfBumpDataMarker     = []byte("END_OF_BUMP_DATA")
	startOfContourMarker    = []byte("START_OF_CONTOUR")
	shapeKey                = []byte("SHAPE")
)

// nextKeyValue scans one line and splits it at the first '='.
//
// Both results alias the scanner's buffer and are invalidated by the next Scan.
// key is safe to compare and value is safe to parse, but anything that outlives
// the current iteration must be copied, with string(value) or cleanBytes(value).
func nextKeyValue(s *bufio.Scanner) (key string, value []byte, found bool) {
	k, v, ok := bytes.Cut(s.Bytes(), eqSep)

	return unsafeString(k), v, ok
}

// unsafeString views b as a string without copying it. The caller must not let
// the result outlive b; every use in this package is a comparison or a parse,
// never a store.
func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	// #nosec G103 -- the zero-copy view is the whole point, and the result never
	// outlives b in this package.
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// cleanBytes is cleanStr for a field the scanner still owns, returning an
// independent string trimmed to its content.
func cleanBytes(v []byte) string {
	return string(bytes.TrimSpace(v))
}

// atof is strconv.ParseFloat for a field the scanner still owns.
func atof(v []byte) (float64, error) {
	return strconv.ParseFloat(unsafeString(v), 64)
}

// atob is strconv.ParseBool for a field the scanner still owns.
func atob(v []byte) (bool, error) {
	return strconv.ParseBool(unsafeString(v))
}

// atoi is strconv.Atoi for a field the scanner still owns.
func atoi(v []byte) (int, error) {
	return strconv.Atoi(unsafeString(v))
}

// The generic file stores "not defined" in three different ways depending on the
// writer: a lone "-" (LISTED_PROFILE end cuts), a completely empty value, and
// space padding. Numeric parsing therefore has to be tolerant everywhere; a bare
// strconv call fails on real files.

// parseOptFloat parses a numeric field, mapping the "-" sentinel and an empty
// value to 0. End-cut parameters use "-" for fields the given ENDCUT_TYPE does
// not define, so ENDCUT_TYPE remains the authoritative way to tell an absent
// radius from a zero one.
func parseOptFloat(v string) (float64, error) {
	v = strings.TrimSpace(v)
	if v == "" || v == "-" {
		return 0, nil
	}

	return strconv.ParseFloat(v, 64)
}

// parseOptInt is parseOptFloat for integer fields.
func parseOptInt(v string) (int, error) {
	v = strings.TrimSpace(v)
	if v == "" || v == "-" {
		return 0, nil
	}

	return strconv.Atoi(v)
}

// parseOptBytes is parseOptFloat for a field the scanner still owns: it trims,
// maps the "-" sentinel and an empty value to 0, and otherwise parses in place.
func parseOptBytes(v []byte) (float64, error) {
	v = bytes.TrimSpace(v)
	if len(v) == 0 || (len(v) == 1 && v[0] == '-') {
		return 0, nil
	}

	return atof(v)
}

// cleanStr trims the space padding AVEVA uses for fixed-width string fields,
// e.g. TEXT_TYPE and SYMBOL_TYPE are padded to 26 characters.
func cleanStr(v string) string {
	return strings.TrimSpace(v)
}

// parseOptIntBytes is parseOptInt for a field the scanner still owns.
func parseOptIntBytes(v []byte) (int, error) {
	v = bytes.TrimSpace(v)
	if len(v) == 0 || (len(v) == 1 && v[0] == '-') {
		return 0, nil
	}

	return atoi(v)
}
