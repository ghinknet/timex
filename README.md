# timex

`timex` is an extension of the Go standard `time` library. It wraps `time.Time`
and `time.Duration` with first-class support for **positive / negative infinity**,
adds **end-of-month aware date arithmetic**, supports **Java `SimpleDateFormat`
style layouts**, and provides a small **time interval** type.

```
import "go.gh.ink/timex"
```

## Features

- **Infinite time & duration** — represent `+∞` / `-∞` time points and durations,
  with consistent comparison, arithmetic and formatting semantics.
- **End-of-month arithmetic** — `AddDateEOM` keeps month-end dates aligned to the
  last day of the target month (useful for financial schedules).
- **Java-style layouts** — parse and format with patterns such as
  `yyyy-MM-dd HH:mm:ss` in addition to Go's reference-time layouts.
- **Intervals** — model `[start, end]`, `(start, end)`, `[start, end)` ranges with
  inclusive/exclusive bounds and containment checks.
- **Serialization** — all types implement `encoding.TextMarshaler` /
  `TextUnmarshaler` (JSON, XML, TOML, map keys) and `encoding.BinaryMarshaler` /
  `BinaryUnmarshaler` (gob and other binary codecs), so they round-trip with no
  format-specific code.
- **Drop-in feel** — method names and signatures mirror the standard `time`
  package, with an extra `InfFlag` return value where infinity matters.

## Installation

```bash
go get go.gh.ink/timex
```

Requires Go `1.24.0` or later.

## Core Concepts

### `InfFlag`

Every value that may be infinite carries an `InfFlag`:

| Constant       | Value | Meaning           |
|----------------|-------|-------------------|
| `FiniteTime`   | `0`   | a finite value    |
| `PosInfTime`   | `1`   | positive infinity |
| `NegInfTime`   | `-1`  | negative infinity |

String representations are exposed as `PosInfTimeStr` (`"positive infinite"`) and
`NegInfTimeStr` (`"negative infinite"`).

Many accessor methods return the flag alongside the value, e.g.
`func (t Time) Year() (year int, inf InfFlag)`. When `inf != FiniteTime`, the
numeric result reflects the extreme bound used internally and should be ignored.

### Types

- `Time` — a time point that is finite or infinite.
- `Duration` — a duration that is finite or infinite.
- `Interval` — a range bounded by two `Time` values with inclusive/exclusive ends.

## Usage

### Creating a `Time`

```go
t := timex.Now()
t = timex.FromStdTime(time.Now())
t = timex.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
t = timex.Unix(1704207845, 0)
t = timex.UnixMilli(1704207845000)
t = timex.UnixMicro(1704207845000000)

posInf := timex.NewPosInfTime() // +∞
negInf := timex.NewNegInfTime() // -∞
```

### Parsing and formatting (Java-style layouts)

```go
t, err := timex.Parse("yyyy-MM-dd HH:mm:ss", "2024-01-02 15:04:05")
t, err = timex.ParseInLocation("yyyy-MM-dd'T'HH:mm:ssZ", "2024-01-02T15:04:05+0800", time.UTC)

t.Format("yyyy-MM-dd")              // "2024-01-02"
t.Format("yyyy-MM-dd HH:mm:ss.SSS") // "2024-01-02 15:04:05.678"
t.Format("EEEE, MMMM dd yyyy")      // "Tuesday, January 02 2024"
```

Standard Go reference layouts (e.g. `"2006-01-02"`) are still accepted; the
Java-style conversion only kicks in when a recognized pattern is present.

#### Supported pattern tokens

