package timex

import (
	"database/sql/driver"
	"encoding"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"go.gh.ink/toolbox/expr"
)

// timex types serialize through the encoding.TextMarshaler / TextUnmarshaler
// pair rather than format-specific interfaces. A single text representation is
// automatically reused by encoding/json, encoding/xml and TOML (BurntSushi/toml,
// pelletier/go-toml) — for both values and map keys — so one implementation
// covers those data-interchange languages. (YAML's gopkg.in/yaml.v3 uses its own
// Marshaler interface and does not consult TextMarshaler; wrap these if needed.)
var (
	_ encoding.TextMarshaler   = Time{}
	_ encoding.TextUnmarshaler = (*Time)(nil)
	_ encoding.TextMarshaler   = Duration{}
	_ encoding.TextUnmarshaler = (*Duration)(nil)
	_ encoding.TextMarshaler   = Interval{}
	_ encoding.TextUnmarshaler = (*Interval)(nil)
)

// MarshalText implements encoding.TextMarshaler. Finite times use the standard
// RFC 3339 text form (delegated to time.Time); infinite times use the
// PosInfTimeStr / NegInfTimeStr sentinels.
func (t Time) MarshalText() ([]byte, error) {
	switch {
	case t.inf > 0:
		return []byte(PosInfTimeStr), nil
	case t.inf < 0:
		return []byte(NegInfTimeStr), nil
	default:
		return t.std.MarshalText()
	}
}

// UnmarshalText implements encoding.TextUnmarshaler, the inverse of MarshalText.
func (t *Time) UnmarshalText(data []byte) error {
	switch string(data) {
	case PosInfTimeStr:
		*t = NewPosInfTime()
		return nil
	case NegInfTimeStr:
		*t = NewNegInfTime()
		return nil
	default:
		var std time.Time
		if err := std.UnmarshalText(data); err != nil {
			return err
		}
		*t = FromStdTime(std)
		return nil
	}
}

// MarshalText implements encoding.TextMarshaler. Finite durations use the
// standard "2h0m0s" text form (delegated to time.Duration); infinite durations
// use the PosInfTimeStr / NegInfTimeStr sentinels.
func (d Duration) MarshalText() ([]byte, error) {
	switch {
	case d.inf > 0:
		return []byte(PosInfTimeStr), nil
	case d.inf < 0:
		return []byte(NegInfTimeStr), nil
	default:
		return []byte(d.std.String()), nil
	}
}

// UnmarshalText implements encoding.TextUnmarshaler, the inverse of MarshalText.
func (d *Duration) UnmarshalText(data []byte) error {
	switch string(data) {
	case PosInfTimeStr:
		*d = NewPosInfDuration()
		return nil
	case NegInfTimeStr:
		*d = NewNegInfDuration()
		return nil
	default:
		std, err := time.ParseDuration(string(data))
		if err != nil {
			return err
		}
		*d = FromStdDuration(std)
		return nil
	}
}

