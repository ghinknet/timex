package timex

import (
	"strings"

	"go.gh.ink/toolbox/expr"
)

func NewInterval(start, end Time, startIncluded, endIncluded bool) Interval {
	return Interval{
		start:         start,
		startIncluded: startIncluded,
		end:           end,
		endIncluded:   endIncluded,
	}
}

func (i Interval) Before(t Time) bool {
	if i.end.Before(t) {
		return true
	}
	if i.end.Equal(t) {
		return !i.endIncluded
	}
	return false
}

func (i Interval) After(t Time) bool {
	if i.start.After(t) {
		return true
	}
	if i.start.Equal(t) {
		return !i.startIncluded
	}
	return false
}

func (i Interval) Contain(t Time) bool {
	leftOk := i.start.Before(t) || (i.startIncluded && i.start.Equal(t))
	rightOk := i.end.After(t) || (i.endIncluded && i.end.Equal(t))
	return leftOk && rightOk
}

func (i Interval) Start() (start Time, included bool) {
	return i.start, i.startIncluded
}

func (i Interval) End() (end Time, included bool) {
	return i.end, i.endIncluded
}

// IsZero reports whether the interval is the zero value: both bounds are the
// zero (finite) time and both bounds excluded. It lets reflection-based
// zero-checkers (e.g. ORM helpers) treat Interval as an opaque value instead of
// recursing into its unexported fields.
func (i Interval) IsZero() bool {
	return i.start.IsZero() && i.end.IsZero() && !i.startIncluded && !i.endIncluded
}

// String renders the interval in mathematical notation using each bound's
// human-readable String form (Go's default time layout, or the infinity
// sentinels) — e.g. "[2024-01-01 00:00:00 +0000 UTC,positive infinite)".
// Brackets encode inclusivity: "[" / "]" included, "(" / ")" excluded. For the
// machine-parseable RFC 3339 form, use MarshalText.
func (i Interval) String() string {
	var b strings.Builder
	b.WriteString(expr.Ternary(i.startIncluded, "[", "("))
	b.WriteString(i.start.String())
	b.WriteByte(',')
	b.WriteString(i.end.String())
	b.WriteString(expr.Ternary(i.endIncluded, "]", ")"))
	return b.String()
}
