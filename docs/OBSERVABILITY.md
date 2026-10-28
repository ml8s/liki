# Observability contract

## Endpoint

All Liki services expose exactly one metrics endpoint:

```text
GET /metrics
```

It is an internal operations endpoint. Public gateways must not expose it.

Do not use service-specific compatibility paths such as
`/api/internal/metrics`.

## Authentication

Use a dedicated metrics token:

```text
LIKI_METRICS_TOKEN
```

Do not reuse `LIKI_MCP_TOKEN`.

| Token state | Response |
| --- | --- |
| unset | `404` |
| missing or wrong bearer | `403` |
| correct bearer | `200` |

## Format

Services render Prometheus text exposition version `0.0.4`:

```text
Content-Type: text/plain; version=0.0.4; charset=utf-8
```

Implementations are dependency-free.

## Naming

```text
liki_<service>_<object>_<unit>
```

Counters end with `_total`. Durations use seconds.

## Required HTTP metrics

All HTTP services expose:

```text
liki_<service>_http_requests_total
liki_<service>_http_request_duration_seconds
```

Labels:

```text
route: fixed route template
method: HTTP method
status_class: 1xx / 2xx / 3xx / 4xx / 5xx
```

Never use raw paths, query strings, IDs, or user input as labels.

## Service tool metrics

### engine

```text
liki_engine_tool_requests_total
liki_engine_tool_request_duration_seconds
```

Labels:

```text
domain
tool
status
```

### counsel

```text
liki_counsel_tool_requests_total
liki_counsel_tool_request_duration_seconds
```

Labels:

```text
tool
domain
status
```

## Privacy

Metrics never contain:

```text
questions
birth data
chart payloads
session identifiers
user identifiers
URL query strings
tokens
exception text
```

## Conformance

Each repository tests:

```text
1. unset metrics token → 404
2. wrong token → 403
3. correct token → 200
4. output is Prometheus text
5. labels are fixed and low-cardinality
6. no payload or PII is present
```
