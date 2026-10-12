# Feedback 运行规则

> Feedback 用于报告 Skill 机制、契约、口径或流程问题。它不收集用户内容，不阻塞主流程。

## 触发与上限

- 触发场景：错误、缺字段、口径冲突、流程卡涩、表述歧义。
- Agent 自主择时提交，不向用户请求确认。
- 失败不阻塞主流程。
- 同一会话最多提交 3 条。
- 提交后在最终收据中加一行：`Feedback: submitted|disabled|failed`。

## 请求配置

- 默认端点：`https://liki.hk/api/feedback`。
- `LIKI_FEEDBACK_URL` 可覆盖端点。
- `LIKI_FEEDBACK_DISABLED=1` 可禁用反馈。
- 未知宿主 / 模型值填 `unknown`。

## Payload 契约

固定模板见根目录 `feedback.schema.json`，`schema_version=feedback-v1`。

`problem.type` 只能是：

- `error`
- `gap`
- `conflict`
- `friction`
- `clarity`

`summary / expected / observed` 只写工程事实：

- 工具名；
- 字段名；
- 流程阶段；
- schema / contract 问题；
- 表述歧义。

## 隐私硬边界

禁止提交：

- 用户原文；
- 出生数据；
- 姓名；
- 地址；
- 卦题；
- 命盘；
- 工具全文。

## 提交方式

1. 把符合 `feedback.schema.json` 的 payload 写入临时文件。
2. 执行：

   ```bash
   python3 scripts/feedback.py --payload-file <临时文件>
   ```

3. 根据返回状态在最终收据中写 `Feedback: submitted|disabled|failed`。
4. 反馈失败时不要重试阻塞用户，也不要把内部错误详情展示为业务失败。
