package stringutil

import (
	"fmt"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{1}, "1"},
		{"10", args{10}, "10"},
		{"100", args{100}, "100"},
		{"1000", args{1000}, "1,000"},
		{"10000", args{10000}, "10,000"},
		{"100000", args{100000}, "100,000"},
		{"1000000", args{1000000}, "1,000,000"},
		{"10000000", args{10000000}, "10,000,000"},
		{"100000000", args{100000000}, "100,000,000"},
		{"1000000000", args{1000000000}, "1,000,000,000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatNumber(tt.args.number); got != tt.want {
				t.Errorf("FormatNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatMetricNumber(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{1}, "1"},
		{"10", args{10}, "10"},
		{"100", args{100}, "100"},
		{"1000", args{1000}, "1k"},
		{"10000", args{10000}, "10k"},
		{"100000", args{100000}, "100k"},
		{"1000000", args{1000000}, "1m"},
		{"10000000", args{10000000}, "10m"},
		{"100000000", args{100000000}, "100m"},
		{"1000000000", args{1000000000}, "1b"},
		{"10000000000", args{10000000000}, "10b"},
		{"100000000000", args{100000000000}, "100b"},
		{"1000000000000", args{1000000000000}, "1t"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatMetricNumber(tt.args.number); got != tt.want {
				t.Errorf("FormatMetricNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMetricNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		// Success cases - plain numbers
		{"plain number: 1", "1", 1, false},
		{"plain number: 42", "42", 42, false},
		{"plain number: 999", "999", 999, false},

		// Success cases - k suffix (thousands)
		{"1k", "1k", 1000, false},
		{"10k", "10k", 10000, false},
		{"100k", "100k", 100000, false},
		{"2.5k", "2.5k", 2500, false},

		// Success cases - m suffix (millions)
		{"1m", "1m", 1000000, false},
		{"10m", "10m", 10000000, false},
		{"100m", "100m", 100000000, false},
		{"3.5m", "3.5m", 3500000, false},

		// Success cases - b suffix (billions)
		{"1b", "1b", 1000000000, false},
		{"10b", "10b", 10000000000, false},
		{"2.5b", "2.5b", 2500000000, false},

		// Success cases - t suffix (trillions)
		{"1t", "1t", 1000000000000, false},
		{"2t", "2t", 2000000000000, false},
		{"1.5t", "1.5t", 1500000000000, false},

		// Error cases - invalid input
		{"invalid: abc", "abc", 0, true},
		{"invalid: 10x", "10x", 0, true},
		{"invalid: empty", "", 0, true},
		{"invalid: k only", "k", 0, true},
		{"invalid: m only", "m", 0, true},
		{"invalid: multiple suffixes", "10km", 0, true},
		{"invalid: special chars", "10@k", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMetricNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMetricNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseMetricNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMetricNumber_RoundTrip(t *testing.T) {
	// Test that formatting and parsing are inverse operations
	numbers := []int{1, 42, 999, 1000, 10000, 100000, 1000000, 10000000, 1000000000}

	for _, num := range numbers {
		t.Run(fmt.Sprintf("roundtrip_%d", num), func(t *testing.T) {
			formatted := FormatMetricNumber(num)
			parsed, err := ParseMetricNumber(formatted)
			if err != nil {
				t.Errorf("ParseMetricNumber() error = %v", err)
				return
			}
			// Allow small rounding differences due to float conversion
			diff := parsed - num
			if diff < 0 {
				diff = -diff
			}
			if diff > num/100 { // Allow 1% difference
				t.Errorf("Round trip failed: %d -> %s -> %d (diff: %d)", num, formatted, parsed, diff)
			}
		})
	}
}
