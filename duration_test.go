package timex

import (
	"math"
	"testing"
)

func TestDurationInfiniteUnits(t *testing.T) {
	cases := []struct {
		name   string
		d      Duration
		fval   float64
		ival   int64
		inf    InfFlag
		fcheck func(Duration) (float64, InfFlag)
		icheck func(Duration) (int64, InfFlag)
	}{
		{
			name: "pos-hours",
			d:    NewPosInfDuration(),
			fval: math.MaxFloat64,
			inf:  PosInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Hours()
			},
		},
		{
			name: "neg-hours",
			d:    NewNegInfDuration(),
			fval: -math.MaxFloat64,
			inf:  NegInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Hours()
			},
		},
		{
			name: "pos-minutes",
			d:    NewPosInfDuration(),
			fval: math.MaxFloat64,
			inf:  PosInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Minutes()
			},
		},
		{
			name: "neg-minutes",
			d:    NewNegInfDuration(),
			fval: -math.MaxFloat64,
			inf:  NegInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Minutes()
			},
		},
		{
			name: "pos-seconds",
			d:    NewPosInfDuration(),
			fval: math.MaxFloat64,
			inf:  PosInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Seconds()
			},
		},
		{
			name: "neg-seconds",
			d:    NewNegInfDuration(),
			fval: -math.MaxFloat64,
			inf:  NegInfTime,
			fcheck: func(d Duration) (float64, InfFlag) {
				return d.Seconds()
			},
		},
		{
			name: "pos-milliseconds",
			d:    NewPosInfDuration(),
			ival: math.MaxInt64,
			inf:  PosInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Milliseconds()
			},
		},
		{
			name: "neg-milliseconds",
			d:    NewNegInfDuration(),
			ival: math.MinInt64,
			inf:  NegInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Milliseconds()
			},
		},
		{
			name: "pos-microseconds",
			d:    NewPosInfDuration(),
			ival: math.MaxInt64,
			inf:  PosInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Microseconds()
			},
		},
		{
			name: "neg-microseconds",
			d:    NewNegInfDuration(),
			ival: math.MinInt64,
			inf:  NegInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Microseconds()
			},
		},
		{
			name: "pos-nanoseconds",
			d:    NewPosInfDuration(),
			ival: math.MaxInt64,
			inf:  PosInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Nanoseconds()
			},
		},
		{
			name: "neg-nanoseconds",
			d:    NewNegInfDuration(),
			ival: math.MinInt64,
			inf:  NegInfTime,
			icheck: func(d Duration) (int64, InfFlag) {
				return d.Nanoseconds()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fcheck != nil {
				val, inf := tc.fcheck(tc.d)
				if val != tc.fval || inf != tc.inf {
					t.Fatalf("got %v,%v want %v,%v", val, inf, tc.fval, tc.inf)
				}
				return
			}
			if tc.icheck != nil {
				val, inf := tc.icheck(tc.d)
				if val != tc.ival || inf != tc.inf {
					t.Fatalf("got %v,%v want %v,%v", val, inf, tc.ival, tc.inf)
				}
				return
			}
			t.Fatalf("missing check")
		})
	}
}

