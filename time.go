package timex

import (
	"math"
	"strings"
	"time"

	"go.gh.ink/toolbox/expr"
)

func FromStdTime(stdTime time.Time) Time {
	return Time{
		std: stdTime,
		inf: FiniteTime,
	}
}

func Now() Time {
	return FromStdTime(time.Now())
}

func Date(year, month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return FromStdTime(time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc))
}

func Unix(sec int64, nsec int64) Time {
	return FromStdTime(time.Unix(sec, nsec))
}

func UnixMilli(msec int64) Time {
	return FromStdTime(time.UnixMilli(msec))
}

func UnixMicro(usec int64) Time {
	return FromStdTime(time.UnixMicro(usec))
}

func Parse(layout, value string) (Time, error) {
	if hasPattern(layout) {
		layout = convertLayout(layout)
	}

	stdTime, err := time.Parse(layout, value)
	if err != nil {
		return Time{}, err
	}

	return FromStdTime(stdTime), nil
}

func ParseInLocation(layout, value string, loc *time.Location) (Time, error) {
	if hasPattern(layout) {
		layout = convertLayout(layout)
	}

	stdTime, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return Time{}, err
	}

	return FromStdTime(stdTime), nil
}

func NewPosInfTime() Time {
	return Time{
		std: time.Unix(0, math.MaxInt64),
		inf: PosInfTime,
	}
}

func NewNegInfTime() Time {
	return Time{
		std: time.Unix(0, math.MinInt64),
		inf: NegInfTime,
	}
}

func Since(t Time) Duration {
	return Now().Sub(t)
}

func Until(t Time) Duration {
	return t.Sub(Now())
}

func (t Time) ToStdTime() (stdTime time.Time, inf InfFlag) {
	if t.inf > 0 {
		return time.Unix(0, math.MaxInt64), PosInfTime
	}
	if t.inf < 0 {
		return time.Unix(0, math.MinInt64), NegInfTime
	}
	return t.std, FiniteTime
}

func (t Time) IsZero() bool {
	return t.std.IsZero() && t.inf == FiniteTime
}

func (t Time) After(u Time) bool {
	tTime, tInf := t.ToStdTime()
	uTime, uInf := u.ToStdTime()

	if tInf == uInf {
		if tInf == FiniteTime {
			return tTime.After(uTime)
		}
		return false
	}
	return tInf > uInf
}

func (t Time) Before(u Time) bool {
	tTime, tInf := t.ToStdTime()
	uTime, uInf := u.ToStdTime()

	if tInf == uInf {
		if tInf == FiniteTime {
			return tTime.Before(uTime)
		}
		return false
	}
	return tInf < uInf
}

func (t Time) Compare(u Time) int {
	if t.After(u) {
		return 1
	} else if t.Before(u) {
		return -1
	}
	return 0
}

func (t Time) Equal(u Time) bool {
	tTime, tInf := t.ToStdTime()
	uTime, uInf := u.ToStdTime()

	if tInf == uInf {
		if tInf == FiniteTime {
			return tTime.Equal(uTime)
		}
		return true
	}
	return false
}

func (t Time) Unix() (stamp int64, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Unix(), inf
}

func (t Time) UnixMilli() (stamp int64, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.UnixMilli(), inf
}

func (t Time) UnixMicro() (stamp int64, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.UnixMicro(), inf
}

func (t Time) UnixNano() (stamp int64, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.UnixNano(), inf
}

func (t Time) Add(d Duration) (Time, error) {
	if d.inf == t.inf {
		if t.inf == FiniteTime {
			return FromStdTime(t.std.Add(d.std)), nil
		}
		return t, nil
	}

	if t.inf != FiniteTime && d.inf != FiniteTime {
		return Time{}, ErrInvalidInfiniteOp
	}

	return Time{
		std: time.Unix(0, 0),
		inf: d.inf + t.inf,
	}, nil
}

func (t Time) AddDate(years, months, days int) Time {
	if t.inf != FiniteTime {
		return t
	}
	return FromStdTime(t.std.AddDate(years, months, days))
}

