package stringutil

import "testing"

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
