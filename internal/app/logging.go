package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type captureWriter struct {
	http.ResponseWriter
	status int
	body   []byte
	limit  int
}

func (w *captureWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
		w.ResponseWriter.WriteHeader(code)
	}
}
func (w *captureWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	remaining := w.limit - len(w.body)
	if remaining > 0 {
		w.body = append(w.body, b[:min(remaining, len(b))]...)
	}
	return w.ResponseWriter.Write(b)
}
/*
func redact(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return s
	}
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			for k, v := range t {
				switch strings.ToLower(k) {
				case "apikey", "api_key", "apisecret", "api_secret", "authorization", "password", "jkey", "stringpkey4", "publickey4", "access_token", "refresh_token":
					t[k] = "[redacted]"
				default:
					walk(v)
				}
			}
		case []any:
			for _, v := range t {
				walk(v)
			}
		}
	}
	walk(v)
	return jsonString(v)
}
*/

// added for sonarqube
func isRedactField(k string) bool {
	switch strings.ToLower(k) {
	case "apikey", "api_key", "apisecret", "api_secret", "authorization", "password", "jkey", "stringpkey4", "publickey4", "access_token", "refresh_token":
		return true
	default:
		return false
	}
}

// added for sonarqube
func redactMapValue(m map[string]any, walk func(any)) {
	for k, val := range m {
		if isRedactField(k) {
			m[k] = "[redacted]"
		} else {
			walk(val)
		}
	}
}

// added for sonarqube
func redact(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return s
	}
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			redactMapValue(t, walk)
		case []any:
			for _, item := range t {
				walk(item)
			}
		}
	}
	walk(v)
	return jsonString(v)
}

// added for sonarqube
const (
	logKeyReqBody = "req_body"
	logKeyResBody = "res_body"
)

func (a *App) accessLog(r *http.Request, rt route, start time.Time, body string, w *captureWriter) {
	c := a.Config.Logging
	if !c.Database && c.ClickHouseURL == "" {
		return
	}
	id := identity(r)
	record := M{"user_id": id.UserID, "ucc": id.UCC, "req_id": r.Header.Get("X-Request-ID"), "source": "", "vendor": "", "in_time": start, "out_time": a.Now(), "lag_time": a.Now().Sub(start).Milliseconds(), "module": "Basket", "method": r.Method, logKeyReqBody: body, logKeyResBody: redact(string(w.body)), "device_ip": r.RemoteAddr, "user_agent": r.UserAgent(), "domain": r.Host, "content_type": r.Header.Get("Content-Type"), "session": "[redacted]", "uri": r.URL.Path}
	table := start.In(a.location).Format(c.AccessTablePattern)
	status := w.status

	job := func() {
		parentCtx := a.ctx
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		logCtx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
		defer cancel()
		a.writeLog(logCtx, table, c.ClickHouseAccessTable, record, status)
	}

	if a.jobs != nil {
		select {
		case a.jobs <- job:
			return
		default:
		}
	}
	a.writeLog(context.Background(), table, c.ClickHouseAccessTable, record, status)
}
func (a *App) restLog(ctx context.Context, u, body, response string, start time.Time) {
	c := a.Config.Logging
	if !c.Database && c.ClickHouseURL == "" {
		return
	}
	if parsed, e := url.Parse(u); e == nil {
		parsed.RawQuery = ""
		u = parsed.String()
	}
	record := M{"user_id": "", "url": u, "in_time": start, "out_time": a.Now(), "total_time": fmt.Sprint(a.Now().Sub(start).Milliseconds()), "module": "Basket", "method": "POST", logKeyReqBody: redact(body), logKeyResBody: redact(response)}
	table := start.In(a.location).Format(c.RestTablePattern)

	job := func() {
		parentCtx := a.ctx
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		logCtx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
		defer cancel()
		a.writeLog(logCtx, table, c.ClickHouseRestTable, record, 0)
	}

	if a.jobs != nil {
		select {
		case a.jobs <- job:
			return
		default:
		}
	}
	a.writeLog(context.Background(), table, c.ClickHouseRestTable, record, 0)
}

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

