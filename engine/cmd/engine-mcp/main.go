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
	resultContent, err := ParseResultContentMode(os.Getenv("LIKI_MCP_RESULT_CONTENT"))
	if err != nil {
		slog.Error("engine: parse result content mode", "err", err)
		os.Exit(1)
	}

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
		if err := runStdio(Version, sdkLogger, resultContent); err != nil {
			slog.Error("stdio", "err", err)
			os.Exit(1)
		}
		return
	}

	// MCP Streamable HTTP endpoints — 每术数一个域（排盘工具），aux 为共享辅助
	mcpServer, err := newMCPServer(Version, sdkLogger, resultContent)
	if err != nil {
		slog.Error("engine: build aggregate MCP server", "err", err)
		os.Exit(1)
	}
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	mcpToken := os.Getenv("LIKI_MCP_TOKEN")
	metricsToken := os.Getenv("LIKI_METRICS_TOKEN")

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
	// 全量端点（counsel 内部调用 + 兼容）
	mux.Handle("/mcp", apphttp.MetricsMiddleware("/mcp", apphttp.MCPAuthMiddleware(mcpToken, mcpHandler)))
	// 分域端点：每术数 + 共享辅助。`/engine` 外部前缀由网关剥离，
	// 本服务只负责 MCP 根路径和 MCP 域路径。
	for _, d := range mcpDomains {
		domainServer, err := newDomainServer(d.Prefixes, Version, sdkLogger, resultContent)
		if err != nil {
			slog.Error("engine: build domain MCP server", "domain", d.Suffix, "err", err)
			os.Exit(1)
		}
		domainHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return domainServer }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
		mux.Handle("/mcp/"+d.Suffix, apphttp.MetricsMiddleware("/mcp/{domain}", apphttp.MCPAuthMiddleware(mcpToken, domainHandler)))
	}

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
	mux.Handle("GET /metrics", apphttp.MetricsMiddleware("/metrics", apphttp.MetricsAuthMiddleware(metricsToken, metricsHandler)))

	handler := apphttp.Recover(apphttp.SecurityHeaders(apphttp.CORSMiddleware(false, apphttp.BodyLimit(mux))))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenAddr := envOr("LISTEN_ADDR", *addr)
	if err := apphttp.ValidateTokenPolicy(mcpToken, listenAddr, allowsNoToken()); err != nil {
		if allowsNoToken() {
			slog.Warn("engine-mcp: LIKI_MCP_TOKEN is empty on a non-loopback address; unauthenticated access is explicitly allowed")
		} else {
			slog.Error("engine-mcp: secure default rejected startup", "err", err)
			os.Exit(1)
		}
	}

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
