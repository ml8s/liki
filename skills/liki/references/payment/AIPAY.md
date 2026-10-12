# Aipay 支付协议

> Liki 采用先使用、后付费模式。支付失败不追溯已交付结果；用户暂缓支付时确认即可，不重复建单。

## 状态检查

启动时执行：

```bash
python3 scripts/aipay.py status
```

- `paid=true`：已有有效收据，直接提供服务。
- `paid=false`：先说明“先使用、后付费”，交付结果后进入支付流程。

## `/api/aipay` 请求契约

- 方法：`POST`。
- 请求体：空即可，无需 JSON body。
- 必需请求头：浏览器风格 `User-Agent`。缺省客户端 UA 可能命中 Cloudflare 反爬，返回 `403` + error code `1010`。
  - 示例：`User-Agent: Mozilla/5.0 ... Safari/537.36`。
- 默认端点：`https://liki.hk/api/aipay`。

### 未支付请求

不带 `Payment-Proof` 头时，后端建单并返回：

- HTTP `402 Payment-Needed`。
- `Payment-Needed` 响应头。
- JSON：`{code, out_trade_no, amount, currency, goods_name}`。

### 支付核销请求

带 `Payment-Proof` 头时，后端核销支付凭证并返回：

- HTTP `200`。
- 收据 JSON；收据位于 `content` 内。
- `content.schema_version=liki-aipay-v1`。

## 支付执行

1. 收到 `402 Payment-Needed` 后，把请求摘要交给当前环境的 `alipay-payment-skill`。
2. `alipay-payment-skill` 不可用时，先征得用户同意，再按官方 `alipay/payment-skills` 指南安装。
3. 用户暂缓支付时回复确认；用户再次发起时从核销续走，不重复建单。

## 收据保存

支付完成或用户补付后，将后端响应通过 stdin 传给：

```bash
python3 scripts/aipay.py save-receipt
```

脚本会校验并原子保存收据。保存成功后再次执行 `status`，应以 `paid=true` 提供服务。

## 失败处理

- 后端返回 `402`：这是正常建单流程，不视为错误。
- 支付 skill 缺失：征得用户同意后安装。
- 收据校验失败：不要手写凭证文件；重新使用后端响应执行 `save-receipt`。
- 用户暂缓：确认即可，不阻塞后续会话。
