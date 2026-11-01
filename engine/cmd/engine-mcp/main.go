package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tables"
	apphttp "liki-engine/internal/http"
)

// Version is injected at build time via -ldflags from skills/liki/VERSION.txt.
var Version = "dev"

func main() {
	addr := flag.String("addr", ":8081", "listen address (HTTP mode)")
	stdio := flag.Bool("stdio", false, "run on stdio instead of HTTP (local debugging)")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	// The MCP SDK logs one connection lifecycle event per stateless request.
	// Keep process logs at INFO, but keep that request noise below WARN.
	sdkLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	slog.SetDefault(logger)
	if err := tables.Load(); err != nil {
		slog.Error("engine: load data tables", "err", err)
		os.Exit(1)
	}
	if err := engineContracts.MaterializeSchemas(); err != nil {
		slog.Error("engine: materialize MCP contracts", "err", err)
		os.Exit(1)
	}

	if *stdio {
		slog.Info("engine-mcp stdio mode")
		if err := runStdio(Version, sdkLogger); err != nil {
			slog.Error("stdio", "err", err)
			os.Exit(1)
		}
		return
	}

	// MCP Streamable HTTP aggregate endpoint. The single Liki expert discovers
	// all capabilities once; methodology remains progressively disclosed by skill.
	mcpServer, err := newMCPServer(Version, sdkLogger)
	if err != nil {
		slog.Error("engine: build aggregate MCP server", "err", err)
		os.Exit(1)
	}
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	// 计算自检只做一次：排一个固定八字（1984-02-04 06:00 男），验证日柱=戊辰。
	// /health 必须保持轻量，不能成为 CPU 放大器。
	selfTestCst := time.FixedZone("CST", 8*3600)
	selfTestWhen := time.Date(1984, 2, 4, 6, 0, 0, 0, selfTestCst)
	selfTestChart := bazi.ComputeChart(
		time.Time(selfTestWhen),
		time.Time(selfTestWhen),
		ganzhi.Male,
	)
	selfTestOK := selfTestChart.Ri.Gan == ganzhi.GanWu

	mux := http.NewServeMux()
	// The aggregate endpoint is the canonical engine surface for the expert,
	// counsel, local tests, and deployment.
	mux.Handle("/mcp", apphttp.MetricsMiddleware("/mcp", mcpHandler))

	mux.Handle("GET /health", apphttp.MetricsMiddleware("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !selfTestOK {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"degraded","reason":"computation self-test failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			return
		}
	})))

	mux.Handle("GET /version", apphttp.MetricsMiddleware("/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"version":"` + Version + `"}`)); err != nil {
			return
		}
	})))

	metricsHandler := apphttp.MetricsHandler()
	mux.Handle("GET /metrics", apphttp.MetricsMiddleware("/metrics", metricsHandler))

	handler := apphttp.Recover(apphttp.SecurityHeaders(apphttp.CORSMiddleware(false, apphttp.BodyLimit(mux))))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenAddr := envOr("LIKI_HTTP_LISTEN_ADDRESS", *addr)

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      handler,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

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

	slog.Info("engine-mcp listening", "addr", srv.Addr, "endpoint", "/mcp")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