/*
func (a *App) writeLog(ctx context.Context, table, chTable string, record M, status int) {
	//added for sonarqube
	// for _, key := range []string{"req_body", "res_body"} {
	for _, key := range []string{logKeyReqBody, logKeyResBody} {
		s := str(record[key])
		if len(s) > a.Config.Logging.MaxBodyBytes {
			s = s[:a.Config.Logging.MaxBodyBytes]
		}
		record[key] = s
	}
	if a.LogsDB != nil {
		if !identifier.MatchString(table) {
			a.Log.Error("invalid log table name")
			return
		}
		if e := a.LogsDB.WithContext(ctx).Table(table).Create(record).Error; e != nil {
			a.Log.Error("database audit log", "error", e)
		}
	}
	c := a.Config.Logging
	if c.ClickHouseURL == "" {
		return
	}
	if !identifier.MatchString(c.ClickHouseDatabase) || !identifier.MatchString(chTable) {
		a.Log.Error("invalid ClickHouse identifier")
		return
	}
	u, e := url.Parse(c.ClickHouseURL)
	if e != nil {
		a.Log.Error("ClickHouse URL", "error", e)
		return
	}
	q := u.Query()
	q.Set("query", "INSERT INTO "+c.ClickHouseDatabase+"."+chTable+" FORMAT JSONEachRow")
	u.RawQuery = q.Encode()
	ch := M{}
	for k, v := range record {
		if t, ok := v.(time.Time); ok {
			v = t.UTC().Format("2006-01-02 15:04:05.000")
		}
		ch[k] = v
	}
	if status > 0 {
		ch["status_code"] = status
	}
	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(jsonString(ch)+"\n"))
	if e != nil {
		return
	}
	req.SetBasicAuth(c.ClickHouseUser, c.ClickHousePassword)
	resp, e := a.HTTP.Do(req)
	if e != nil {
		a.Log.Error("ClickHouse audit log", "error", e)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		a.Log.Error("ClickHouse audit log", "status", resp.StatusCode)
	}
}
*/

// added for sonarqube
func (a *App) writeClickHouseLog(ctx context.Context, chTable string, record M, status int) {
	c := a.Config.Logging
	if c.ClickHouseURL == "" {
		return
	}
	if !identifier.MatchString(c.ClickHouseDatabase) || !identifier.MatchString(chTable) {
		a.Log.Error("invalid ClickHouse identifier")
		return
	}
	u, e := url.Parse(c.ClickHouseURL)
	if e != nil {
		a.Log.Error("ClickHouse URL", "error", e)
		return
	}
	q := u.Query()
	q.Set("query", "INSERT INTO "+c.ClickHouseDatabase+"."+chTable+" FORMAT JSONEachRow")
	u.RawQuery = q.Encode()
	ch := M{}
	for k, v := range record {
		if t, ok := v.(time.Time); ok {
			v = t.UTC().Format("2006-01-02 15:04:05.000")
		}
		ch[k] = v
	}
	if status > 0 {
		ch["status_code"] = status
	}
	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(jsonString(ch)+"\n"))
	if e != nil {
		return
	}
	req.SetBasicAuth(c.ClickHouseUser, c.ClickHousePassword)
	resp, e := a.HTTP.Do(req)
	if e != nil {
		a.Log.Error("ClickHouse audit log", "error", e)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		a.Log.Error("ClickHouse audit log", "status", resp.StatusCode)
	}
}

// added for sonarqube
func (a *App) writeLog(ctx context.Context, table, chTable string, record M, status int) {
	for _, key := range []string{logKeyReqBody, logKeyResBody} {
		s := str(record[key])
		if len(s) > a.Config.Logging.MaxBodyBytes {
			s = s[:a.Config.Logging.MaxBodyBytes]
		}
		record[key] = s
	}
	if a.LogsDB != nil {
		if !identifier.MatchString(table) {
			a.Log.Error("invalid log table name")
			return
		}
		if e := a.LogsDB.WithContext(ctx).Table(table).Create(record).Error; e != nil {
			a.Log.Error("database audit log", "error", e)
		}
	}
	a.writeClickHouseLog(ctx, chTable, record, status)
}
