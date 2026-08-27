package main

import (
	"context"
	_ "embed"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"liki-engine/internal/agent"
	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	apphttp "liki-engine/internal/http"
)

// BuildTime is set at compile time via -ldflags. Defaults to VERSION file.
//
//go:embed VERSION
var versionFile string

var BuildTime = strings.TrimSpace(versionFile)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	// Structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Setup JSON-RPC registry
	rpcReg := agent.NewRPCRegistry()
	rpcReg.SetVersion(BuildTime)

	// Setup HTTP with rate limiter
	rateLimiter := apphttp.NewRateLimiter()
	defer rateLimiter.Stop()

	mux := http.NewServeMux()

	// JSON-RPC endpoint
	mux.HandleFunc("POST /jsonrpc", rateLimiter.Wrap(6000.0/60, 200, apphttp.HandleRPC(rpcReg)))

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 计算自检：排一个固定八字（1984-02-04 06:00 男），验证日柱=戊辰——确认引擎计算能力正常
		cst := time.FixedZone("CST", 8*3600)
		st := tianwen.SolarTime(time.Date(1984, 2, 4, 6, 0, 0, 0, cst))
		chart := bazi.ComputeChart(st, ganzhi.Male)
		if chart.Ri.Gan != ganzhi.GanWu {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"degraded","reason":"computation self-test failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			return
		}
	})

	// Version
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"version":"` + BuildTime + `"}`)); err != nil {
			return
		}
	})

	// Middleware stack
	handler := apphttp.Recover(apphttp.SecurityHeaders(apphttp.CORSMiddleware(false, apphttp.BodyLimit(mux))))

	// Context for server BaseContext and graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &http.Server{
		Addr:         envOr("LISTEN_ADDR", *addr),
		Handler:      handler,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Signal handling
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		slog.Info("received signal, shutting down", "signal", sig.String())
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("forced shutdown", "err", err)
		}
	}()

	slog.Info("liki engine listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
