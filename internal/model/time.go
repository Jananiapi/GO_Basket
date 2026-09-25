package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Timestamp preserves Jackson's epoch-millisecond Date JSON representation.
type Timestamp time.Time

func Now() *Timestamp               { t := Timestamp(time.Now()); return &t }
func (t Timestamp) Time() time.Time { return time.Time(t) }
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).UnixMilli(), 10)), nil
}
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		*t = Timestamp(time.Time{})
		return nil
	}
	var n int64
	if json.Unmarshal(b, &n) == nil {
		*t = Timestamp(time.UnixMilli(n))
		return nil
	}
	var s string
	if e := json.Unmarshal(b, &s); e != nil {
		return e
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999", "2006-01-02"} {
		if v, e := time.Parse(layout, s); e == nil {
			*t = Timestamp(v)
			return nil
		}
	}
	return fmt.Errorf("invalid date %q", s)
}
func (t Timestamp) Value() (driver.Value, error) { return time.Time(t), nil }
func (t *Timestamp) Scan(v any) error {
	switch x := v.(type) {
	case nil:
		*t = Timestamp(time.Time{})
		return nil
	case time.Time:
		*t = Timestamp(x)
		return nil
	case []byte:
		return t.Scan(string(x))
	case string:
		b, _ := json.Marshal(x)
		return t.UnmarshalJSON(b)
	}
	return fmt.Errorf("invalid timestamp type %T", v)
}
func (Timestamp) GormDataType() string { return "time" }
