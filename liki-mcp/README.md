# LIKI-MCP

Thin profile-free MCP passthrough gateway for SKILL / EXPERT.

## Contract

- `tools/list` aggregates the fixed ENGINE + COUNSEL expose allowlist.
- Tool descriptors and schemas are passed through unchanged.
- `tools/call` routes by exact tool name.
- No raw Qimen / Liuyao upstream chart tools are exposed.
- Upstream errors are passed through.
- Calls are never retried.

## Configuration

| Variable | Default |
| --- | --- |
| `LIKI_BIND_HOST` | `127.0.0.1` |
| `LIKI_BIND_PORT` | `8087` |
| `LIKI_VERSION` | `0.1.0` |
| `LIKI_UPSTREAM_ENGINE_URL` | `http://engine-mcp:8081/mcp` |
| `LIKI_UPSTREAM_COUNSEL_URL` | `http://counsel-mcp:8086/mcp` |
| `LIKI_EXPOSE_CONFIG` | built-in `expose.json` |
