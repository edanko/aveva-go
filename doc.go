// Package gen reads and writes AVEVA Marine generic files (".gen").
//
// A generic file is line-oriented KEY=VALUE text, terminated by bare section
// markers such as PART_DATA … END_OF_PART_DATA. Two document families are in
// use and both are supported here:
//
//   - PLATE_PART, a nesting. Read it with [ReadParts] or [ReadDocument] and
//     write it with [WriteParts] or [WriteDocument].
//   - LISTED_PROFILE, a profile list. Read it with [ReadProfiles] or
//     [ReadProfileDocument] and write it with [WriteProfiles].
//
// # Reading production data
//
// The readers accept what production files actually contain rather than an
// idealised grammar. Documents may omit PART_DATA entirely, a labelling scope may
// carry an empty PART_NAME, and a document may contain more PART_INFORMATION
// scopes than part placements — a part can be placed several times and opened by
// a different number of scopes, so payloads bind positionally and surplus scopes
// accumulate on the last placement. A "-" value means the parameter does not
// apply, which is not the same as zero.
//
// A document of the other family returns [ErrNotPlatePart] or
// [ErrNotListedProfile] instead of an empty result, so a caller can distinguish a
// wrong document type from a document that genuinely holds no parts. Real parse
// failures return their own error and are never reported as a family mismatch.
//
// # Writing
//
// Writing is structural, not byte-for-byte. The corpus contains several writer
// dialects that disagree about field presence and about empty-versus-zero, so no
// single writer can reproduce all of them. A parse → write → parse cycle
// reproduces the model exactly; the original bytes are not guaranteed.
//
// # Incomplete schemas
//
// Some sections have no available sample. They are captured verbatim in
// [Part.RawSections] rather than dropped or modelled on guesswork, so no data is
// lost while their schema remains unvalidated.
//
// # Example
//
//	f, err := os.Open("nest.gen")
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//
//	parts, err := ReadParts(f)
//	if errors.Is(err, ErrNotPlatePart) {
//		return nil // a profile file or another family
//	}
//	if err != nil {
//		return err
//	}
//
//	for _, p := range parts {
//		p.Mirror()
//	}
//
//	return WriteParts(w, parts)
package gen
