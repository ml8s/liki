"""Thread-safe in-process Prometheus metrics.

Labels are limited to fixed tool/domain/method/status vocabularies. Never add
questions, chart payloads, birth data, endpoint URLs, tokens, or error text.
"""

from __future__ import annotations

import threading
from dataclasses import dataclass
from typing import Mapping


@dataclass
class _Counter:
    value: int = 0


@dataclass
class _Histogram:
    buckets: tuple[float, ...]
    counts: dict[float, int]
    sum: float = 0.0
    count: int = 0

    @classmethod
    def create(cls, buckets: tuple[float, ...]) -> "_Histogram":
        return cls(buckets=buckets, counts={bucket: 0 for bucket in buckets})

    def observe(self, value: float) -> None:
        self.sum += value
        self.count += 1
        for bucket in self.buckets:
            if value <= bucket:
                self.counts[bucket] += 1


_INFINITY = float("inf")
_BUCKETS = (
    0.001, 0.0025, 0.005, 0.01, 0.025, 0.05,
    0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
) + (_INFINITY,)
_LOCK = threading.Lock()
_tool_counters: dict[tuple[str, str, str], _Counter] = {}
_tool_histograms: dict[tuple[str, str, str], _Histogram] = {}
_engine_counters: dict[tuple[str, str], _Counter] = {}
_engine_histograms: dict[tuple[str, str], _Histogram] = {}
_http_counters: dict[tuple[str, str, str], _Counter] = {}
_http_histograms: dict[tuple[str, str, str], _Histogram] = {}


def reset_metrics() -> None:
    """Clear all process metrics; intended for tests only."""
    global _tool_counters, _tool_histograms
    global _engine_counters, _engine_histograms
    global _http_counters, _http_histograms
    with _LOCK:
        _tool_counters = {}
        _tool_histograms = {}
        _engine_counters = {}
        _engine_histograms = {}
        _http_counters = {}
        _http_histograms = {}


def observe_http_request(
    route: str, method: str, status_class: str, duration: float
) -> None:
    counter_key = (route, method, status_class)
    histogram_key = (route, method, status_class)
    with _LOCK:
        counter = _http_counters.setdefault(counter_key, _Counter())
        counter.value += 1
        histogram = _http_histograms.setdefault(
            histogram_key, _Histogram.create(_BUCKETS)
        )
        histogram.observe(duration)


def observe_tool_request(tool: str, domain: str, status: str, duration: float) -> None:
    with _LOCK:
        counter = _tool_counters.setdefault((tool, domain, status), _Counter())
        counter.value += 1
        histogram = _tool_histograms.setdefault(
            (tool, domain, status), _Histogram.create(_BUCKETS)
        )
        histogram.observe(duration)


def observe_engine_request(method: str, status: str, duration: float) -> None:
    with _LOCK:
        counter = _engine_counters.setdefault((method, status), _Counter())
        counter.value += 1
        histogram = _engine_histograms.setdefault(
            (method, status), _Histogram.create(_BUCKETS)
        )
        histogram.observe(duration)


def _bucket(value: float) -> str:
    if value == _INFINITY:
        return "+Inf"
    return str(int(value)) if value == int(value) else repr(value)


def _label_text(names: tuple[str, ...], values: tuple[str, ...]) -> str:
    return ",".join(f'{name}="{value}"' for name, value in zip(names, values, strict=True))


def _render_counter(
    name: str,
    description: str,
    counters: Mapping[tuple[str, ...], _Counter],
    labels: tuple[str, ...],
) -> list[str]:
    lines = [f"# HELP {name} {description}", f"# TYPE {name} counter"]
    for values in sorted(counters):
        rendered_labels = _label_text(labels, values)
        lines.append(f"{name}{{{rendered_labels}}} {counters[values].value}")
    return lines


def _render_histogram(
    name: str,
    description: str,
    histograms: Mapping[tuple[str, ...], _Histogram],
    labels: tuple[str, ...],
) -> list[str]:
    lines = [f"# HELP {name} {description}", f"# TYPE {name} histogram"]
    for values in sorted(histograms):
        histogram = histograms[values]
        base = _label_text(labels, values)
        for bucket, count in histogram.counts.items():
            le = _bucket(bucket)
            label_text = f"{base},le=\"{le}\"" if base else f"le=\"{le}\""
            lines.append(f"{name}_bucket{{{label_text}}} {count}")
        count_suffix = f"{{{base}}}" if base else ""
        lines.append(f"{name}_count{count_suffix} {histogram.count}")
        lines.append(f"{name}_sum{count_suffix} {histogram.sum:.9f}")
    return lines


def render_metrics() -> bytes:
    """Render a deterministic Prometheus text exposition payload."""
    with _LOCK:
        tool_counters = {
            key: _Counter(value=counter.value)
            for key, counter in _tool_counters.items()
        }
        tool_histograms = {
            key: _Histogram(
                buckets=histogram.buckets,
                counts=dict(histogram.counts),
                sum=histogram.sum,
                count=histogram.count,
            )
            for key, histogram in _tool_histograms.items()
        }
        engine_counters = {
            key: _Counter(value=counter.value)
            for key, counter in _engine_counters.items()
        }
        engine_histograms = {
            key: _Histogram(
                buckets=histogram.buckets,
                counts=dict(histogram.counts),
                sum=histogram.sum,
                count=histogram.count,
            )
            for key, histogram in _engine_histograms.items()
        }
        http_counters = {
            key: _Counter(value=counter.value)
            for key, counter in _http_counters.items()
        }
        http_histograms = {
            key: _Histogram(
                buckets=histogram.buckets,
                counts=dict(histogram.counts),
                sum=histogram.sum,
                count=histogram.count,
            )
            for key, histogram in _http_histograms.items()
        }

    lines = [
        *_render_counter(
            "liki_counsel_tool_requests_total",
            "Counsel MCP tool requests by tool, domain, and terminal status.",
            tool_counters,
            ("tool", "domain", "status"),
        ),
        *_render_histogram(
            "liki_counsel_tool_request_duration_seconds",
            "Counsel MCP tool request duration in seconds.",
            tool_histograms,
            ("tool", "domain", "status"),
        ),
        *_render_counter(
            "liki_counsel_engine_requests_total",
            "Counsel engine RPCs by method and terminal status.",
            engine_counters,
            ("method", "status"),
        ),
        *_render_histogram(
            "liki_counsel_engine_request_duration_seconds",
            "Counsel engine RPC duration in seconds.",
            engine_histograms,
            ("method", "status"),
        ),
        *_render_counter(
            "liki_counsel_http_requests_total",
            "Counsel HTTP requests by route template, method, and status class.",
            http_counters,
            ("route", "method", "status_class"),
        ),
        *_render_histogram(
            "liki_counsel_http_request_duration_seconds",
            "Counsel HTTP request duration in seconds.",
            http_histograms,
            ("route", "method", "status_class"),
        ),
    ]
    return ("\n".join(lines) + "\n").encode("utf-8")
