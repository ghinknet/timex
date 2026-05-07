package timex

import "time"

type InfFlag int8

const (
	FiniteTime InfFlag = 0
	PosInfTime         = 1
	NegInfTime         = -1
)

const (
	PosInfTimeStr = "positive infinite"
	NegInfTimeStr = "negative infinite"
)

// Time provides the basic time point struct,
// ext storages the time point in ns since 1970-01-01 00:00:00 UTC,
// infinite indicates whether the time point is infinite (positive or negative),
// 0 stands for normal time point, 1 stands for positive infinite, -1 stands for negative infinite
type Time struct {
	std time.Time
	inf InfFlag
}

// Duration provides the basic time duration struct,
// ext storages the time duration length in ns,
// infinite indicates whether the time duration is infinite (positive or negative),
// 0 stands for normal time point, 1 stands for positive infinite, -1 stands for negative infinite
type Duration struct {
	std time.Duration
	inf InfFlag
}

// Interval provides the basic time interval struct,
// start storages the time start, and end storages the time end,
// startIncluded indicates whether the time start is included in the interval,
// endIncluded indicates whether the time end is included in the interval
type Interval struct {
	start         Time
	startIncluded bool
	end           Time
	endIncluded   bool
}
