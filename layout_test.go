package timex

import (
	"testing"
	"time"
)

func TestLayoutYYYYMMDDParse(t *testing.T) {
	got, err := Parse("yyyy-MM-dd", "2024-02-29")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	y, m, d, inf := got.Date()
	if inf != FiniteTime {
		t.Fatalf("expected finite time, got %v", inf)
	}
	if y != 2024 || m != 2 || d != 29 {
		t.Fatalf("unexpected date: %d-%02d-%02d", y, m, d)
	}
}

func TestLayoutYYYYMMDDFormat(t *testing.T) {
	std := time.Date(2023, 7, 9, 0, 0, 0, 0, time.UTC)
	got := FromStdTime(std).Format("yyyy-MM-dd")
	if got != "2023-07-09" {
		t.Fatalf("Format got %q want %q", got, "2023-07-09")
	}
}

func TestLayoutCommonFormats(t *testing.T) {
	base := time.Date(2024, 1, 2, 15, 4, 5, 678000000, time.UTC)
	cases := []struct {
		name   string
		layout string
		want   string
	}{
		{name: "date-slash", layout: "yyyy/MM/dd", want: "2024/01/02"},
		{name: "date-time", layout: "yyyy-MM-dd HH:mm:ss", want: "2024-01-02 15:04:05"},
		{name: "date-time-t", layout: "yyyy-MM-dd'T'HH:mm:ss", want: "2024-01-02T15:04:05"},
		{name: "millis", layout: "yyyy-MM-dd HH:mm:ss.SSS", want: "2024-01-02 15:04:05.678"},
		{name: "micros", layout: "yyyy-MM-dd HH:mm:ss.SSSSSS", want: "2024-01-02 15:04:05.678000"},
		{name: "millis-zone", layout: "yyyy-MM-dd'T'HH:mm:ss.SSSZ", want: "2024-01-02T15:04:05.678+0000"},
		{name: "weekday-short", layout: "EEE, MMM dd yyyy", want: "Tue, Jan 02 2024"},
		{name: "weekday-long", layout: "EEEE, MMMM dd yyyy", want: "Tuesday, January 02 2024"},
		{name: "month-short", layout: "dd MMM yyyy", want: "02 Jan 2024"},
		{name: "month-long", layout: "dd MMMM yyyy", want: "02 January 2024"},
		{name: "zone-name", layout: "yyyy-MM-dd HH:mm:ss z", want: "2024-01-02 15:04:05 UTC"},
		{name: "12h", layout: "yyyy-MM-dd hh:mm:ss a", want: "2024-01-02 03:04:05 PM"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FromStdTime(base).Format(tc.layout)
			if got != tc.want {
				t.Fatalf("Format got %q want %q", got, tc.want)
			}
		})
	}
}

func TestLayoutCommonParses(t *testing.T) {
	cases := []struct {
		name     string
		layout   string
		value    string
		year     int
		month    int
		day      int
		hour     int
		minute   int
		second   int
		nsec     int
		zoneOff  int
		hasZone  bool
	}{
		{
			name:   "date-slash",
			layout: "yyyy/MM/dd",
			value:  "2024/01/02",
			year:   2024, month: 1, day: 2,
		},
		{
			name:   "date-time",
			layout: "yyyy-MM-dd HH:mm:ss",
			value:  "2024-01-02 15:04:05",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5,
		},
		{
			name:   "weekday-short",
			layout: "EEE, MMM dd yyyy",
			value:  "Tue, Jan 02 2024",
			year:   2024, month: 1, day: 2,
		},
		{
			name:   "month-short",
			layout: "dd MMM yyyy",
			value:  "02 Feb 2024",
			year:   2024, month: 2, day: 2,
		},
		{
			name:   "month-long",
			layout: "dd MMMM yyyy",
			value:  "02 February 2024",
			year:   2024, month: 2, day: 2,
		},
		{
			name:   "12h-pm",
			layout: "yyyy-MM-dd hh:mm:ss a",
			value:  "2024-01-02 03:04:05 PM",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5,
		},
		{
			name:   "12h-am",
			layout: "yyyy-MM-dd hh:mm:ss a",
			value:  "2024-01-02 03:04:05 AM",
			year:   2024, month: 1, day: 2, hour: 3, minute: 4, second: 5,
		},
		{
			name:   "millis",
			layout: "yyyy-MM-dd HH:mm:ss.SSS",
			value:  "2024-01-02 15:04:05.678",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5, nsec: 678000000,
		},
		{
			name:   "micros",
			layout: "yyyy-MM-dd HH:mm:ss.SSSSSS",
			value:  "2024-01-02 15:04:05.678901",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5, nsec: 678901000,
		},
		{
			name:   "millis-zone",
			layout: "yyyy-MM-dd'T'HH:mm:ss.SSSZ",
			value:  "2024-01-02T15:04:05.678+0800",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5, nsec: 678000000,
			zoneOff: 8 * 3600,
			hasZone: true,
		},
		{
			name:   "zone-name",
			layout: "yyyy-MM-dd HH:mm:ss z",
			value:  "2024-01-02 15:04:05 UTC",
			year:   2024, month: 1, day: 2, hour: 15, minute: 4, second: 5,
			hasZone: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.layout, tc.value)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			y, m, d, inf := got.Date()
			if inf != FiniteTime {
				t.Fatalf("expected finite time, got %v", inf)
			}
			if y != tc.year || m != tc.month || d != tc.day {
				t.Fatalf("unexpected date: %d-%02d-%02d", y, m, d)
			}
			h, min, s, inf := got.Clock()
			if inf != FiniteTime {
				t.Fatalf("expected finite time, got %v", inf)
			}
			if h != tc.hour || min != tc.minute || s != tc.second {
				t.Fatalf("unexpected clock: %02d:%02d:%02d", h, min, s)
			}
			nsec, inf := got.Nanosecond()
			if inf != FiniteTime {
				t.Fatalf("expected finite time, got %v", inf)
			}
			if nsec != tc.nsec {
				t.Fatalf("unexpected nsec: %d", nsec)
			}
			if tc.hasZone {
				_, off := got.Zone()
				if tc.zoneOff != 0 && off != tc.zoneOff {
					t.Fatalf("unexpected zone offset: %d", off)
				}
			}
		})
	}
}
