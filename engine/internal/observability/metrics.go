// Package observability provides the engine's dependency-free Prometheus
// metrics. Labels are restricted to fixed, low-cardinality vocabulary.
package observability

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type counter struct {
	value uint64
}

type histogram struct {
	buckets []float64
	counts  []uint64
	sum     float64
	count   uint64
}

var buckets = []float64{
	0.001, 0.0025, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

var (
	mu             sync.Mutex
	httpCounters   = map[httpKey]*counter{}
	httpHistograms = map[httpKey]*histogram{}
	toolCounters   = map[toolKey]*counter{}
	toolHistograms = map[toolKey]*histogram{}
)

type httpKey struct {
	route       string
	method      string
	statusClass string
}

type toolKey struct {
	domain string
	tool   string
	status string
}

// ObserveHTTP records one completed HTTP request. route must be a route
// template, never a raw path with identifiers or query values.
func ObserveHTTP(route, method string, status int, seconds float64) {
	key := httpKey{
		route:       route,
		method:      method,
		statusClass: statusClass(status),
	}
	mu.Lock()
	defer mu.Unlock()
	httpCounter, ok := httpCounters[key]
	if !ok {
		httpCounter = &counter{}
		httpCounters[key] = httpCounter
	}
	httpCounter.value++
	httpHistogram, ok := httpHistograms[key]
	if !ok {
		httpHistogram = &histogram{buckets: buckets, counts: make([]uint64, len(buckets))}
		httpHistograms[key] = httpHistogram
	}
	httpHistogram.observe(seconds)
}

// ObserveTool records one completed engine tool invocation.
func ObserveTool(domain, tool, status string, seconds float64) {
	key := toolKey{domain: domain, tool: tool, status: status}
	mu.Lock()
	defer mu.Unlock()
	toolCounter, ok := toolCounters[key]
	if !ok {
		toolCounter = &counter{}
		toolCounters[key] = toolCounter
	}
	toolCounter.value++
	toolHistogram, ok := toolHistograms[key]
	if !ok {
		toolHistogram = &histogram{buckets: buckets, counts: make([]uint64, len(buckets))}
		toolHistograms[key] = toolHistogram
	}
	toolHistogram.observe(seconds)
}

func (h *histogram) observe(seconds float64) {
	h.sum += seconds
	h.count++
	for index, bound := range h.buckets {
		if seconds <= bound {
			h.counts[index]++
		}
	}
}

func statusClass(status int) string {
	if status < 100 || status > 599 {
		return "unknown"
	}
	return fmt.Sprintf("%dxx", status/100)
}

func renderCounter(
	name, help string,
	counters map[string]uint64,
	labelNames []string,
) []string {
	lines := []string{fmt.Sprintf("# HELP %s %s", name, help), fmt.Sprintf("# TYPE %s counter", name)}
	keys := make([]string, 0, len(counters))
	for key := range counters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s{%s} %d", name, labels(key, labelNames), counters[key]))
	}
	return lines
}

func renderHistogram(
	name, help string,
	histograms map[string]*histogram,
	labelNames []string,
) []string {
	lines := []string{fmt.Sprintf("# HELP %s %s", name, help), fmt.Sprintf("# TYPE %s histogram", name)}
	keys := make([]string, 0, len(histograms))
	for key := range histograms {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		item := histograms[key]
		base := labels(key, labelNames)
		for index, bound := range item.buckets {
			le := formatFloat(bound)
			labelText := fmt.Sprintf("%s,le=%q", base, le)
			if base == "" {
				labelText = fmt.Sprintf("le=%q", le)
			}
			lines = append(lines, fmt.Sprintf("%s_bucket{%s} %d", name, labelText, item.counts[index]))
		}
		le := "+Inf"
		labelText := fmt.Sprintf("%s,le=%q", base, le)
		if base == "" {
			labelText = fmt.Sprintf("le=%q", le)
		}
		lines = append(lines, fmt.Sprintf("%s_bucket{%s} %d", name, labelText, item.count))
		lines = append(lines, fmt.Sprintf("%s_count{%s} %d", name, labels(key, labelNames), item.count))
		lines = append(lines, fmt.Sprintf("%s_sum{%s} %f", name, labels(key, labelNames), item.sum))
	}
	return lines
}

func labels(rawKey string, names []string) string {
	pairs := strings.Split(rawKey, "\x1f")
	escaped := make([]string, 0, len(pairs))
	for index, value := range pairs {
		if index < len(names) {
			escaped = append(escaped, fmt.Sprintf("%s=%q", names[index], value))
		}
	}
	return strings.Join(escaped, ",")
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%g", value)
}

// RenderMetrics returns Prometheus text exposition.
func RenderMetrics() []byte {
	mu.Lock()
	httpCountersCopy := make(map[string]uint64, len(httpCounters))
	for key, counter := range httpCounters {
		httpCountersCopy[httpCounterKey(key)] = counter.value
	}
	httpHistogramsCopy := make(map[string]*histogram, len(httpHistograms))
	for key, histogram := range httpHistograms {
		httpHistogramsCopy[httpCounterKey(key)] = histogram.clone()
	}
	toolCountersCopy := make(map[string]uint64, len(toolCounters))
	for key, counter := range toolCounters {
		toolCountersCopy[toolCounterKey(key)] = counter.value
	}
	toolHistogramsCopy := make(map[string]*histogram, len(toolHistograms))
	for key, histogram := range toolHistograms {
		toolHistogramsCopy[toolCounterKey(key)] = histogram.clone()
	}
	mu.Unlock()

	lines := renderCounter(
		"liki_engine_http_requests_total",
		"Completed engine HTTP requests.",
		httpCountersCopy,
		[]string{"route", "method", "status_class"},
	)
	lines = append(lines, renderHistogram(
		"liki_engine_http_request_duration_seconds",
		"Engine HTTP request duration in seconds.",
		httpHistogramsCopy,
		[]string{"route", "method", "status_class"},
	)...)
	lines = append(lines, renderCounter(
		"liki_engine_tool_requests_total",
		"Completed engine tool invocations.",
		toolCountersCopy,
		[]string{"domain", "tool", "status"},
	)...)
	lines = append(lines, renderHistogram(
		"liki_engine_tool_request_duration_seconds",
		"Engine tool invocation duration in seconds.",
		toolHistogramsCopy,
		[]string{"domain", "tool", "status"},
	)...)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func (h *histogram) clone() *histogram {
	counts := make([]uint64, len(h.counts))
	copy(counts, h.counts)
	return &histogram{buckets: h.buckets, counts: counts, sum: h.sum, count: h.count}
}

func httpCounterKey(key httpKey) string {
	return strings.Join([]string{key.route, key.method, key.statusClass}, "\x1f")
}

func toolCounterKey(key toolKey) string {
	return strings.Join([]string{key.domain, key.tool, key.status}, "\x1f")
}

// ResetMetrics clears all process metrics; intended for tests only.
func ResetMetrics() {
	mu.Lock()
	defer mu.Unlock()
	httpCounters = map[httpKey]*counter{}
	httpHistograms = map[httpKey]*histogram{}
	toolCounters = map[toolKey]*counter{}
	toolHistograms = map[toolKey]*histogram{}
}
