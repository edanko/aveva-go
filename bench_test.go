package gen

import (
	"bytes"
	"os"
	"testing"
)

func loadFixture(b *testing.B, path string) []byte {
	b.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("fixture %s: %v", path, err)
	}

	return data
}

func BenchmarkReadDocument(b *testing.B) {
	for _, path := range plateFixtures {
		b.Run(path, func(b *testing.B) {
			data := loadFixture(b, path)

			b.ReportAllocs()
			b.SetBytes(int64(len(data)))

			for b.Loop() {
				if _, err := ReadDocument(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkReadProfileDocument(b *testing.B) {
	for _, path := range profileFixtures {
		b.Run(path, func(b *testing.B) {
			data := loadFixture(b, path)

			b.ReportAllocs()
			b.SetBytes(int64(len(data)))

			for b.Loop() {
				if _, err := ReadProfileDocument(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWriteDocument(b *testing.B) {
	for _, path := range plateFixtures {
		b.Run(path, func(b *testing.B) {
			data := loadFixture(b, path)

			doc, err := ReadDocument(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				var buf bytes.Buffer
				if err := WriteDocument(&buf, doc); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWriteProfiles(b *testing.B) {
	for _, path := range profileFixtures {
		b.Run(path, func(b *testing.B) {
			data := loadFixture(b, path)

			doc, err := ReadProfileDocument(bytes.NewReader(data))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				var buf bytes.Buffer
				if err := WriteProfiles(&buf, doc); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Reading from a plain byte slice is the realistic path for a caller that has
// already loaded the file.
func BenchmarkReadDocumentParallel(b *testing.B) {
	data := loadFixture(b, "testdata/nest.gen")

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := ReadDocument(bytes.NewReader(data)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
