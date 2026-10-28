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
| `LISTEN_ADDR` | `:8081` | Internal listen address |
| `LIKI_MCP_TOKEN` | empty | Optional Bearer gate for private deployments |
| `LIKI_ALLOWED_ORIGINS` | hosted default | Comma-separated browser CORS origins |
| `LIKI_MCP_RESULT_CONTENT` | `structured` | Successful tool result shape: `structured` or legacy `compat` with duplicate TextContent |

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
| `LIKI_MCP_HTTP_IDLE_TIMEOUT` | `55` | Idle seconds before a worker closes an outbound engine keep-alive connection |
| `LIKI_MCP_MAX_BODY_BYTES` | `1048576` | Maximum MCP JSON request body |
| `LIKI_VERSION` | build-injected | Runtime version reported by counsel |

`GET /healthz` is liveness. `GET /readyz` reports that the counsel service is up; it does not contact the engine. Rate limiting is
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
