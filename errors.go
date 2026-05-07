package timex

import "errors"

var ErrInvalidInfiniteOp = errors.New("timex: cannot add ±∞ to ∓∞")
