package timex

import (
	"testing"
	"time"
)

func TestIntervalContainInclusive(t *testing.T) {
	start := FromStdTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	end := FromStdTime(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
	mid := FromStdTime(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))

	cases := []struct {
		name          string
		startIncluded bool
		endIncluded   bool
		probe         Time
		wantContained bool
	}{
		{"start-included", true, false, start, true},
		{"start-excluded", false, false, start, false},
		{"end-included", false, true, end, true},
		{"end-excluded", false, false, end, false},
		{"mid-always", false, false, mid, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			interval := NewInterval(start, end, tc.startIncluded, tc.endIncluded)
			if got := interval.Contain(tc.probe); got != tc.wantContained {
				t.Fatalf("Contain got %v want %v", got, tc.wantContained)
			}
		})
	}
}

func TestIntervalString(t *testing.T) {
	start := FromStdTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	end := FromStdTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))

	cases := []struct {
		name string
		in   Interval
		want string
	}{
		{
			name: "closed-open-finite",
			in:   NewInterval(start, end, true, false),
			want: "[" + start.String() + "," + end.String() + ")",
		},
		{
			name: "open-to-pos-inf",
			in:   NewInterval(start, NewPosInfTime(), false, false),
			want: "(" + start.String() + "," + PosInfTimeStr + ")",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.String(); got != tc.want {
				t.Fatalf("String got %q want %q", got, tc.want)
			}
		})
	}
}

func TestIntervalIsZero(t *testing.T) {
	finite := FromStdTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	cases := []struct {
		name string
		in   Interval
		want bool
	}{
		{"zero-value", Interval{}, true},
		{"finite-interval", NewInterval(finite, finite, false, false), false},
		{"included-bounds-only", NewInterval(Time{}, Time{}, true, false), false},
		{"both-infinite", NewInterval(NewNegInfTime(), NewPosInfTime(), false, false), false},
		{"start-infinite", NewInterval(NewNegInfTime(), Time{}, false, false), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.IsZero(); got != tc.want {
				t.Fatalf("IsZero got %v want %v", got, tc.want)
			}
		})
	}
}

func TestIntervalBeforeAfterInclusive(t *testing.T) {
	start := FromStdTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	end := FromStdTime(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))

	cases := []struct {
		name          string
		startIncluded bool
		endIncluded   bool
		probe         Time
		wantBefore    bool
		wantAfter     bool
	}{
		{"end-excluded", true, false, end, true, false},
		{"end-included", true, true, end, false, false},
		{"start-excluded", false, true, start, false, true},
		{"start-included", true, true, start, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			interval := NewInterval(start, end, tc.startIncluded, tc.endIncluded)
			if got := interval.Before(tc.probe); got != tc.wantBefore {
				t.Fatalf("Before got %v want %v", got, tc.wantBefore)
			}
			if got := interval.After(tc.probe); got != tc.wantAfter {
				t.Fatalf("After got %v want %v", got, tc.wantAfter)
			}
		})
	}
}
