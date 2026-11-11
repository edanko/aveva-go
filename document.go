package gen

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// A .gen file holds one of several document families, and they are implemented
// by different components of the AVEVA installation. Readers therefore report a
// wrong document family explicitly instead of returning an empty result that is
// indistinguishable from a valid, payload-free document.
var (
	// ErrNotPlatePart is returned by ReadParts for a document that is not a
	// plate-part family file (LISTED_PROFILE, or not a generic file at all).
	ErrNotPlatePart = errors.New("gen: not a plate part document")

	// ErrNotListedProfile is returned by ReadProfiles for a document that is not
	// a LISTED_PROFILE file.
	ErrNotListedProfile = errors.New("gen: not a listed profile document")
)

// Document is a parsed plate-part generic file. Everything that belongs to a
// part hangs off Part; blocks emitted without an owning part are collected at
// document level instead of being dropped.
//
// A nest-wide labelling pass does exactly that: one PART_INFORMATION with an
// empty PART_NAME and TYPE_OF_WORK=LABELLING, followed by every label for the
// whole nest. Those labels carry the part identity inside their own TEXT value,
// so they are exposed here rather than attributed to whichever part happened to
// precede them.
type Document struct {
	Type     string
	Manufact string
	General  *GeneralData
	Parts    []*Part

	// LabelTexts / LabelSymbols hold blocks that were not preceded by a
	// PART_INFORMATION naming a part.
	LabelTexts   []LabelText
	LabelSymbols []LabelSymbol

	// WriteOptions, when set, is used by WriteDocument unless an explicit
	// option is passed to that call.
	WriteOptions WriteOptions
}

// DocumentType reports the value of TYPE_OF_GENERIC_FILE without parsing the
// document. Plate-part files carry it inside GENERAL_DATA, listed-profile files
// on the first line, so the first lines are scanned rather than one fixed line.
func DocumentType(r io.Reader) (string, error) {
	s := bufio.NewScanner(r)

	for i := 0; i < 16 && s.Scan(); i++ {
		if v, ok := strings.CutPrefix(s.Text(), "TYPE_OF_GENERIC_FILE="); ok {
			return strings.TrimSpace(v), nil
		}
	}

	return "", s.Err()
}
