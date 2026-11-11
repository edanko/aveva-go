package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOptFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    float64
		wantErr bool
	}{
		{name: "plain", in: "1.5", want: 1.5},
		{name: "negative", in: "-2.25", want: -2.25},
		{name: "integer form", in: "42", want: 42},
		{name: "scientific", in: "7.85e-06", want: 7.85e-06},
		{name: "expanded small", in: "0.00000785", want: 0.00000785},
		{name: "leading and trailing spaces", in: "  3.5  ", want: 3.5},
		{name: "dash sentinel", in: "-", want: 0},
		{name: "empty", in: "", want: 0},
		{name: "only spaces", in: "   ", want: 0},
		{name: "zero", in: "0", want: 0},
		{name: "negative zero", in: "-0.00", want: 0},
		{name: "garbage", in: "abc", wantErr: true},
		{name: "partial", in: "1.2.3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOptFloat(tt.in)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.InDelta(t, tt.want, got, 1e-12)
		})
	}
}

func TestParseOptInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{name: "plain", in: "42", want: 42},
		{name: "negative", in: "-1", want: -1},
		{name: "zero", in: "0", want: 0},
		{name: "dash sentinel", in: "-", want: 0},
		{name: "empty", in: "", want: 0},
		{name: "spaces", in: " 7 ", want: 7},
		{name: "float is not an int", in: "1.5", wantErr: true},
		{name: "garbage", in: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOptInt(tt.in)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCleanStr(t *testing.T) {
	t.Parallel()

	// AVEVA right-pads fixed-width fields such as TEXT_TYPE to 26 characters.
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "space padded", in: "PANEL_NAME                ", want: "PANEL_NAME"},
		{name: "already clean", in: "POSNO", want: "POSNO"},
		{name: "empty", in: "", want: ""},
		{name: "all spaces", in: "   ", want: ""},
		{name: "leading spaces", in: "  X", want: "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, cleanStr(tt.in))
		})
	}
}

func TestFormatFloatWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  float64
		format string
		want   string
	}{
		{name: "coord truncates to five decimals", value: 1.23456789, format: fmtCoord, want: "1.23457"},
		{name: "total truncates to two decimals", value: 1.239, format: fmtTotal, want: "1.24"},
		// The whole point of fmtExact: no precision loss for LISTED_PROFILE.
		{name: "exact keeps small value", value: 0.00000785, format: fmtExact, want: "0.00000785"},
		{name: "exact keeps long value", value: 2578.6708327727, format: fmtExact, want: "2578.6708327727"},
		{name: "exact drops trailing zeros", value: 17.282, format: fmtExact, want: "17.282"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, formatFloat(tt.value, tt.format))
		})
	}
}
