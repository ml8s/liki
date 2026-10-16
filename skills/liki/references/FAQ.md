# Liki FAQ

> 运行失败、依赖缺失、版本 / digest / schema 校验失败、输出不确定时先读本文件；不得绕过校验，不得自行降级。

支付与反馈失败分别见 `references/payment/AIPAY.md`、`references/ops/FEEDBACK.md`。

## MCP 前置

- MCP 未连接或端点不可达时，明确说明「MCP 不可用」；不用模型记忆补齐，不虚构盘面或断语。
- MCP 工具返回 `isError` 时按错误消息修正参数重试；确定性校验错误必须先修参数，不得绕过。
- 引擎版本低于本地 `VERSION.txt` 时停止；不得用旧结果、缓存或模型知识补齐引擎事实。

## 缺输入与失败

- 时辰模糊、接近交界或用户不确定时先校准；不得默认时辰。
- 城市未收录时请求附近较大城市；仍缺经度时相关校正输出「不可用」。
- 只有网络超时可重试。
- 流程中断后恢复时，从当前 ⛔ / 💬 步骤继续；不跳过阻塞确认。

## 版本与升级

- 提示升级时执行 `npx skills add ml8s/liki -y`；升级中断或网络失败时重跑同一命令，不手改本地文件。
- 请求远程版本时携带浏览器 / curl 风格 `User-Agent`（Cloudflare 会拦截默认脚本 UA，返回 `403` + error code 1010）；`curl` 默认 UA 即可。
- GitHub 直连受限（国内网络常见）时，`npx skills add ml8s/liki -y` 可能因内部 git clone 超时失败。优先改用官方托管源：`npx skills add https://liki.hk -y`。它通过 liki.hk 的 well-known 清单下载 archive 并校验 digest，不依赖 GitHub clone。仍需保证 npx / npm registry 可用；不要手改本地 Skill 文件。