| Pattern   | Go layout    | Meaning                  |
|-----------|--------------|--------------------------|
| `yyyy`    | `2006`       | 4-digit year             |
| `yy`      | `06`         | 2-digit year             |
| `MMMM`    | `January`    | full month name          |
| `MMM`     | `Jan`        | short month name         |
| `MM`      | `01`         | 2-digit month            |
| `M`       | `1`          | month                    |
| `dd`      | `02`         | 2-digit day              |
| `d`       | `2`          | day                      |
| `HH`/`H`  | `15`         | hour (24h)               |
| `hh`      | `03`         | hour (12h, padded)       |
| `h`       | `3`          | hour (12h)               |
| `mm`/`m`  | `04`/`4`     | minute                   |
| `ss`/`s`  | `05`/`5`     | second                   |
| `SSSSSS`  | `000000`     | microseconds             |
| `SSS`     | `000`        | milliseconds             |
| `EEEE`    | `Monday`     | full weekday name        |
| `EEE`/`E` | `Mon`        | short weekday name       |
| `a`       | `PM`         | AM/PM marker             |
| `z`       | `MST`        | timezone name            |
| `Z`       | `-0700`      | timezone offset          |

### Comparing

```go
a.After(b)   // bool
a.Before(b)  // bool
a.Equal(b)   // bool
a.Compare(b) // -1, 0, or 1
a.IsZero()   // bool

timex.NewPosInfTime().After(timex.Now()) // true
```

### Arithmetic

```go
later, err := t.Add(timex.FromStdDuration(time.Hour))

// Standard calendar arithmetic (day overflows into the next month).
t.AddDate(0, 1, 0)

// End-of-month aware: 2024-01-31 + 1 month => 2024-02-29.
t.AddDateEOM(0, 1, 0)

// Difference between two times.
d := a.Sub(b)

d = timex.Since(t) // Now().Sub(t)
d = timex.Until(t) // t.Sub(Now())
```

Adding two opposite infinities (e.g. `+∞ + (-∞)`) returns `ErrInvalidInfiniteOp`.

### Rounding & truncation

```go
rt, err := t.Round(timex.FromStdDuration(time.Hour))
tt, err := t.Truncate(timex.FromStdDuration(time.Minute))
```

### Field accessors

All return the value plus an `InfFlag`:

```go
year, month, day, inf := t.Date()
hour, min, sec, inf   := t.Clock()
y, inf := t.Year()
m, inf := t.Month()
d, inf := t.Day()
h, inf := t.Hour()
mi, inf := t.Minute()
s, inf := t.Second()
ns, inf := t.Nanosecond()
isoYear, isoWeek, inf := t.ISOWeek()
wd, inf := t.Weekday() // 0 = Sunday … 6 = Saturday
t.IsDst()              // bool
```

### Unix timestamps

```go
sec, inf  := t.Unix()
ms, inf   := t.UnixMilli()
us, inf   := t.UnixMicro()
ns, inf   := t.UnixNano()

stdTime, inf := t.ToStdTime() // back to time.Time + InfFlag
```

### Time zones

```go
t = t.In(loc)
t = t.Local()
t = t.UTC()
name, offset := t.Zone()
loc := t.Location()
bounds, inf := t.ZoneBounds() // Interval covering the current zone offset
```

Infinite times are preserved across `In` / `Local` / `UTC`.

### Durations

```go
d := timex.FromStdDuration(2 * time.Hour)
d = timex.NewPosInfDuration()
d = timex.NewNegInfDuration()

std, inf := d.ToStdDuration()

h, inf  := d.Hours()
m, inf  := d.Minutes()
s, inf  := d.Seconds()
ms, inf := d.Milliseconds()
us, inf := d.Microseconds()
ns, inf := d.Nanoseconds()

rd, err := d.Round(timex.FromStdDuration(time.Minute))
td, err := d.Truncate(timex.FromStdDuration(time.Minute))
ad, err := d.Abs()

d.String() // "2h0m0s" or "positive infinite" / "negative infinite"
```

For infinite durations the unit accessors return `±math.MaxFloat64` (floats) or
`math.MinInt64`/`math.MaxInt64` (integers) together with the matching `InfFlag`.

### Intervals