// MarshalText implements encoding.TextMarshaler using mathematical interval
// notation: a "[" / "(" bound, the start, a comma, the end, and a "]" / ")"
// bound — e.g. "[2024-01-01T00:00:00Z,2024-04-01T00:00:00Z)". Brackets encode
// inclusivity ("[" / "]" included, "(" / ")" excluded) and each endpoint reuses
// Time's own text form, so infinite endpoints render as their sentinels.
func (i Interval) MarshalText() ([]byte, error) {
	start, err := i.start.MarshalText()
	if err != nil {
		return nil, err
	}
	end, err := i.end.MarshalText()
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString(expr.Ternary(i.startIncluded, "[", "("))
	b.Write(start)
	b.WriteByte(',')
	b.Write(end)
	b.WriteString(expr.Ternary(i.endIncluded, "]", ")"))
	return []byte(b.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, the inverse of MarshalText.
// Whitespace around the endpoints is tolerated. Neither RFC 3339 nor the
// infinity sentinels contain a comma, so the first comma always separates the
// two endpoints.
func (i *Interval) UnmarshalText(data []byte) error {
	s := strings.TrimSpace(string(data))
	if len(s) < 2 {
		return ErrInvalidInterval
	}

	var startIncluded, endIncluded bool
	switch s[0] {
	case '[':
		startIncluded = true
	case '(':
		startIncluded = false
	default:
		return ErrInvalidInterval
	}
	switch s[len(s)-1] {
	case ']':
		endIncluded = true
	case ')':
		endIncluded = false
	default:
		return ErrInvalidInterval
	}

	body := s[1 : len(s)-1]
	comma := strings.IndexByte(body, ',')
	if comma < 0 {
		return ErrInvalidInterval
	}

	var start, end Time
	if err := start.UnmarshalText([]byte(strings.TrimSpace(body[:comma]))); err != nil {
		return err
	}
	if err := end.UnmarshalText([]byte(strings.TrimSpace(body[comma+1:]))); err != nil {
		return err
	}

	*i = Interval{
		start:         start,
		startIncluded: startIncluded,
		end:           end,
		endIncluded:   endIncluded,
	}
	return nil
}

// timex types also implement encoding.BinaryMarshaler / BinaryUnmarshaler. This
// single pair is reused by encoding/gob (which prefers it over reflection when
// present) and any other codec built on those interfaces, giving a compact,
// exact, non-textual round-trip. Every layout begins with the InfFlag byte, so
// infinite values encode in a single byte.
var (
	_ encoding.BinaryMarshaler   = Time{}
	_ encoding.BinaryUnmarshaler = (*Time)(nil)
	_ encoding.BinaryMarshaler   = Duration{}
	_ encoding.BinaryUnmarshaler = (*Duration)(nil)
	_ encoding.BinaryMarshaler   = Interval{}
	_ encoding.BinaryUnmarshaler = (*Interval)(nil)
)

// MarshalBinary implements encoding.BinaryMarshaler. The first byte holds the
// InfFlag; a finite time appends time.Time's own binary form, while an infinite
// time needs no further bytes.
func (t Time) MarshalBinary() ([]byte, error) {
	if t.inf != FiniteTime {
		return []byte{byte(t.inf)}, nil
	}
	std, err := t.std.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(t.inf)}, std...), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, the inverse of MarshalBinary.
func (t *Time) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidBinary
	}
	switch InfFlag(int8(data[0])) {
	case PosInfTime:
		*t = NewPosInfTime()
		return nil
	case NegInfTime:
		*t = NewNegInfTime()
		return nil
	case FiniteTime:
		var std time.Time
		if err := std.UnmarshalBinary(data[1:]); err != nil {
			return err
		}
		*t = FromStdTime(std)
		return nil
	default:
		return ErrInvalidBinary
	}
}

// MarshalBinary implements encoding.BinaryMarshaler. The first byte holds the
// InfFlag; a finite duration appends its int64 nanoseconds (big-endian), while
// an infinite duration needs no further bytes.
func (d Duration) MarshalBinary() ([]byte, error) {
	if d.inf != FiniteTime {
		return []byte{byte(d.inf)}, nil
	}
	buf := make([]byte, 1+8)
	buf[0] = byte(d.inf)
	binary.BigEndian.PutUint64(buf[1:], uint64(d.std))
	return buf, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, the inverse of MarshalBinary.
func (d *Duration) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidBinary
	}
	switch InfFlag(int8(data[0])) {
	case PosInfTime:
		*d = NewPosInfDuration()
		return nil
	case NegInfTime:
		*d = NewNegInfDuration()
		return nil
	case FiniteTime:
		if len(data) != 1+8 {
			return ErrInvalidBinary
		}
		*d = FromStdDuration(time.Duration(binary.BigEndian.Uint64(data[1:])))
		return nil
	default:
		return ErrInvalidBinary
	}
}

