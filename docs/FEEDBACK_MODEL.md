# 反馈领域模型与契约

本文定义 skill 运行期反馈的 `feedback-v1` 契约。反馈用于定位 skill、engine、agent 宿主和模型协作中的问题；它不是用户行为埋点，也不是用户画像采集。

当前仓库只定义客户端契约与 skill 分发内容，不包含反馈后端实现。

## 1. 设计原则

- 只保留能判断问题层级的字段：skill / engine / agent / model。
- 不追求完整观测；RPC 明细、全流程序列、性能指标和输出质量都不进入 v1。
- 不采集对话原文、用户身份、出生数据、完整命盘或完整 RPC 报文。
- 缺少宿主信息时写 `unknown`；LLM 不得自行猜测模型身份。
- 同一会话同一问题只提交一次；agent 自主择时，不请求用户确认且失败不阻塞；最终回执必须包含 `Feedback: submitted|disabled|failed`。

## 2. Schema 位置

统一 skill 内置一份 JSON Schema：

```text
skills/liki/feedback.schema.json
```

契约版本固定为 `feedback-v1`。字段结构或语义变化时升级版本。

## 3. 分组

| 组 | 必填 | 字段 | 用途 |
|---|---|---|---|
| `meta` | 是 | `source`, `skill`, `skill_version`, `engine_version`, `session_hash?` | 识别反馈来源、版本组合和会话去重 |
| `agent` | 是 | `name`, `version` | 判断是否为 agent 宿主或框架问题 |
| `llm` | 是 | `provider`, `model`, `model_version?` | 判断是否为模型能力或快照差异 |
| `problem` | 是 | `type`, `severity`, `summary`, `tool?`, `expected?`, `observed?` | 描述问题本体 |

`meta.source` 只能是 `skill-agent` 或 `user`；统一 skill 后 `meta.skill` 固定为 `liki`。
`session_hash` 只能是 SHA-256 摘要，不能是明文 session ID。去重指纹由后端基于会话与问题内容计算。

## 4. 问题类型

| 类型 | 含义 |
|---|---|
| `error` | 错误、调用失败、能力失效 |
| `gap` | 缺字段、缺文档、缺规则、缺能力 |
| `conflict` | 口径、文档、工具结果或领域事实冲突 |
| `friction` | 能走通但卡涩、重复、低效的流程 |
| `clarity` | 命名、表述、字段语义或解释歧义 |

## 5. 示例

```json
{
  "schema_version": "feedback-v1",
  "meta": {
    "source": "skill-agent",
    "skill": "liki",
    "skill_version": "x.y.z",
    "engine_version": "x.y.z"
  },
  "agent": {
    "name": "codex-cli",
    "version": "0.21.6"
  },
  "llm": {
    "provider": "openai",
    "model": "gpt-5.1"
  },
  "problem": {
    "type": "clarity",
    "severity": "medium",
    "tool": "ziwei.liuyue",
    "summary": "resolved_period 的解释入口不够直观",
    "expected": "结果契约或文档直接说明应解释哪个字段",
    "observed": "需要跨文档推断"
  }
}
```

## 6. 后端兼容

`schema_version`, `meta`, `agent`, `llm`, `problem` 是必填项。后端是否需要改动取决于现有 `/api/feedback` 的形状。

如果后端接受 JSON body、允许未知字段，并把反馈保存为 JSON / JSONB / 文本，则不需要迁移。  
如果后端使用固定 DTO 或 strict schema，需要做一次 additive update：保留旧字段，新增可选的 `feedback-v1` 字段。

第一版数据库可以只用一个 payload 字段：

```text
feedback(id, schema_version, payload_json, created_at)
```

## 7. 运行治理

统一 skill 内置一个 sender：

```text
feedback.py
```

agent 可通过 stdin 传入 problem / meta / agent / llm payload。sender 会补齐 contract 默认值、校验 payload，并提交到 endpoint。

治理规则：

- 默认 endpoint：`https://liki.hk/api/feedback`
- 覆盖 endpoint：`LIKI_FEEDBACK_URL=https://your-host.example/api/feedback`
- 禁用：`LIKI_FEEDBACK_DISABLED=1`
- timeout：2 秒；失败不重试、不阻塞
- 同一会话最多 3 条
- 宿主可信覆盖：`LIKI_FEEDBACK_SKILL`、`LIKI_FEEDBACK_SKILL_VERSION`、`LIKI_ENGINE_VERSION`、`LIKI_FEEDBACK_SESSION_HASH`
- 上下文只来自显式 payload 和宿主覆盖变量；sender 不读取额外本地文件
- 最终优先级：宿主覆盖变量 > 显式 payload > sender 默认值
- payload 上限：32 KiB

backend 必须独立执行 payload 大小限制、rate limit、基础 PII 扫描与会话去重。

## 8. 隐私边界

禁止采集：

- 对话原文；
- 用户姓名、账号、明文 session ID；
- 出生时间、出生地点；
- 完整命盘；
- 完整 RPC request / response；
- 断语输出全文；
- IP、hostname、本地路径、API key。

只允许技术摘要、版本号、受控枚举和非隐私问题描述。`summary / expected / observed` 不得复述用户原文，也不得包含出生数据、姓名、地点、卦题、命盘或 RPC 全文。
