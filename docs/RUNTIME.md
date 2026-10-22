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

## Time facts contract

All engine tools that accept a precise time use the same two wire fields:

| Field | Rule |
| --- | --- |
| `instant` | RFC3339 absolute instant. Canonical output is UTC and ends with `Z`. Numeric offsets such as `+08:00` are accepted and normalized. |
| `timezone` | Explicit IANA timezone (for example `Asia/Shanghai`). `Local`, fixed offsets, and abbreviations are rejected. |

The offset in the input is a serialization detail; the engine normalizes it to
UTC immediately. The explicit `timezone` remains authoritative for local civil
time and offset. Tools that need true-solar time also accept `longitude`.

Do not infer `timezone` from an input offset. Do not default a missing
timezone to `Asia/Shanghai`. Missing timezone or longitude is fail-closed for
any capability that needs it.

## Process configuration

## Runtime settings

### engine-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8081` | Internal listen address |
| `LIKI_MCP_TOKEN` | empty | Optional Bearer gate for private deployments |
| `LIKI_ALLOWED_ORIGINS` | hosted default | Comma-separated browser CORS origins |
| `LIKI_EXTERNAL_GEOCODING` | `on` | Allow Nominatim fallback for unknown cities |

`GET /health` and `GET /version` are public process probes. MCP requests can be
token-protected without exposing those probes. Rate limiting and X-Forwarded-For
trust are owned by the edge (cloud LB / WAF / Caddy); the service itself does
not rate-limit.

### counsel-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` / uvicorn port | `:8086` | Internal listen address |
| `LIKI_MCP_URL` | *(required)* | Engine MCP base URL (fail-fast at startup when empty) |
| `LIKI_MCP_TOKEN` | empty | Optional inbound Bearer gate |
| `LIKI_ENGINE_MCP_TOKEN` | falls back to `LIKI_MCP_TOKEN` | Optional outbound Bearer token for engine-mcp |
| `LIKI_MCP_TIMEOUT` | `30` | Engine call timeout in seconds |
| `LIKI_MCP_MAX_RETRIES` | `2` | Engine retry count |
| `LIKI_MCP_MAX_BODY_BYTES` | `1048576` | Maximum MCP JSON request body |

`GET /healthz` is liveness. `GET /readyz` checks that counsel and engine runtime
versions are compatible; use it for orchestration readiness. Rate limiting is
owned by the edge (cloud LB / WAF / Caddy); the service itself does not
rate-limit.

`/mcp` is the aggregate surface for the multi-expert runtime. It exposes the
natal judgment tools (`bazi_yongshen`, `compute_factors`, `natal_query`, and
`period_query`) plus naming, Liuyao, and Qimen tools. `/mcp/bazi` and
`/mcp/ziwei` remain fixed-domain endpoints for clients that need a narrower
discovery surface; `bazi_yongshen` is published only on the bazi and aggregate
surfaces.

## Deployment rules

1. Run engine and counsel as stateless containers.
2. Terminate TLS at the edge and set trusted-proxy hops explicitly.
3. Keep birth data in the current request/session context only.
4. Configure `LIKI_ALLOWED_ORIGINS` for browser clients on a private domain.
5. Protect private engine calls with `LIKI_ENGINE_MCP_TOKEN`; using the same
   value as counsel's `LIKI_MCP_TOKEN` remains supported.
6. Monitor `/readyz`, tool error rates, rate-limit rejections, and engine latency.