// MarshalBinary implements encoding.BinaryMarshaler. Layout: a flags byte (bit 0
// startIncluded, bit 1 endIncluded), the uvarint length of the start endpoint's
// binary form, that start form, then the end endpoint's binary form (the
// remaining bytes). Both endpoints reuse Time's own binary encoding.
func (i Interval) MarshalBinary() ([]byte, error) {
	start, err := i.start.MarshalBinary()
	if err != nil {
		return nil, err
	}
	end, err := i.end.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var flags byte
	if i.startIncluded {
		flags |= 1
	}
	if i.endIncluded {
		flags |= 2
	}

	buf := make([]byte, 0, 1+binary.MaxVarintLen64+len(start)+len(end))
	buf = append(buf, flags)
	buf = binary.AppendUvarint(buf, uint64(len(start)))
	buf = append(buf, start...)
	buf = append(buf, end...)
	return buf, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler, the inverse of MarshalBinary.
func (i *Interval) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return ErrInvalidBinary
	}
	flags := data[0]

	rest := data[1:]
	n, read := binary.Uvarint(rest)
	if read <= 0 {
		return ErrInvalidBinary
	}
	rest = rest[read:]
	if uint64(len(rest)) < n {
		return ErrInvalidBinary
	}
	startBin, endBin := rest[:n], rest[n:]

	var start, end Time
	if err := start.UnmarshalBinary(startBin); err != nil {
		return err
	}
	if err := end.UnmarshalBinary(endBin); err != nil {
		return err
	}

	*i = Interval{
		start:         start,
		startIncluded: flags&1 != 0,
		end:           end,
		endIncluded:   flags&2 != 0,
	}
	return nil
}

// timex types implement database/sql/driver.Valuer and database/sql.Scanner,
// round-tripping through their text form, so they work as column values with
// database/sql and ORMs built on it (xorm, gorm, sqlx, ...) without an
// app-level adapter. A text/varchar column preserves everything — infinite
// bounds and interval inclusivity included. Scan accepts the string or []byte a
// driver yields for a text column and maps SQL NULL to the zero value.
//
// The Scanner side is asserted against an inline interface so importing timex
// does not pull all of database/sql into binaries that never touch a database;
// only the lightweight database/sql/driver package is required (for Value).
var (
	_ driver.Valuer                = Time{}
	_ driver.Valuer                = Duration{}
	_ driver.Valuer                = Interval{}
	_ interface{ Scan(any) error } = (*Time)(nil)
	_ interface{ Scan(any) error } = (*Duration)(nil)
	_ interface{ Scan(any) error } = (*Interval)(nil)
)

// textValue is the shared driver.Valuer body: marshal to the canonical text form
// and hand the driver a string, i.e. a text/varchar column value.
func textValue(m encoding.TextMarshaler) (driver.Value, error) {
	b, err := m.MarshalText()
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// scanText is the shared sql.Scanner body for a non-NULL src: it forwards the
// string / []byte a driver returns for a text column to u.UnmarshalText. NULL is
// handled by each Scan method (reset to the zero value) before calling this.
func scanText(u encoding.TextUnmarshaler, src any) error {
	switch v := src.(type) {
	case string:
		return u.UnmarshalText([]byte(v))
	case []byte:
		return u.UnmarshalText(v)
	default:
		return fmt.Errorf("timex: cannot scan %T into %T", src, u)
	}
}

// Value implements database/sql/driver.Valuer.
func (t Time) Value() (driver.Value, error) { return textValue(t) }

// Scan implements database/sql.Scanner. A NULL column yields the zero Time.
func (t *Time) Scan(src any) error {
	if src == nil {
		*t = Time{}
		return nil
	}
	return scanText(t, src)
}

// Value implements database/sql/driver.Valuer.
func (d Duration) Value() (driver.Value, error) { return textValue(d) }

// Scan implements database/sql.Scanner. A NULL column yields the zero Duration.
func (d *Duration) Scan(src any) error {
	if src == nil {
		*d = Duration{}
		return nil
	}
	return scanText(d, src)
}

// Value implements database/sql/driver.Valuer.
func (i Interval) Value() (driver.Value, error) { return textValue(i) }

// Scan implements database/sql.Scanner. A NULL column yields the zero Interval.
func (i *Interval) Scan(src any) error {
	if src == nil {
		*i = Interval{}
		return nil
	}
	return scanText(i, src)
}
