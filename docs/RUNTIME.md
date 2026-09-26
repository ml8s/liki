# Runtime model

Liki currently ships one deterministic Go engine with two HTTP runtimes:

| Runtime | Interface | Purpose | Lifecycle |
| --- | --- | --- | --- |
| `engine-mcp` | Streamable MCP | Skill, counsel, and agent-facing compute API | Current architecture |
| `engine-rpc` | JSON-RPC 2.0 | The active `liki-web` free-chart API | Transition only |
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

The published engine image defaults to `engine-rpc` because the active
`liki-web` deployment still owns public `/jsonrpc`. MCP deployments override the
image entrypoint to `/usr/local/bin/engine-mcp`. This default changes back to
MCP only at the `engine-rpc` removal gate below.

## Process configuration

### engine-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8081` | Internal listen address |
| `LIKI_MCP_TOKEN` | empty | Optional Bearer gate for private deployments |
| `LIKI_TRUSTED_PROXY_HOPS` | `0` | Number of trusted reverse proxies |
| `LIKI_ALLOWED_ORIGINS` | hosted default | Comma-separated browser CORS origins |
| `LIKI_EXTERNAL_GEOCODING` | `on` | Allow Nominatim fallback for unknown cities |

`GET /health` and `GET /version` are public process probes. MCP requests can be
token-protected without exposing those probes.

### engine-rpc

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8080` | Transition JSON-RPC listen address |
| `LIKI_TRUSTED_PROXY_HOPS` | `0` | Rate-limiter proxy trust |
| `LIKI_ALLOWED_ORIGINS` | hosted default | Browser CORS origins |

`POST /jsonrpc` is the active contract used by `liki-web` free charts. It is not
a permanent compatibility surface: remove this runtime only after `liki-web`,
`liki-deploy`, and their E2E suites have moved to engine MCP.

### counsel-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LISTEN_ADDR` / uvicorn port | `:8086` | Internal listen address |
| `LIKI_MCP_URL` | local engine URL | Engine MCP base URL |
| `LIKI_MCP_TOKEN` | empty | Optional inbound Bearer gate |
| `LIKI_ENGINE_MCP_TOKEN` | falls back to `LIKI_MCP_TOKEN` | Optional outbound Bearer token for engine-mcp |
| `LIKI_MCP_TIMEOUT` | `30` | Engine call timeout in seconds |
| `LIKI_MCP_MAX_RETRIES` | `2` | Engine retry count |
| `LIKI_TRUSTED_PROXY_HOPS` | `0` | Number of trusted reverse proxies |
| `LIKI_COUNSEL_RATE_LIMIT` | `240` | Requests per rate window |
| `LIKI_COUNSEL_RATE_WINDOW_SECONDS` | `60` | Rate window |
| `LIKI_COUNSEL_RATE_MAX_KEYS` | `65536` | Bounded in-memory limiter keys |
| `LIKI_MCP_MAX_BODY_BYTES` | `1048576` | Maximum MCP JSON request body |

`GET /healthz` is liveness. `GET /readyz` checks that counsel and engine runtime
versions are compatible; use it for orchestration readiness.

## Deployment rules

1. Run engine and counsel as stateless containers.
2. Terminate TLS at the edge and set trusted-proxy hops explicitly.
3. Keep birth data in the current request/session context only.
4. Configure `LIKI_ALLOWED_ORIGINS` for browser clients on a private domain.
5. Protect private engine calls with `LIKI_ENGINE_MCP_TOKEN`; using the same
   value as counsel's `LIKI_MCP_TOKEN` remains supported.
6. Monitor `/readyz`, tool error rates, rate-limit rejections, and engine latency.

## Removal gate for engine-rpc

The JSON-RPC runtime may be deleted only when all of these are true:

1. `liki-web` consumes engine MCP for free charts.
2. `liki-deploy` no longer routes public `/jsonrpc`.
3. Local and CI E2E cover the MCP path.
4. A release note documents the endpoint removal.
5. No active deployment still reports `engine-rpc` traffic.
