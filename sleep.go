package timex

import "time"

func Sleep(d Duration) {
	if d.inf == PosInfTime {
		select {}
	}
	if d.inf == NegInfTime {
		return
	}

	time.Sleep(d.std)
}
