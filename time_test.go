package timex

import (
	"testing"
	"time"
)

func TestTimeLocalUTCPreserveInf(t *testing.T) {
	cases := []struct {
		name string
		t    Time
		inf  InfFlag
	}{
		{name: "pos", t: NewPosInfTime(), inf: PosInfTime},
		{name: "neg", t: NewNegInfTime(), inf: NegInfTime},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local := tc.t.Local()
			if local.inf != tc.inf {
				t.Fatalf("Local inf got %v want %v", local.inf, tc.inf)
			}
			utc := tc.t.UTC()
			if utc.inf != tc.inf {
				t.Fatalf("UTC inf got %v want %v", utc.inf, tc.inf)
			}
		})
	}
}

func TestTimeAddDateEOM(t *testing.T) {
	cases := []struct {
		name   string
		base   Time
		years  int
		months int
		days   int
		want   Time
	}{
		{
			name:   "eom-leap-feb",
			base:   Date(2024, 1, 31, 10, 20, 30, 123000000, time.UTC),
			years:  0,
			months: 1,
			days:   0,
			want:   Date(2024, 2, 29, 10, 20, 30, 123000000, time.UTC),
		},
		{
			name:   "eom-nonleap-feb",
			base:   Date(2023, 1, 31, 10, 20, 30, 0, time.UTC),
			years:  0,
			months: 1,
			days:   0,
			want:   Date(2023, 2, 28, 10, 20, 30, 0, time.UTC),
		},
		{
			name:   "normal-day",
			base:   Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
			years:  0,
			months: 1,
			days:   0,
			want:   Date(2024, 2, 15, 8, 0, 0, 0, time.UTC),
		},
		{
			name:   "overflow-to-eom",
			base:   Date(2024, 1, 30, 9, 0, 0, 0, time.UTC),
			years:  0,
			months: 1,
			days:   0,
			want:   Date(2024, 2, 29, 9, 0, 0, 0, time.UTC),
		},
		{
			name:   "eom-plus-days",
			base:   Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			years:  0,
			months: 1,
			days:   1,
			want:   Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.base.AddDateEOM(tc.years, tc.months, tc.days)
			if !got.Equal(tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestTimeAddDateEOMInf(t *testing.T) {
	cases := []struct {
		name string
		in   Time
		inf  InfFlag
	}{
		{name: "pos", in: NewPosInfTime(), inf: PosInfTime},
		{name: "neg", in: NewNegInfTime(), inf: NegInfTime},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.AddDateEOM(1, 2, 3)
			if got.inf != tc.inf {
				t.Fatalf("inf got %v want %v", got.inf, tc.inf)
			}
		})
	}
}
