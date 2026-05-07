package timex

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
