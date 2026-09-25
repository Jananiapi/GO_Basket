package main

import (
	"basket/internal/app"
	"basket/internal/config"
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	path := flag.String("config", "config.yaml", "YAML configuration path")
	check := flag.Bool("check-config", false, "validate configuration without connecting to services")
	migrate := flag.Bool("migrate", false, "create/update database schema and exit")
	flag.Parse()
	c, e := config.Load(*path)
	if e != nil {
		slog.Error("configuration error", "error", e)
		os.Exit(1)
	}
	if *check {
		slog.Info("configuration is valid")
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, e := app.OpenDatabase(c)
	if e != nil {
		slog.Error("database connection failed", "error", e)
		os.Exit(1)
	}
	sql, _ := db.DB()
	defer sql.Close()
	if *migrate {
		if e = app.Migrate(db); e != nil {
			slog.Error("migration failed", "error", e)
			os.Exit(1)
		}
		return
	}
	cache, e := app.NewCache(ctx, c)
	if e != nil {
		slog.Error("cache initialization failed", "error", e)
		os.Exit(1)
	}
	a, e := app.New(ctx, c, db, cache)
	if e != nil {
		slog.Error("application initialization failed", "error", e)
		os.Exit(1)
	}
	if e = a.Start(); e != nil {
		slog.Error("scheduler initialization failed", "error", e)
		os.Exit(1)
	}
	srv := &http.Server{Addr: c.Server.Address, Handler: a.Handler(), ReadHeaderTimeout: c.Server.ReadTimeout, ReadTimeout: c.Server.ReadTimeout, WriteTimeout: c.Server.WriteTimeout}
	errch := make(chan error, 1)
	go func() {
		slog.Info("basket service listening", "address", c.Server.Address)
		errch <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
	case e = <-errch:
		if !errors.Is(e, http.ErrServerClosed) {
			slog.Error("HTTP server stopped", "error", e)
		}
	}
	shutdown, cancel := context.WithTimeout(context.Background(), c.Server.ShutdownTimeout)
	defer cancel()
	if e = srv.Shutdown(shutdown); e != nil {
		slog.Error("HTTP shutdown", "error", e)
	}
	if e = a.Close(shutdown); e != nil {
		slog.Error("application shutdown", "error", e)
	}
}
