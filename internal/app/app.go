package app

import (
	"basket/internal/config"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type App struct {
	LogsDB *gorm.DB

	Config    config.Config
	DB        *gorm.DB
	Cache     Cache
	HTTP      *http.Client
	auth      *Authenticator
	Log       *slog.Logger
	Now       func() time.Time
	location  *time.Location
	jobs      chan func()
	wg        sync.WaitGroup
	cron      *cron.Cron
	cancel    context.CancelFunc
	ctx       context.Context
	contracts sync.Map
}

/*
func New(ctx context.Context, c config.Config, db *gorm.DB, cache Cache) (*App, error) {
	auth, e := newAuth(ctx, c)
	if e != nil {
		return nil, e
	}
	loc, e := time.LoadLocation(c.Business.Timezone)
	if e != nil {
		return nil, e
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.Upstream.TLSInsecureSkipVerify}
	if c.Upstream.CAFile != "" {
		pem, e := os.ReadFile(c.Upstream.CAFile)
		if e != nil {
			return nil, e
		}
		pool, e := x509.SystemCertPool()
		if e != nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("invalid upstream CA file")
		}
		transport.TLSClientConfig.RootCAs = pool
	}
	jobctx, cancel := context.WithCancel(context.Background())
	a := &App{Config: c, DB: db, Cache: cache, HTTP: &http.Client{Timeout: c.Upstream.Timeout, Transport: transport}, auth: auth, Log: slog.Default(), Now: time.Now, location: loc, jobs: make(chan func(), c.Business.AdminQueue), ctx: jobctx, cancel: cancel}
	for i := 0; i < c.Business.AdminWorkers; i++ {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			for job := range a.jobs {
				func() {
					defer func() {
						if v := recover(); v != nil {
							a.Log.Error("background job panic", "error", v)
						}
					}()
					job()
				}()
			}
		}()
	}
	if c.Logging.Database {
		a.LogsDB = db
		if c.Logging.DatabaseConfig != nil {
			lc := c
			lc.Database = *c.Logging.DatabaseConfig
			lc.Database.AutoMigrate = false
			var e error
			a.LogsDB, e = OpenDatabase(lc)
			if e != nil {
				close(a.jobs)
				a.wg.Wait()
				cancel()
				return nil, e
			}
		}
	}
	return a, nil
}
*/

// added for sonarqube
func setupAppTransport(c config.Config) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 90 * time.Second
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: c.Upstream.TLSInsecureSkipVerify}
	if c.Upstream.CAFile != "" {
		pem, e := os.ReadFile(c.Upstream.CAFile)
		if e != nil {
			return nil, e
		}
		pool, e := x509.SystemCertPool()
		if e != nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("invalid upstream CA file")
		}
		transport.TLSClientConfig.RootCAs = pool
	}
	return transport, nil
}

// added for sonarqube
func (a *App) startAdminWorkers(workers int) {
	for i := 0; i < workers; i++ {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			for job := range a.jobs {
				func() {
					defer func() {
						if v := recover(); v != nil {
							a.Log.Error("background job panic", "error", v)
						}
					}()
					job()
				}()
			}
		}()
	}
}

// added for sonarqube
func (a *App) setupLogsDB(c config.Config, cancel context.CancelFunc) error {
	if !c.Logging.Database {
		return nil
	}
	a.LogsDB = a.DB
	if c.Logging.DatabaseConfig != nil {
		lc := c
		lc.Database = *c.Logging.DatabaseConfig
		lc.Database.AutoMigrate = false
		db, e := OpenDatabase(lc)
		if e != nil {
			close(a.jobs)
			a.wg.Wait()
			cancel()
			return e
		}
		a.LogsDB = db
	}
	return nil
}

