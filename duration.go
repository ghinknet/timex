package timex

import (
	"math"
	"time"

	"go.gh.ink/toolbox/expr"
)

func FromStdDuration(stdDuration time.Duration) Duration {
	return Duration{
		std: stdDuration,
		inf: FiniteTime,
	}
}

func (d Duration) ToStdDuration() (stdDuration time.Duration, inf InfFlag) {
	if d.inf > 0 {
		return time.Duration(math.MaxInt64), PosInfTime
	}
	if d.inf < 0 {
		return time.Duration(math.MinInt64), NegInfTime
	}
	return d.std, FiniteTime
}

func NewPosInfDuration() Duration {
	return Duration{
		std: time.Duration(math.MaxInt64),
		inf: PosInfTime,
	}
}

func NewNegInfDuration() Duration {
	return Duration{
		std: time.Duration(math.MinInt64),
		inf: NegInfTime,
	}
}

func (d Duration) Hours() (hours float64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			math.MaxFloat64,
			-math.MaxFloat64,
		), d.inf
	}
	return d.std.Hours(), FiniteTime
}

func (d Duration) Minutes() (minutes float64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			math.MaxFloat64,
			-math.MaxFloat64,
		), d.inf
	}
	return d.std.Minutes(), FiniteTime
}

func (d Duration) Seconds() (seconds float64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			math.MaxFloat64,
			-math.MaxFloat64,
		), d.inf
	}
	return d.std.Seconds(), FiniteTime
}

func (d Duration) Milliseconds() (milliseconds int64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			int64(math.MaxInt64),
			math.MinInt64,
		), d.inf
	}
	return d.std.Milliseconds(), FiniteTime
}

func (d Duration) Microseconds() (microseconds int64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			int64(math.MaxInt64),
			math.MinInt64,
		), d.inf
	}
	return d.std.Microseconds(), FiniteTime
}

func (d Duration) Nanoseconds() (nanoseconds int64, inf InfFlag) {
	if d.inf != FiniteTime {
		return expr.Ternary(
			d.inf > 0,
			int64(math.MaxInt64),
			math.MinInt64,
		), d.inf
	}
	return d.std.Nanoseconds(), FiniteTime
}

func (d Duration) Round(m Duration) (Duration, error) {
	if d.inf != FiniteTime {
		return d, nil
	}
	if m.inf != FiniteTime {
		return Duration{}, ErrInvalidInfiniteOp
	}

	return FromStdDuration(d.std.Round(m.std)), nil
}

func (d Duration) Truncate(m Duration) (Duration, error) {
	if d.inf != FiniteTime {
		return d, nil
	}
	if m.inf != FiniteTime {
		return Duration{}, ErrInvalidInfiniteOp
	}

	return FromStdDuration(d.std.Truncate(m.std)), nil
}

func (d Duration) Abs() (Duration, error) {
	if d.inf != FiniteTime {
		return NewPosInfDuration(), nil
	}

	return FromStdDuration(d.std.Abs()), nil
}

func (d Duration) String() string {
	if d.inf != FiniteTime {
		return expr.Ternary(d.inf > 0, PosInfTimeStr, NegInfTimeStr)
	}
	return d.std.String()
}
