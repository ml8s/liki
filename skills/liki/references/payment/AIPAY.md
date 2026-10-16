# Aipay 支付协议

> 本文只约定状态检查、`/api/aipay` 请求契约、收据保存与失败处理；后付费使用步骤见根 `SKILL.md` 的 Aipay 一节。

## 状态检查

`python3 scripts/aipay.py status` 输出 `paid`：`true` 表示已有有效收据，直接提供服务；`false` 进入后付费流程。

## `/api/aipay` 请求契约

- 方法：`POST`；请求体空即可。
- 必需请求头：浏览器风格 `User-Agent`（缺省客户端 UA 可能命中 Cloudflare 反爬，返回 `403` + error code `1010`）。示例：`User-Agent: Mozilla/5.0 ... Safari/537.36`。
- 默认端点：`https://liki.hk/api/aipay`。
- 未带 `Payment-Proof`：后端建单，返回 HTTP `402 Payment-Needed`、`Payment-Needed` 响应头与 JSON `{code, out_trade_no, amount, currency, goods_name}`。
- 带 `Payment-Proof`：后端核销并返回 HTTP `200` 收据 JSON（位于 `content` 内），`content.schema_version=liki-aipay-v1`。

## 收据保存

支付完成或用户补付后，将后端响应通过 stdin 传给 `python3 scripts/aipay.py save-receipt`；脚本校验并原子保存，随后 `status` 应为 `paid=true`。

## 失败处理

- 后端返回 `402`：正常建单流程，不视为错误；支付失败不追溯已交付结果。
- `alipay-payment-skill` 缺失：先征得用户同意，再按官方 `alipay/payment-skills` 指南安装。
- 收据校验失败：不要手写凭证文件；用后端响应重新执行 `save-receipt`。
- 用户暂缓支付：确认即可，不阻塞后续会话；再次发起时从核销续走，不重复建单。
