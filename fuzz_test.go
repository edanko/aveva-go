package gen

import (
	"bytes"
	"testing"
)

// The readers must never panic on arbitrary input, and anything they accept must
// be writable and re-readable. Since the format is line-oriented and unescaped,
// a malformed line is normal input rather than an impossible state, so this
// invariant is the real robustness contract.

func FuzzReadDocument(f *testing.F) {
	f.Add([]byte("GENERAL_DATA\r\nRAW_THICKNESS=1.00\r\nNO_OF_PARTS=0\r\nEND_OF_GENERAL_DATA\r\n"))
	f.Add([]byte(""))
	f.Add([]byte("PART_DATA\r\nNAME=X\r\nEND_OF_PART_DATA\r\n"))
	f.Add([]byte("GENERAL_DATA\r\nEND_OF_GENERAL_DATA\r\nPART_INFORMATION\r\nPART_NAME=P1\r\nEND_OF_PART_INFORMATION\r\n"))
	f.Add([]byte("GENERAL_DATA\nRAW_THICKNESS=1.00\nEND_OF_GENERAL_DATA\n"))
	f.Add([]byte("GENERAL_DATA\r\nRAW_THICKNESS=-\r\nEND_OF_GENERAL_DATA\r\n"))
	f.Add(bytes.Repeat([]byte("A=B\r\n"), 50))
	f.Add([]byte("GENERAL_DATA\r\nSTART_OF_CONTOUR\r\nEND_OF_CONTOUR\r\nEND_OF_GENERAL_DATA\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := ReadDocument(bytes.NewReader(data))
		if err != nil {
			return
		}

		var buf bytes.Buffer
		if err := WriteDocument(&buf, doc); err != nil {
			t.Fatalf("write failed for a document that parsed: %v", err)
		}

		if _, err := ReadDocument(bytes.NewReader(buf.Bytes())); err != nil {
			t.Fatalf("output of WriteDocument is not readable by this package: %v\ninput: %q", err, data)
		}
	})
}

func FuzzReadProfileDocument(f *testing.F) {
	const header = "TYPE_OF_GENERIC_FILE=LISTED_PROFILE\r\nVERSION=1.0\r\nUSAGE=LIST\r\n"

	f.Add([]byte(header + "COMMON_DATA\r\nNO_OF_PROFS=0\r\nEND_OF_COMMON_DATA\r\n"))
	f.Add([]byte(header))
	f.Add([]byte(header + "COMMON_DATA\r\nEND_OF_COMMON_DATA\r\nPROFILE_DATA\r\nNAME=P\r\nEND_OF_PROFILE_DATA\r\n"))
	f.Add([]byte(header + "COMMON_DATA\r\nSTART_OF_CONTOUR\r\nEND_OF_CONTOUR\r\nEND_OF_COMMON_DATA\r\n"))
	f.Add([]byte(header + "COMMON_DATA\r\nEND_OF_COMMON_DATA\r\nPROFILE_DATA\r\nEND_OF_PROFILE_DATA\r\nLEFT_END\r\nEND_OF_LEFT_END\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := ReadProfileDocument(bytes.NewReader(data))
		if err != nil {
			return
		}

		var buf bytes.Buffer
		if err := WriteProfiles(&buf, doc); err != nil {
			t.Fatalf("write failed for a document that parsed: %v", err)
		}

		if _, err := ReadProfileDocument(bytes.NewReader(buf.Bytes())); err != nil {
			t.Fatalf("output of WriteProfiles is not readable: %v\ninput: %q", err, data)
		}
	})
}
