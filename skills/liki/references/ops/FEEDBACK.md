# Feedback 运行规则

> 本文只约定请求配置、payload 契约与提交命令；触发条件、上限与隐私硬边界见根 `SKILL.md` 的 Feedback 一节。

## 请求配置

- 默认端点：`https://liki.hk/api/feedback`；`LIKI_FEEDBACK_URL` 可覆盖端点。
- `LIKI_FEEDBACK_DISABLED=1` 可禁用反馈。
- 未知宿主 / 模型值填 `unknown`。

## Payload 契约

固定模板见根目录 `feedback.schema.json`，`schema_version=feedback-v1`。

`problem.type` 只能是 `error`、`gap`、`conflict`、`friction`、`clarity`。

`summary / expected / observed` 只写工程事实：工具名、字段名、流程阶段、schema / contract 问题、表述歧义。

## 提交方式

1. 把符合 `feedback.schema.json` 的 payload 写入临时文件。
2. 执行：

   ```bash
   python3 scripts/feedback.py --payload-file <临时文件>
   ```

3. 按返回状态回写收据行（格式见根 `SKILL.md`）；失败不重试、不阻塞、不把内部错误展示为业务失败。
