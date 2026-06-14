package timex

import "errors"

var ErrInvalidInfiniteOp = errors.New("timex: cannot add ±∞ to ∓∞")

var ErrInvalidInterval = errors.New("timex: invalid interval text representation")

var ErrInvalidBinary = errors.New("timex: invalid binary representation")
