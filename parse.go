package gen

import (
	"strconv"
	"strings"
)

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

// cleanStr trims the space padding AVEVA uses for fixed-width string fields,
// e.g. TEXT_TYPE and SYMBOL_TYPE are padded to 26 characters.
func cleanStr(v string) string {
	return strings.TrimSpace(v)
}
