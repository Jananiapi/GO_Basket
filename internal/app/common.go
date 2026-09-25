package app

import (
	"basket/internal/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type M = map[string]any
type Response struct {
	Status  any `json:"status"`
	Message any `json:"message"`
	Result  any `json:"result"`
}

func success(v any) Response {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) {
		v = []any{v}
	}
	return Response{"Ok", "Success", v}
}
func message(s string) Response { return Response{"Ok", s, nil} }
func failed(s string) Response  { return Response{"Not ok", s, []any{}} }
func ptr[T any](v T) *T         { return &v }
func val[T any](v *T) (z T) {
	if v != nil {
		return *v
	}
	return
}
func str(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case *string:
		return val(x)
	case []byte:
		return string(x)
	case time.Time:
		return x.Format("2006-01-02 15:04:05")
	case *model.Timestamp:
		if x != nil {
			return x.Time().Format("2006-01-02 15:04:05")
		}
		return ""
	}
	return fmt.Sprint(v)
}
func integer(v any) int    { f, _ := strconv.ParseFloat(str(v), 64); return int(f) }
func number(v any) float64 { f, _ := strconv.ParseFloat(str(v), 64); return f }
func dec(v any) decimal.Decimal {
	d, e := decimal.NewFromString(str(v))
	if e != nil {
		return decimal.Zero
	}
	return d
}
func nonempty(v any) bool {
	s := strings.TrimSpace(str(v))
	return s != ""
}
func jsonString(v any) string { b, _ := json.Marshal(v); return string(b) }
func jsonWrite(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decode retains Jackson scalar coercions (for example ltp: "43.04") and
// case-sensitive DTO field names, while ignoring unknown request properties.
func decode(r *http.Request, v any) error {
	b, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}
	if len(b) == 0 {
		return io.EOF
	}
	// Fast path: try standard direct decode first
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e := d.Decode(v); e == nil {
		var extra any
		if extraErr := d.Decode(&extra); extraErr == io.EOF {
			return nil
		}
		return errors.New("multiple JSON values")
	}

	// Fallback path for Jackson scalar coercions (e.g. string numbers to float)
	d2 := json.NewDecoder(bytes.NewReader(b))
	d2.UseNumber()
	var raw any
	if e := d2.Decode(&raw); e != nil {
		return e
	}
	var extra any
	if e := d2.Decode(&extra); e != io.EOF {
		return errors.New("multiple JSON values")
	}
	raw = coerce(raw, reflect.TypeOf(v).Elem())
	cb, e := json.Marshal(raw)
	if e != nil {
		return e
	}
	return json.Unmarshal(cb, v)
}
/*
func coerce(v any, t reflect.Type) any {
	if v == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(model.Timestamp{}) {
		if v == "" {
			return nil
		}
		return v
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return v
		}
		out := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			key := strings.Split(f.Tag.Get("json"), ",")[0]
			if key == "-" {
				continue
			}
			if key == "" {
				key = f.Name
			}
			if x, ok := m[key]; ok {
				out[key] = coerce(x, f.Type)
			}
		}
		return out
	case reflect.Slice:
		if xs, ok := v.([]any); ok {
			for i := range xs {
				xs[i] = coerce(xs[i], t.Elem())
			}
			return xs
		}
	case reflect.String:
		if _, ok := v.(json.Number); ok {
			return str(v)
		}
		if b, ok := v.(bool); ok {
			return strconv.FormatBool(b)
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Float32, reflect.Float64:
		if s, ok := v.(string); ok {
			if s == "" {
				return nil
			}
			if _, e := strconv.ParseFloat(s, 64); e == nil {
				return json.Number(s)
			}
		}
	}
	return v
}
*/

// added for sonarqube
func coerceStruct(m map[string]any, t reflect.Type) any {
	out := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		key := strings.Split(f.Tag.Get("json"), ",")[0]
		if key == "-" {
			continue
		}
		if key == "" {
			key = f.Name
		}
		if x, ok := m[key]; ok {
			out[key] = coerce(x, f.Type)
		}
	}
	return out
}

// added for sonarqube
func coerceSlice(xs []any, elemType reflect.Type) any {
	for i := range xs {
		xs[i] = coerce(xs[i], elemType)
	}
	return xs
}

// added for sonarqube
func coerceScalar(v any, k reflect.Kind) any {
	switch k {
	case reflect.String:
		if _, ok := v.(json.Number); ok {
			return str(v)
		}
		if b, ok := v.(bool); ok {
			return strconv.FormatBool(b)
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Float32, reflect.Float64:
		if s, ok := v.(string); ok {
			if s == "" {
				return nil
			}
			if _, e := strconv.ParseFloat(s, 64); e == nil {
				return json.Number(s)
			}
		}
	}
	return v
}

// added for sonarqube
func coerce(v any, t reflect.Type) any {
	if v == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(model.Timestamp{}) {
		if v == "" {
			return nil
		}
		return v
	}
	switch t.Kind() {
	case reflect.Struct:
		if m, ok := v.(map[string]any); ok {
			return coerceStruct(m, t)
		}
		return v
	case reflect.Slice:
		if xs, ok := v.([]any); ok {
			return coerceSlice(xs, t.Elem())
		}
		return v
	default:
		return coerceScalar(v, t.Kind())
	}
}

func copyJSON(src, dst any) error {
	b, e := json.Marshal(src)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, dst)
}
func rows(q *gorm.DB) ([]M, error) {
	out := []M{}
	e := q.Find(&out).Error
	for _, r := range out {
		for k, v := range r {
			if b, ok := v.([]byte); ok {
				r[k] = string(b)
			}
		}
	}
	return out, e
}
func first(q *gorm.DB) (M, error) {
	r, e := rows(q.Limit(1))
	if e != nil {
		return nil, e
	}
	if len(r) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r[0], nil
}
func project(r M, fields map[string]string) M {
	m := M{}
	for j, c := range fields {
		m[j] = r[c]
	}
	return m
}
func dateOnly(v any) any {
	if v == nil {
		return nil
	}
	s := str(v)
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}
func epoch(v any) any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case time.Time:
		return t.UnixMilli()
	case *model.Timestamp:
		if t == nil {
			return nil
		}
		return t.Time().UnixMilli()
	}
	var t model.Timestamp
	if t.Scan(v) == nil {
		return t.Time().UnixMilli()
	}
	return v
}
