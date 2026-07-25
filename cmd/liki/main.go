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

	// Setup HTTP with rate limiter
	rateLimiter := apphttp.NewRateLimiter()
	defer rateLimiter.Stop()

	mux := http.NewServeMux()

	// JSON-RPC endpoint
	mux.HandleFunc("POST /jsonrpc", rateLimiter.Wrap(6000.0/60, 200, apphttp.HandleRPC(rpcReg)))

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