```go
start := timex.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end   := timex.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
mid   := timex.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

// [start, end)
iv := timex.NewInterval(start, end, true, false)

iv.Contain(start) // true  (start included)
iv.Contain(end)   // false (end excluded)
iv.Contain(mid)   // true

iv.Before(t) // whole interval lies before t
iv.After(t)  // whole interval lies after t

s, sIncluded := iv.Start()
e, eIncluded := iv.End()
```

Combined with infinite times, intervals can model open-ended ranges such as
`[start, +∞)`.

### Serialization

`Time`, `Duration` and `Interval` implement `encoding.TextMarshaler` /
`encoding.TextUnmarshaler`. A single text form is therefore reused automatically
by `encoding/json`, `encoding/xml` and TOML — for both values and map keys — so
no format-specific code is needed. (YAML's `gopkg.in/yaml.v3` only consults its
own `Marshaler` interface, so add a small `MarshalYAML`/`UnmarshalYAML` adapter
there if you need it.)

```go
type Schedule struct {
	Name   string         `json:"name"`
	Window timex.Interval `json:"window"`
	Every  timex.Duration `json:"every"`
	Until  timex.Time     `json:"until"`
}

s := Schedule{
	Name:   "Q1",
	Window: timex.NewInterval(start, end, true, false),
	Every:  timex.FromStdDuration(24 * time.Hour),
	Until:  timex.NewPosInfTime(),
}

data, _ := json.Marshal(s)
// {
//   "name":   "Q1",
//   "window": "[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)",
//   "every":  "24h0m0s",
//   "until":  "positive infinite"
// }

var back Schedule
_ = json.Unmarshal(data, &back) // round-trips exactly
```

Text forms:

| Type       | Finite                                  | Infinite                                   |
|------------|-----------------------------------------|--------------------------------------------|
| `Time`     | RFC 3339, e.g. `2024-01-02T15:04:05Z`   | `positive infinite` / `negative infinite`  |
| `Duration` | Go duration, e.g. `24h0m0s`             | `positive infinite` / `negative infinite`  |
| `Interval` | `[start,end)` notation (see below)      | endpoints use the `Time` forms above       |

Interval brackets encode the bounds: `[`/`]` is inclusive, `(`/`)` is exclusive
— e.g. `[a,b]` closed, `(a,b)` open, `[a,b)` left-closed/right-open. Whitespace
around endpoints is tolerated when decoding. Open-ended ranges serialize
naturally, e.g. `[2024-01-01T00:00:00Z,positive infinite)`.

For binary protocols, the same types implement `encoding.BinaryMarshaler` /
`encoding.BinaryUnmarshaler`, which `encoding/gob` uses automatically — no type
registration required:

```go
var buf bytes.Buffer
_ = gob.NewEncoder(&buf).Encode(s) // s is the Schedule above
var back Schedule
_ = gob.NewDecoder(&buf).Decode(&back)
```

The binary form is compact (each value is prefixed by its `InfFlag` byte, so an
infinite `Time`/`Duration` is a single byte) and is meant for machine exchange,
not human inspection — use the text form for JSON and friends.

### Sleeping

```go
timex.Sleep(timex.FromStdDuration(time.Second)) // sleeps 1s
timex.Sleep(timex.NewNegInfDuration())          // returns immediately
timex.Sleep(timex.NewPosInfDuration())          // blocks forever
```

## Errors

| Error                   | When it occurs                                              |
|-------------------------|------------------------------------------------------------|
| `ErrInvalidInfiniteOp`  | Adding `±∞` to `∓∞`, or rounding/truncating with infinite bounds |
| `ErrInvalidInterval`    | Decoding an interval from text that is not valid `[start,end)` notation |
| `ErrInvalidBinary`      | Decoding a value from malformed or truncated binary data    |

## Testing

```bash
go test ./... -v
```

## License

Licensed under the [Apache License 2.0](LICENSE).
