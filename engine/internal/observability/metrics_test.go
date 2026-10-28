package observability

import (
	"strings"
	"testing"
)

func TestMetricsRenderFixedLowCardinalityLabels(t *testing.T) {
	ObserveHTTP("/mcp/{domain}", "POST", 200, 0.01)
	ObserveTool("bazi", "bazi_chart", "success", 0.02)

	text := string(RenderMetrics())
	for _, fragment := range []string{
		`liki_engine_http_requests_total{route="/mcp/{domain}",method="POST",status_class="2xx"} 1`,
		`liki_engine_tool_requests_total{domain="bazi",tool="bazi_chart",status="success"} 1`,
		"liki_engine_http_request_duration_seconds",
		"liki_engine_tool_request_duration_seconds",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("metrics missing %q:\n%s", fragment, text)
		}
	}
}

func TestMetricsUseCumulativePrometheusBuckets(t *testing.T) {
	ObserveTool("bazi", "bazi_chart", "success", 0.001)
	ObserveTool("bazi", "bazi_chart", "success", 0.003)

	text := string(RenderMetrics())
	if !strings.Contains(text, `le="0.0025"} 1`) {
		t.Fatalf("first bucket missing:\n%s", text)
	}
	if !strings.Contains(text, `le="0.005"} 2`) {
		t.Fatalf("cumulative bucket missing:\n%s", text)
	}
}
