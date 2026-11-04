# LIKI-MCP

Product-facing MCP gateway for SKILL with SkillPay paywall.
EXPERT (agent-runtime) connects directly to ENGINE and COUNSEL.

## Contract

- `tools/list` aggregates the fixed ENGINE + COUNSEL expose allowlist.
- Tool descriptors and schemas are passed through unchanged.
- `tools/call` routes by exact tool name.
- No raw Qimen / Liuyao upstream chart tools are exposed.
- Upstream errors are passed through.
- Calls are never retried.
- Paid tools (default: `subject_candidates`) require Payment-Proof.
  Without proof, the gateway returns HTTP 402 + Payment-Needed (SkillPay A2M protocol).
  The agent runtime delegates to alipay-payment-skill for payment.

## Configuration

| Variable | Default |
| --- | --- |
| `LIKI_BIND_HOST` | `127.0.0.1` |
| `LIKI_BIND_PORT` | `8087` |
| `LIKI_VERSION` | `0.1.0` |
| `LIKI_UPSTREAM_ENGINE_URL` | `http://engine-mcp:8081/mcp` |
| `LIKI_UPSTREAM_COUNSEL_URL` | `http://counsel-mcp:8086/mcp` |
| `LIKI_EXPOSE_CONFIG` | built-in `expose.json` |
| `LIKI_PAYWALL_ENABLED` | `false` |
| `ALIPAY_APP_ID` | — |
| `ALIPAY_PRIVATE_KEY` | — |
| `ALIPAY_PUBLIC_KEY` | — |
| `ALIPAY_GATEWAY` | sandbox gateway |
| `ALIPAY_SELLER_ID` | — |
| `ALIPAY_SERVICE_ID` | — |
| `ALIPAY_AIPAY_AMOUNT` | `9.90` |
| `ALIPAY_PAID_TOOLS` | `subject_candidates` |
| `LIKI_AIPAY_DB_PATH` | `/data/aipay.db` |

## Payment

When `LIKI_PAYWALL_ENABLED=true`, paid tools require a valid Payment-Proof
header. The gateway uses the official `alipay-sdk-python` SDK for
A2M (Agent-to-Merchant) protocol: order creation, proof verification,
and fulfillment confirmation. Order state is persisted in SQLite.

With `LIKI_PAYWALL_ENABLED=false` (expert deployment), the gateway is
a pure passthrough and the alipay SDK is never loaded.
