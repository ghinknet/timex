package timex

import "testing"

func TestTimeLocalUTCPreserveInf(t *testing.T) {
	cases := []struct {
		name string
		t    Time
		inf  InfFlag
	}{
		{name: "pos", t: NewPosInfTime(), inf: PosInfTime},
		{name: "neg", t: NewNegInfTime(), inf: NegInfTime},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local := tc.t.Local()
			if local.inf != tc.inf {
				t.Fatalf("Local inf got %v want %v", local.inf, tc.inf)
			}
			utc := tc.t.UTC()
			if utc.inf != tc.inf {
				t.Fatalf("UTC inf got %v want %v", utc.inf, tc.inf)
			}
		})
	}
}