// added for sonarqube
func New(ctx context.Context, c config.Config, db *gorm.DB, cache Cache) (*App, error) {
	auth, e := newAuth(ctx, c)
	if e != nil {
		return nil, e
	}
	loc, e := time.LoadLocation(c.Business.Timezone)
	if e != nil {
		return nil, e
	}
	transport, e := setupAppTransport(c)
	if e != nil {
		return nil, e
	}
	jobctx, cancel := context.WithCancel(context.Background())
	a := &App{Config: c, DB: db, Cache: cache, HTTP: &http.Client{Timeout: c.Upstream.Timeout, Transport: transport}, auth: auth, Log: slog.Default(), Now: time.Now, location: loc, jobs: make(chan func(), c.Business.AdminQueue), ctx: jobctx, cancel: cancel}
	a.startAdminWorkers(c.Business.AdminWorkers)
	if e := a.setupLogsDB(c, cancel); e != nil {
		return nil, e
	}
	return a, nil
}
func (a *App) Start() error {
	if a.Config.Modules.Cache && a.Config.Scheduler.StartupCleanup {
		a.deleteExpiredScrips(a.ctx)
	}
	if a.Config.Modules.Scheduler {
		a.cron = cron.New(cron.WithSeconds(), cron.WithLocation(a.location), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
		_, e := a.cron.AddFunc(a.Config.Scheduler.Cron, func() {
			if a.Config.Modules.Cache && a.Config.Scheduler.DeleteScrips {
				a.deleteExpiredScrips(a.ctx)
			}
			if a.Config.Modules.Admin && a.Config.Scheduler.DeleteBaskets {
				a.deleteExpiredBaskets(a.ctx)
			}
		})
		if e != nil {
			return e
		}
		a.cron.Start()
	}
	return nil
}
func (a *App) Close(ctx context.Context) error {
	if a.cron != nil {
		select {
		case <-a.cron.Stop().Done():
		case <-ctx.Done():
			a.cancel()
			return ctx.Err()
		}
	}
	close(a.jobs)
	done := make(chan struct{})
	go func() { a.wg.Wait(); close(done) }()
	select {
	case <-done:
		a.cancel()
		if a.LogsDB != nil && a.LogsDB != a.DB {
			if sql, e := a.LogsDB.DB(); e == nil {
				sql.Close()
			}
		}
		return a.Cache.Close(ctx)
	case <-ctx.Done():
		a.cancel()
		return ctx.Err()
	}
}

type route struct {
	Method, Path, Module, Action string
	List                         bool
	Public                       bool
}

var routes = []route{
	{"POST", "/basketorder/create", "basket", "create", false, false}, {"GET", "/basketorder/get", "basket", "get", false, false}, {"POST", "/basketorder/rename", "basket", "rename", false, false}, {"DELETE", "/basketorder/delete/{basketId}", "basket", "delete", false, false},
	{"POST", "/basketorder/add/scrips", "basket", "addScrip", false, false}, {"POST", "/basketorder/delete/scrips", "basket", "deleteScrip", false, false}, {"POST", "/basketorder/update/scrips", "basket", "updateScrip", false, false}, {"POST", "/basketorder/update/scrips/list", "basket", "updateScripList", false, false}, {"GET", "/basketorder/get/scrips/{basketId}", "basket", "scrips", false, false}, {"POST", "/basketorder/execute", "basket", "execute", true, false}, {"GET", "/basketorder/reset/{basketId}", "basket", "reset", false, false},
	{"POST", "/basketorder/spanmargin", "margin", "span", false, false}, {"POST", "/basketorder/nest/spanmargin", "margin", "nestSpan", false, false},
	{"POST", "/basketorderapi/adminCreate", "admin", "adminCreate", false, true}, {"POST", "/basketorderapi/adminCreate/temp", "admin", "adminTemp", false, true}, {"DELETE", "/basketorderapi/deleteExpiredBasket", "admin", "deleteExpired", false, false},
	{"POST", "/research/getResearchWithBasket", "research", "researchBasket", false, false}, {"POST", "/research/getResearchCall", "research", "research", false, false}, {"GET", "/research/getUniqStatus", "research", "status", false, false}, {"GET", "/research/getall/sector/data", "research", "sectors", false, false}, {"GET", "/research/get/sector/data/{id}", "research", "sector", false, false}, {"POST", "/research/getall/rc/report", "research", "reports", false, false},
	{"GET", "/thematic/basket/getall", "thematic", "thematicAll", false, false}, {"GET", "/thematic/basket/get/{id}", "thematic", "thematicDetails", false, false}, {"POST", "/thematic/basket/invest", "thematic", "invest", true, false}, {"POST", "/thematic/basket/v1/invest", "thematic", "investV1", true, false}, {"POST", "/thematic/basket/review", "thematic", "review", false, false}, {"POST", "/thematic/basket/rebalance/details", "thematic", "rebalance", false, false}, {"POST", "/thematic/basket/report", "thematic", "view", false, false}, {"GET", "/thematic/basket/get/category/header", "thematic", "category", false, false}, {"GET", "/thematic/basket/holdings", "thematic", "legacyHoldings", false, false},
	{"GET", "/thematic/holdings/get", "holdings", "holdings", false, false}, {"GET", "/thematic/holdings/get/V1", "holdings", "holdingsV1", false, false}, {"POST", "/cache/delete/expiry", "cache", "expiry", false, false}, {"GET", "/token", "token", "token", false, false}, {"GET", "/token/logout", "token", "logout", false, false},
}

func (a *App) enabled(m string) bool {
	c := a.Config.Modules
	return map[string]bool{"basket": c.Basket, "admin": c.Admin, "research": c.Research, "thematic": c.Thematic, "holdings": c.Holdings, "margin": c.Margin, "cache": c.Cache, "token": c.Token}[m]
}
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	for _, rt := range routes {
		rt := rt
		if !a.enabled(rt.Module) {
			continue
		}
		mux.HandleFunc(rt.Method+" "+strings.TrimRight(a.Config.Server.BasePath, "/")+rt.Path, func(w http.ResponseWriter, r *http.Request) { a.serve(w, r, rt) })
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			for _, allowed := range a.Config.Server.AllowedOrigins {
				if origin == allowed || allowed == "*" {
					w.Header().Set("Access-Control-Allow-Origin", allowed)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
					break
				}
			}
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func (a *App) serve(w http.ResponseWriter, r *http.Request, rt route) {
	start := a.Now()
	capture := &captureWriter{ResponseWriter: w, limit: a.Config.Logging.MaxBodyBytes}
	w = capture
	var requestBody string
	defer func() { a.accessLog(r, rt, start, requestBody, capture) }()
	defer func() {
		if v := recover(); v != nil {
			a.Log.Error("request panic", "path", r.URL.Path, "error", v)
			//added for sonarqube
			// jsonWrite(w, 500, failed("Failed"))
			jsonWrite(w, 500, failed(msgFailed))
		}
		if a.Config.Logging.Access {
			a.Log.Info("request", "method", r.Method, "path", r.URL.Path, "elapsed", a.Now().Sub(start))
		}
	}()
	r.Body = http.MaxBytesReader(w, r.Body, a.Config.Server.MaxBodyBytes)
	if a.Config.Logging.Database || a.Config.Logging.ClickHouseURL != "" {
		b, e := io.ReadAll(r.Body)
		if e != nil {
			//added for sonarqube
			// jsonWrite(w, 400, failed("Invalid Parameter"))
			jsonWrite(w, 400, failed(invalidParameter))
			return
		}
		requestBody = redact(string(b))
		r.Body = io.NopCloser(bytes.NewReader(b))
	}
	authedReq, status, errResp, ok := a.authenticate(r, rt)
	if !ok {
		jsonWrite(w, status, errResp)
		return
	}
	r = authedReq
	if handleSpecialActions(w, r, rt) {
		return
	}
	out, e := a.dispatchModule(r, rt)
	if e != nil {
		a.Log.Error("request failed", "path", r.URL.Path, "error", e)
		out = failed(msgFailed)
	}
	writeResponse(w, out, rt)
}

// added for sonarqube
func (a *App) authenticate(r *http.Request, rt route) (*http.Request, int, any, bool) {
	if rt.Public {
		return r, 0, nil, true
	}
	for _, p := range a.Config.Auth.PublicPaths {
		if p == rt.Path {
			return r, 0, nil, true
		}
	}
	id, e := a.auth.verify(r)
	if e != nil {
		if rt.List {
			return r, 401, []any{}, false
		}
		return r, 401, Response{}, false
	}
	return r.WithContext(context.WithValue(r.Context(), identityKey{}, id)), 0, nil, true
}

// added for sonarqube
func handleSpecialActions(w http.ResponseWriter, r *http.Request, rt route) bool {
	switch rt.Action {
	case "token":
		id := identity(r)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body><ul><li>username: %s</li><li>scopes: %s</li><li>refresh_token: false</li></ul></body></html>", html.EscapeString(id.UserID), html.EscapeString(id.Scope))
		return true
	case "logout":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "You are logged out")
		return true
	case "category":
		w.WriteHeader(204)
		return true
	default:
		return false
	}
}

// added for sonarqube
func writeResponse(w http.ResponseWriter, out any, rt route) {
	status := 200
	if h, ok := out.(httpResult); ok {
		status = h.Status
		out = h.Body
	}
	if rt.List {
		if _, ok := out.(Response); ok {
			out = []any{out}
		}
	}
	jsonWrite(w, status, out)
}

// added for sonarqube
func (a *App) dispatchModule(r *http.Request, rt route) (any, error) {
	switch rt.Module {
	case "basket":
		return a.basket(r, rt.Action)
	case "admin":
		return a.admin(r, rt.Action)
	case "margin":
		return a.margin(r, rt.Action)
	case "research":
		return a.research(r, rt.Action)
	case "thematic":
		return a.thematic(r, rt.Action)
	case "holdings":
		return a.holdings(r, rt.Action == "holdingsV1")
	case "cache":
		return a.deleteExpiredScrips(r.Context()), nil
	default:
		return nil, nil
	}
}

type httpResult struct {
	Status int
	Body   any
}
