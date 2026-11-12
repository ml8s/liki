# Runtime model

Liki currently ships one deterministic Go engine with an MCP runtime:

| Runtime | Interface | Purpose | Lifecycle |
| --- | --- | --- | --- |
| `engine-mcp` | Streamable MCP | Internal chart / fact API; also used by WEB chart profile | Current architecture |
| `counsel-mcp` | Streamable MCP | Internal factor, assertion, divination, and naming judgment API | Current architecture |
| `liki-mcp` | Streamable MCP | Product-facing aggregate MCP with SkillPay paywall over selected ENGINE + COUNSEL tools; EXPERT uses ENGINE + COUNSEL directly | Current architecture |

`engine-rpc` and its public JSON-RPC endpoint are retired.

The edge owns public service prefixes:

```text
https://<host>/mcp              → liki-mcp /mcp
https://<host>/engine/mcp       → engine-mcp /mcp, technical/internal only
https://<host>/counsel/mcp      → counsel-mcp /mcp, technical/internal only
```

Services must not depend on the public prefix. The edge strips `/engine` or
`/counsel` before proxying.

`liki-mcp` is a separate image and process. Its container exposes
`/mcp`, reads its bundled `app/expose.json`, and routes only the configured product
allowlist to internal ENGINE / COUNSEL endpoints. Keep ENGINE / COUNSEL
endpoints off the public SKILL discovery declaration.

The published engine image defaults to `engine-mcp`.
Performance policy and benchmark entry points are described in
[PERFORMANCE.md](./PERFORMANCE.md).

## Time facts contract

Time inputs are split into two public forms:

| Tool layer | Fields | Rule |
| --- | --- | --- |
| `subject_candidates` | `calendar_kind` + `recorded_date`/`recorded_lunar` + `recorded_time`/`recorded_shichen`/`recorded_time_range` + `record_timezone` + `longitude` | Civil wall-clock input. DST is resolved once; fold returns two candidates and gap fails. |
| Chart/flow tools | `standard_time`, plus `longitude` only where LAST is required | Local standard time (no DST), RFC3339 with a known offset. `Z` is valid UTC standard time; `-00:00` is rejected as unknown offset. `timezone`, `instant`, and `fold` are rejected. Huangli and Xuankong do not accept longitude. |

`standard_time` is not an IANA timezone and its offset is not used to infer
one. The offset selects the local standard representation: two RFC3339 values
can identify the same instant while projecting different standard dates.
Tools that need true-solar time also accept `longitude`.

Do not default a missing timezone to `Asia/Shanghai`. Missing timezone at
`subject_candidates`, or missing longitude where LAST is required, is fail-closed for
any capability that needs it.

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
| `LIKI_BIND_HOST` | `127.0.0.1` | Internal bind host (uvicorn port is hard-coded `8086`) |
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

### liki-mcp

| Variable | Default | Purpose |
| --- | --- | --- |
| `LIKI_BIND_HOST` / uvicorn port | `127.0.0.1` / `8087` | Gateway bind host and port |
| `LIKI_BIND_PORT` | `8087` | Gateway listen port (uvicorn) |
| `LIKI_UPSTREAM_ENGINE_URL` | `http://engine-mcp:8081/mcp` | Internal ENGINE MCP endpoint |
| `LIKI_UPSTREAM_COUNSEL_URL` | `http://counsel-mcp:8086/mcp` | Internal COUNSEL MCP endpoint |
| `LIKI_UPSTREAM_ENGINE_TIMEOUT_SECONDS` | `35` | ENGINE tools/call timeout |
| `LIKI_UPSTREAM_COUNSEL_TIMEOUT_SECONDS` | `35` | COUNSEL tools/call timeout |
| `LIKI_VERSION` | `0.1.0` | Runtime version reported by Liki MCP |
| `LIKI_MCP_SERVER_NAME` | `hk.liki/liki` | Server name in MCP `initialize` response |
| `LIKI_MCP_SERVER_TITLE` | `Liki MCP` | Server title in MCP `initialize` response |
| `LIKI_PAYWALL_ENABLED` | `false` | Enable AI A2M paywall for paid tools |
| `ALIPAY_APP_ID` | — | Alipay application ID (required when paywall enabled) |
| `ALIPAY_PRIVATE_KEY_PKCS1` | — | Alipay application private key for the Python SDK (PKCS#1) |
| `ALIPAY_PUBLIC_KEY` | — | Alipay public key |
| `ALIPAY_GATEWAY` | sandbox | Alipay gateway URL |
| `ALIPAY_SELLER_ID` | — | Alipay seller ID (2088 format) |
| `ALIPAY_SERVICE_ID` | — | Alipay service marketplace ID |
| `ALIPAY_AIPAY_AMOUNT` | `9.90` | Per-consultation fee in CNY |
| `ALIPAY_PAID_TOOLS` | `subject_candidates` | Comma-separated paid tool names |
| `LIKI_AIPAY_DB_PATH` | `/data/aipay.db` | SQLite order store path (mount volume in production) |
| `LIKI_ALIPAY_SANDBOX_CONFIG` | unset | Optional absolute path to a local sandbox profile; never loaded implicitly |

`GET /healthz` is product MCP liveness. `GET /readyz` pings both upstream sessions
(engine and counsel); if either is unreachable, it returns 503. `GET /metrics`
is always enabled inside the service network and must be blocked by the public
edge. Rate limiting is owned by the edge (cloud LB / WAF / Caddy); the product
MCP itself does not rate-limit.

LIKI-MCP is the public SKILL MCP surface with SkillPay paywall.
EXPERT (agent-runtime) connects directly to ENGINE and COUNSEL,
bypassing the product MCP and its paywall. It exposes the selected
ENGINE chart / fact tools, COUNSEL judgment tools, naming tools, and
`liuyao_ask` / `qimen_ask`. Raw `qimen_chart`, `liuyao_chart`, and
`liuyao_qigua` are not exposed. ENGINE / COUNSEL endpoints remain technical
services for WEB, tests, and internal composition.

## Deployment rules

LLM clients may compose arguments from explicit user facts and prior MCP
structured results. `subject_candidates` resolves uncertain or ambiguous civil
time and should be followed by user confirmation when it returns multiple
candidates. Canonical outputs may be copied into later inputs, but there is no
mandatory source digest or provenance binding.

1. Run engine and counsel as stateless containers.
2. Terminate TLS at the edge and set trusted-proxy hops explicitly.
3. Keep birth data in the current request/session context only.
4. Configure `LIKI_ENGINE_ALLOWED_ORIGINS` for browser clients on a private domain.
5. Keep MCP public and stateless; do not add service tokens unless a new
   capability carries user data, side effects, or direct cost.
6. Monitor `/readyz`, `/metrics`, tool error rates, rate-limit rejections, and engine latency. See [OBSERVABILITY.md](./OBSERVABILITY.md).
