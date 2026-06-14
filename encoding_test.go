package timex

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"
	"time"
)

// finite reference points reused across the encoding tests.
var (
	encStart = Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	encEnd   = Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
)

func TestTimeJSONRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Time
		want string
	}{
		{"finite", Date(2024, 1, 2, 15, 4, 5, 0, time.UTC), `"2024-01-02T15:04:05Z"`},
		{"pos-inf", NewPosInfTime(), `"positive infinite"`},
		{"neg-inf", NewNegInfTime(), `"negative infinite"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal err %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("Marshal got %s want %s", got, tc.want)
			}

			var back Time
			if err := json.Unmarshal(got, &back); err != nil {
				t.Fatalf("Unmarshal err %v", err)
			}
			if !back.Equal(tc.in) {
				t.Fatalf("round trip got %v want %v", back, tc.in)
			}
		})
	}
}

func TestDurationJSONRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Duration
		want string
	}{
		{"finite", FromStdDuration(2*time.Hour + 30*time.Minute), `"2h30m0s"`},
		{"zero", FromStdDuration(0), `"0s"`},
		{"pos-inf", NewPosInfDuration(), `"positive infinite"`},
		{"neg-inf", NewNegInfDuration(), `"negative infinite"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal err %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("Marshal got %s want %s", got, tc.want)
			}

			var back Duration
			if err := json.Unmarshal(got, &back); err != nil {
				t.Fatalf("Unmarshal err %v", err)
			}
			if back.std != tc.in.std || back.inf != tc.in.inf {
				t.Fatalf("round trip got %v,%v want %v,%v", back.std, back.inf, tc.in.std, tc.in.inf)
			}
		})
	}
}

func TestIntervalJSONRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Interval
		want string
	}{
		{
			name: "closed-open",
			in:   NewInterval(encStart, encEnd, true, false),
			want: `"[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)"`,
		},
		{
			name: "open-closed",
			in:   NewInterval(encStart, encEnd, false, true),
			want: `"(2024-01-01T00:00:00Z,2024-04-01T00:00:00Z]"`,
		},
		{
			name: "closed-closed",
			in:   NewInterval(encStart, encEnd, true, true),
			want: `"[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z]"`,
		},
		{
			name: "open-open",
			in:   NewInterval(encStart, encEnd, false, false),
			want: `"(2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)"`,
		},
		{
			name: "half-open-to-pos-inf",
			in:   NewInterval(encStart, NewPosInfTime(), true, false),
			want: `"[2024-01-01T00:00:00Z,positive infinite)"`,
		},
		{
			name: "neg-inf-to-finite",
			in:   NewInterval(NewNegInfTime(), encEnd, false, true),
			want: `"(negative infinite,2024-04-01T00:00:00Z]"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal err %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("Marshal got %s want %s", got, tc.want)
			}

			var back Interval
			if err := json.Unmarshal(got, &back); err != nil {
				t.Fatalf("Unmarshal err %v", err)
			}
			assertIntervalEqual(t, back, tc.in)
		})
	}
}

func TestIntervalUnmarshalTextLenient(t *testing.T) {
	// Whitespace around endpoints and the whole value is tolerated.
	var iv Interval
	if err := iv.UnmarshalText([]byte("  [ 2024-01-01T00:00:00Z , 2024-04-01T00:00:00Z ) ")); err != nil {
		t.Fatalf("UnmarshalText err %v", err)
	}
	assertIntervalEqual(t, iv, NewInterval(encStart, encEnd, true, false))
}

func TestIntervalUnmarshalTextInvalid(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"too-short", "["},
		{"no-open-bracket", "2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)"},
		{"no-close-bracket", "[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z"},
		{"no-comma", "[2024-01-01T00:00:00Z]"},
		{"bad-start", "[not-a-time,2024-04-01T00:00:00Z)"},
		{"bad-end", "[2024-01-01T00:00:00Z,not-a-time)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var iv Interval
			if err := iv.UnmarshalText([]byte(tc.in)); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

// TestNestedStructJSON exercises the headline use case: the timex types embedded
// as fields of a user struct round-trip cleanly through encoding/json.
func TestNestedStructJSON(t *testing.T) {
	type schedule struct {
		Name   string   `json:"name"`
		Window Interval `json:"window"`
		Every  Duration `json:"every"`
		Until  Time     `json:"until"`
	}

	in := schedule{
		Name:   "Q1",
		Window: NewInterval(encStart, encEnd, true, false),
		Every:  FromStdDuration(24 * time.Hour),
		Until:  NewPosInfTime(),
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal err %v", err)
	}

	want := `{"name":"Q1","window":"[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)","every":"24h0m0s","until":"positive infinite"}`
	if string(data) != want {
		t.Fatalf("Marshal got %s want %s", data, want)
	}

	var back schedule
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal err %v", err)
	}
	if back.Name != in.Name {
		t.Fatalf("Name got %q want %q", back.Name, in.Name)
	}
	assertIntervalEqual(t, back.Window, in.Window)
	if back.Every.std != in.Every.std || back.Every.inf != in.Every.inf {
		t.Fatalf("Every got %v want %v", back.Every, in.Every)
	}
	if !back.Until.Equal(in.Until) {
		t.Fatalf("Until got %v want %v", back.Until, in.Until)
	}
}

func TestTimeBinaryRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Time
	}{
		{"finite-utc", Date(2024, 1, 2, 15, 4, 5, 678901234, time.UTC)},
		{"finite-zone", Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("X", 8*3600))},
		{"pos-inf", NewPosInfTime()},
		{"neg-inf", NewNegInfTime()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.in.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary err %v", err)
			}
			var back Time
			if err := back.UnmarshalBinary(b); err != nil {
				t.Fatalf("UnmarshalBinary err %v", err)
			}
			if !back.Equal(tc.in) || back.inf != tc.in.inf {
				t.Fatalf("round trip got %v want %v", back, tc.in)
			}
		})
	}
}

func TestDurationBinaryRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Duration
	}{
		{"positive", FromStdDuration(2*time.Hour + 30*time.Minute)},
		{"negative", FromStdDuration(-90 * time.Second)},
		{"zero", FromStdDuration(0)},
		{"pos-inf", NewPosInfDuration()},
		{"neg-inf", NewNegInfDuration()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.in.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary err %v", err)
			}
			var back Duration
			if err := back.UnmarshalBinary(b); err != nil {
				t.Fatalf("UnmarshalBinary err %v", err)
			}
			if back.std != tc.in.std || back.inf != tc.in.inf {
				t.Fatalf("round trip got %v,%v want %v,%v", back.std, back.inf, tc.in.std, tc.in.inf)
			}
		})
	}
}

func TestIntervalBinaryRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   Interval
	}{
		{"closed-open", NewInterval(encStart, encEnd, true, false)},
		{"open-closed", NewInterval(encStart, encEnd, false, true)},
		{"closed-closed", NewInterval(encStart, encEnd, true, true)},
		{"both-inf", NewInterval(NewNegInfTime(), NewPosInfTime(), false, false)},
		{"half-open-inf", NewInterval(encStart, NewPosInfTime(), true, false)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.in.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary err %v", err)
			}
			var back Interval
			if err := back.UnmarshalBinary(b); err != nil {
				t.Fatalf("UnmarshalBinary err %v", err)
			}
			assertIntervalEqual(t, back, tc.in)
		})
	}
}

func TestBinaryUnmarshalInvalid(t *testing.T) {
	var ti Time
	if err := ti.UnmarshalBinary(nil); err == nil {
		t.Fatal("Time: expected error on empty input")
	}
	var d Duration
	if err := d.UnmarshalBinary([]byte{0, 1, 2}); err == nil { // finite flag, wrong length
		t.Fatal("Duration: expected error on short input")
	}
	var iv Interval
	if err := iv.UnmarshalBinary(nil); err == nil {
		t.Fatal("Interval: expected error on empty input")
	}
	if err := iv.UnmarshalBinary([]byte{0, 5}); err == nil { // start length 5 exceeds body
		t.Fatal("Interval: expected error on truncated body")
	}
}

// TestGobRoundTrip proves encoding/gob picks up the BinaryMarshaler pair: a
// struct of timex fields round-trips through gob without any registration.
func TestGobRoundTrip(t *testing.T) {
	type schedule struct {
		Window Interval
		Every  Duration
		Until  Time
	}

	in := schedule{
		Window: NewInterval(encStart, encEnd, true, false),
		Every:  FromStdDuration(24 * time.Hour),
		Until:  NewPosInfTime(),
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(in); err != nil {
		t.Fatalf("gob encode err %v", err)
	}
	var back schedule
	if err := gob.NewDecoder(&buf).Decode(&back); err != nil {
		t.Fatalf("gob decode err %v", err)
	}

	assertIntervalEqual(t, back.Window, in.Window)
	if back.Every.std != in.Every.std || back.Every.inf != in.Every.inf {
		t.Fatalf("Every got %v want %v", back.Every, in.Every)
	}
	if !back.Until.Equal(in.Until) || back.Until.inf != in.Until.inf {
		t.Fatalf("Until got %v want %v", back.Until, in.Until)
	}
}

func TestTimeValueScan(t *testing.T) {
	cases := []struct {
		name string
		in   Time
	}{
		{"finite", encStart},
		{"pos-inf", NewPosInfTime()},
		{"neg-inf", NewNegInfTime()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tc.in.Value()
			if err != nil {
				t.Fatalf("Value err %v", err)
			}
			s, ok := v.(string)
			if !ok {
				t.Fatalf("Value type = %T, want string", v)
			}
			// Drivers hand a text column back as string or []byte; both must scan.
			var fromStr, fromBytes Time
			if err := fromStr.Scan(s); err != nil {
				t.Fatalf("Scan(string) err %v", err)
			}
			if err := fromBytes.Scan([]byte(s)); err != nil {
				t.Fatalf("Scan([]byte) err %v", err)
			}
			if !fromStr.Equal(tc.in) || fromStr.inf != tc.in.inf {
				t.Fatalf("Scan(string) got %v want %v", fromStr, tc.in)
			}
			if !fromBytes.Equal(tc.in) || fromBytes.inf != tc.in.inf {
				t.Fatalf("Scan([]byte) got %v want %v", fromBytes, tc.in)
			}
		})
	}
}

func TestDurationValueScan(t *testing.T) {
	cases := []struct {
		name string
		in   Duration
	}{
		{"finite", FromStdDuration(90 * time.Minute)},
		{"pos-inf", NewPosInfDuration()},
		{"neg-inf", NewNegInfDuration()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tc.in.Value()
			if err != nil {
				t.Fatalf("Value err %v", err)
			}
			s, ok := v.(string)
			if !ok {
				t.Fatalf("Value type = %T, want string", v)
			}
			var fromStr, fromBytes Duration
			if err := fromStr.Scan(s); err != nil {
				t.Fatalf("Scan(string) err %v", err)
			}
			if err := fromBytes.Scan([]byte(s)); err != nil {
				t.Fatalf("Scan([]byte) err %v", err)
			}
			if fromStr.std != tc.in.std || fromStr.inf != tc.in.inf {
				t.Fatalf("Scan(string) got %v,%v want %v,%v", fromStr.std, fromStr.inf, tc.in.std, tc.in.inf)
			}
			if fromBytes.std != tc.in.std || fromBytes.inf != tc.in.inf {
				t.Fatalf("Scan([]byte) got %v,%v want %v,%v", fromBytes.std, fromBytes.inf, tc.in.std, tc.in.inf)
			}
		})
	}
}

func TestIntervalValueScan(t *testing.T) {
	cases := []struct {
		name string
		in   Interval
	}{
		{"closed-open", NewInterval(encStart, encEnd, true, false)},
		{"open-to-pos-inf", NewInterval(encStart, NewPosInfTime(), false, false)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tc.in.Value()
			if err != nil {
				t.Fatalf("Value err %v", err)
			}
			s, ok := v.(string)
			if !ok {
				t.Fatalf("Value type = %T, want string", v)
			}
			var fromStr, fromBytes Interval
			if err := fromStr.Scan(s); err != nil {
				t.Fatalf("Scan(string) err %v", err)
			}
			if err := fromBytes.Scan([]byte(s)); err != nil {
				t.Fatalf("Scan([]byte) err %v", err)
			}
			assertIntervalEqual(t, fromStr, tc.in)
			assertIntervalEqual(t, fromBytes, tc.in)
		})
	}
}

// TestScanNullAndInvalid checks the two edge cases every Scanner must get right:
// SQL NULL resets to the zero value, and an unsupported source type errors
// instead of panicking.
func TestScanNullAndInvalid(t *testing.T) {
	ti := encStart
	if err := ti.Scan(nil); err != nil || !ti.IsZero() {
		t.Fatalf("Time.Scan(nil): err=%v isZero=%v", err, ti.IsZero())
	}
	d := FromStdDuration(time.Hour)
	if err := d.Scan(nil); err != nil || !d.IsZero() {
		t.Fatalf("Duration.Scan(nil): err=%v isZero=%v", err, d.IsZero())
	}
	iv := NewInterval(encStart, encEnd, true, true)
	if err := iv.Scan(nil); err != nil || !iv.IsZero() {
		t.Fatalf("Interval.Scan(nil): err=%v isZero=%v", err, iv.IsZero())
	}

	var t2 Time
	if err := t2.Scan(12345); err == nil {
		t.Fatal("Time.Scan(int): expected error")
	}
	var iv2 Interval
	if err := iv2.Scan(3.14); err == nil {
		t.Fatal("Interval.Scan(float): expected error")
	}
}

func assertIntervalEqual(t *testing.T, got, want Interval) {
	t.Helper()
	gs, gsi := got.Start()
	ge, gei := got.End()
	ws, wsi := want.Start()
	we, wei := want.End()
	if !gs.Equal(ws) || gsi != wsi || !ge.Equal(we) || gei != wei {
		t.Fatalf("interval got %v want %v", got, want)
	}
}
