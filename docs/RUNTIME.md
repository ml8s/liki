# Runtime model

Liki currently ships one deterministic Go engine with an MCP runtime:

| Runtime | Interface | Purpose | Lifecycle |
| --- | --- | --- | --- |
| `engine-mcp` | Streamable MCP | Skill, counsel, and agent-facing compute API | Current architecture |
| `counsel-mcp` | Streamable MCP | Factor, assertion, divination, and naming judgment | Current architecture |

`engine-rpc` and its public JSON-RPC endpoint are retired.

The edge owns public service prefixes:

```text
https://<host>/engine/mcp       → engine-mcp /mcp
https://<host>/engine/mcp/{domain} → engine-mcp /mcp/{domain}
https://<host>/counsel/mcp      → counsel-mcp /mcp
https://<host>/counsel/mcp/{domain} → counsel-mcp /mcp/{domain}
```

Services must not depend on the public prefix. The gateway strips `/engine` or
`/counsel` before proxying.

The published engine image defaults to `engine-mcp`.
Performance policy and benchmark entry points are described in
[PERFORMANCE.md](./PERFORMANCE.md).

## Time facts contract

Time inputs are split into two public forms:

| Tool layer | Fields | Rule |
| --- | --- | --- |
| `tianwen.resolve` | `date`/`lunar` + `time`/`shichen`/`range` + `timezone` + `longitude` | Civil wall-clock input. DST is resolved once; fold returns two candidates and gap fails. |
| Chart/flow tools | `standard_time` + `longitude` | Local standard time (no DST), RFC3339 with a known offset. `Z` is valid UTC standard time; `-00:00` is rejected as unknown offset. `timezone`, `instant`, and `fold` are rejected. |

`standard_time` is not an IANA timezone and its offset is not used to infer
one. The offset selects the local standard representation: two RFC3339 values
can identify the same instant while projecting different standard dates.
Tools that need true-solar time also accept `longitude`.

Do not default a missing timezone to `Asia/Shanghai`. Missing timezone at
`tianwen.resolve`, or missing longitude where LAST is required, is fail-closed for
any capability that needs it.

## Process configuration

## Runtime settings

### engine-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LIKI_HTTP_LISTEN_ADDRESS` | `:8081` | Internal listen address |
| `LIKI_ENGINE_ALLOWED_ORIGINS` | hosted default | Comma-separated browser CORS origins |

`GET /health` and `GET /version` are public process probes. MCP requests are
public. `GET /metrics` is an internal operations endpoint and must be blocked by
the public edge. Rate limiting and X-Forwarded-For trust are owned by the edge
(cloud LB / WAF / Caddy); the service itself does not rate-limit.

### counsel-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LIKI_BIND_HOST` / uvicorn port | `127.0.0.1` / `8086` | Internal bind host and port |
| `LIKI_ENGINE_MCP_URL` | *(required)* | Engine MCP base URL (fail-fast at startup when empty) |
| `LIKI_ENGINE_MCP_TIMEOUT_SECONDS` | `30` | Engine call timeout in seconds |
| `LIKI_READYZ_TIMEOUT_SECONDS` | `1` | Short engine dependency timeout used by counsel `/readyz` |
| `LIKI_ENGINE_MCP_MAX_RETRIES` | `2` | Engine retry count |
| `LIKI_HTTP_IDLE_TIMEOUT_SECONDS` | `55` | Idle seconds before a worker closes an outbound engine keep-alive connection |
| `LIKI_HTTP_MAX_BODY_BYTES` | `1048576` | Maximum HTTP request body |
| `LIKI_VERSION` | build-injected | Runtime version reported by counsel |

`GET /healthz` is counsel liveness and does not contact the engine. `GET /readyz`
performs one bounded `server/discover` call to the engine; if that dependency is
unreachable, it returns 503. `GET /metrics` exposes low-cardinality Prometheus
counters and histograms for counsel HTTP/tool outcomes and engine RPC
latency/duration. It is always enabled inside the service network and must be
blocked by the public edge. Metrics never contain questions,
birth data, chart payloads, endpoint URLs, tokens, or exception text. Rate
limiting is owned by the edge (cloud LB / WAF / Caddy); the service itself does
not rate-limit.

`/mcp` is the aggregate surface for the single-expert aggregate runtime. It exposes the
natal judgment tools (`bazi_yongshen`, `compute_factors`, `natal_query`, and
`period_query`) plus naming, Liuyao, and Qimen tools. `/mcp/bazi` and
`/mcp/ziwei` remain fixed-domain endpoints for clients that need a narrower
discovery surface; `bazi_yongshen` is published only on the bazi and aggregate
surfaces.

## Deployment rules

1. Run engine and counsel as stateless containers.
2. Terminate TLS at the edge and set trusted-proxy hops explicitly.
3. Keep birth data in the current request/session context only.
4. Configure `LIKI_ENGINE_ALLOWED_ORIGINS` for browser clients on a private domain.
5. Keep MCP public and stateless; do not add service tokens unless a new
   capability carries user data, side effects, or direct cost.
6. Monitor `/readyz`, `/metrics`, tool error rates, rate-limit rejections, and engine latency. See [OBSERVABILITY.md](./OBSERVABILITY.md).
