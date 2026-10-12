# Runtime model

Liki currently ships one deterministic Go engine with two HTTP runtimes:

| Runtime | Interface | Purpose | Lifecycle |
| --- | --- | --- | --- |
| `engine-mcp` | Streamable MCP | Skill, counsel, and agent-facing compute API | Current architecture |
| `engine-rpc` | JSON-RPC 2.0 | Compatibility API for installed legacy skills | Retained |
| `counsel-mcp` | Streamable MCP | Factor, assertion, divination, and naming judgment | Current architecture |

The edge owns public service prefixes:

```text
https://<host>/engine/mcp       → engine-mcp /mcp
https://<host>/engine/mcp/{domain} → engine-mcp /mcp/{domain}
https://<host>/counsel/mcp      → counsel-mcp /mcp
https://<host>/counsel/mcp/{domain} → counsel-mcp /mcp/{domain}
https://<host>/jsonrpc          → engine-rpc /jsonrpc
```

Services must not depend on the public prefix. The gateway strips `/engine` or
`/counsel` before proxying.

The published engine image defaults to `engine-rpc` because public `/jsonrpc`
remains available for installed legacy skills. MCP deployments override the
image entrypoint to `/usr/local/bin/engine-mcp`.

## Process configuration

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

### engine-rpc

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8080` | JSON-RPC compatibility listen address |
| `LIKI_ALLOWED_ORIGINS` | hosted default | Browser CORS origins |

`POST /jsonrpc` is not used by the web BFF. It is a retained compatibility
surface for installed legacy skills. Traffic data may inform product owners,
but it does not automatically trigger removal.

Removing `engine-rpc` requires an explicit product decision and follows the
public-entry-removal release policy. Until that decision is made, the binary,
the `/jsonrpc` entry point, and image smoke coverage remain part of the
supported runtime surface.

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
natal judgment tools (`compute_factors`, `natal_query`, and `period_query`) plus
naming, Liuyao, and Qimen tools. `/mcp/bazi` and `/mcp/ziwei` remain fixed-domain
endpoints for clients that need a narrower discovery surface.

## Deployment rules

1. Run engine and counsel as stateless containers.
2. Terminate TLS at the edge and set trusted-proxy hops explicitly.
3. Keep birth data in the current request/session context only.
4. Configure `LIKI_ALLOWED_ORIGINS` for browser clients on a private domain.
5. Protect private engine calls with `LIKI_ENGINE_MCP_TOKEN`; using the same
   value as counsel's `LIKI_MCP_TOKEN` remains supported.
6. Monitor `/readyz`, tool error rates, rate-limit rejections, and engine latency.