func (t Time) AddDateEOM(years, months, days int) Time {
	if t.inf != FiniteTime {
		return t
	}

	y, m, d := t.std.Date()
	hour, minute, sec := t.std.Clock()
	nsec := t.std.Nanosecond()
	loc := t.std.Location()

	lastOfMonth := time.Date(y, m+1, 0, 0, 0, 0, 0, loc).Day()
	isEndOfMonth := d == lastOfMonth

	// Compute target month/year without day overflow.
	target := time.Date(y+years, m+time.Month(months), 1, hour, minute, sec, nsec, loc)
	lastOfTarget := time.Date(target.Year(), target.Month()+1, 0, 0, 0, 0, 0, loc).Day()

	day := d
	if isEndOfMonth || d > lastOfTarget {
		day = lastOfTarget
	}

	base := time.Date(target.Year(), target.Month(), day, hour, minute, sec, nsec, loc)
	if days != 0 {
		base = base.AddDate(0, 0, days)
	}

	return FromStdTime(base)
}

func (t Time) Sub(u Time) Duration {
	tTime, tInf := t.ToStdTime()
	uTime, uInf := u.ToStdTime()

	if tInf == uInf {
		if tInf == FiniteTime {
			return FromStdDuration(tTime.Sub(uTime))
		}
		return FromStdDuration(0)
	}

	return expr.Ternary(tInf-uInf > 0, NewPosInfDuration(), NewNegInfDuration())
}

func (t Time) Clock() (hour, min, sec int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	h, m, s := stdTime.Clock()
	return h, m, s, inf
}

func (t Time) Date() (year int, month int, day int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	y, m, d := stdTime.Date()
	return y, int(m), d, inf
}

func (t Time) ISOWeek() (year, week int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	y, w := stdTime.ISOWeek()
	return y, w, inf
}

func (t Time) IsDst() bool {
	if t.inf != FiniteTime {
		return false
	}
	return t.std.IsDST()
}

func (t Time) Weekday() (weekday int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return int(stdTime.Weekday()), inf
}

func (t Time) Year() (year int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Year(), inf
}

func (t Time) Month() (month int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return int(stdTime.Month()), inf
}

func (t Time) Day() (day int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Day(), inf
}

func (t Time) Hour() (hour int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Hour(), inf
}

func (t Time) Minute() (minute int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Minute(), inf
}

func (t Time) Second() (second int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Second(), inf
}

func (t Time) Nanosecond() (nanosecond int, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	return stdTime.Nanosecond(), inf
}

func (t Time) In(loc *time.Location) Time {
	return Time{
		std: t.std.In(loc),
		inf: t.inf,
	}
}

func (t Time) Round(d Duration) (Time, error) {
	if t.inf != FiniteTime {
		return t, nil
	}
	if d.inf != FiniteTime {
		return Time{}, ErrInvalidInfiniteOp
	}
	return FromStdTime(t.std.Round(d.std)), nil
}

func (t Time) Truncate(d Duration) (Time, error) {
	if t.inf != FiniteTime {
		return t, nil
	}
	if d.inf != FiniteTime {
		return Time{}, ErrInvalidInfiniteOp
	}
	return FromStdTime(t.std.Truncate(d.std)), nil
}

func (t Time) Local() Time {
	if t.inf != FiniteTime {
		return t
	}
	return FromStdTime(t.std.Local())
}

func (t Time) Location() *time.Location {
	return t.std.Location()
}

func (t Time) UTC() Time {
	if t.inf != FiniteTime {
		return t
	}
	return FromStdTime(t.std.UTC())
}

func (t Time) Zone() (name string, offset int) {
	return t.std.Zone()
}

func (t Time) ZoneBounds() (interval Interval, inf InfFlag) {
	stdTime, inf := t.ToStdTime()
	if inf != FiniteTime {
		return Interval{
			start: NewNegInfTime(),
			end:   NewPosInfTime(),
		}, inf
	}

	s, e := stdTime.ZoneBounds()
	if s == e {
		return Interval{
			start: NewNegInfTime(),
			end:   NewPosInfTime(),
		}, FiniteTime
	}

	return Interval{
		start:         FromStdTime(s),
		startIncluded: true,
		end:           FromStdTime(e),
		endIncluded:   false,
	}, FiniteTime
}

func (t Time) Format(layout string) string {
	if hasPattern(layout) {
		layout = convertLayout(layout)
	}
	return t.std.Format(layout)
}

func (t Time) String() string {
	if t.inf != FiniteTime {
		return expr.Ternary(t.inf > 0, PosInfTimeStr, NegInfTimeStr)
	}
	return t.std.String()
}

func (t Time) GoString() string {
	return strings.Replace(t.std.GoString(), "time.Time", "timex.Time", 1)
}
