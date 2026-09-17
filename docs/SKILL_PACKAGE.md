# Unified Skill 包结构

Liki 对外只有一个可安装 skill：`liki`。它是产品入口，不是四个领域的拼接 prompt。

## 目录契约

```text
skills/liki/
├── SKILL.md              # 唯一 skill 入口
├── VERSION.txt           # 唯一分发版本
├── FAQ.md                # 唯一运行失败 / 反模式恢复入口
├── requirements.txt      # 唯一 Python 依赖清单
├── feedback.py           # 唯一 feedback sender
├── feedback.schema.json  # feedback-v1 contract
├── bazi/ENTRY.md         # 八字 + 紫微领域入口
├── bazi/TOOLS.md         # Python 工具完整 stdin 报文
├── divination/ENTRY.md   # 六爻 + 奇门 + 黄历领域入口
├── divination/TOOLS.md   # Python 工具完整 stdin 报文
├── fengshui/ENTRY.md     # 八宅 + 玄空领域入口
├── fengshui/RPC.md       # 固定 JSON-RPC 完整报文
├── naming/ENTRY.md       # 起名领域入口
└── naming/RPC.md         # 固定 JSON-RPC 完整报文
```

## 分层

| 层 | 职责 |
|---|---|
| `SKILL.md` | 识别主意图、选择领域、声明全局 RPC / 安全 / feedback |
| `<domain>/ENTRY.md` | 领域内路由和领域硬边界 |
| `<domain>/app/*.md` | 用户任务流程卡与交互门控；核心段固定为 `流程`、`边界条件`、`输出模板` |
| `<domain>/domains/**/*.md` | 稳定领域知识和决策表 |
| `<domain>/tools/` | 如存在：Python 工具、schema、长表与 RPC 编排；naming / fengshui 当前直接使用固定 discover scope 内声明的 RPC |

## 硬规则

1. 全包只能有一个 `SKILL.md`。
2. 根入口保持轻量，不写具体命理流程。
3. 每个领域必须有 `ENTRY.md`。
4. 所有文档路径从 `skills/liki` 根开始书写。
5. 有 `tools/` 的领域，LLM 只能通过 Python 工具层调用 RPC；无 `tools/` 的领域，LLM 只能使用 `ENTRY.md` / `RPC.md` 固定 discover scope 内声明的 RPC。
6. `VERSION.txt`、`FAQ.md`、`requirements.txt`、`feedback.py`、`feedback.schema.json` 不允许在领域内重复。
7. App 卡（`app/README.md` 除外）必须包含 `## 流程`、`## 边界条件`、`## 输出模板`；不得回退为 emoji 变体、`边界` 或带括号的专用变体。

## Release model

Liki uses dual versions:

- Runtime / compatibility contract: CalVer in `VERSION.txt`.
- Product release identity: SemVer Git tag, for example `v5.0.0`.

Do not merge these into one value. See [RELEASE_MODEL.md](./RELEASE_MODEL.md).

## Ready-to-use payload

`TOOLS.md` 和 `RPC.md` 是报文库，不是解释文档。app 卡应引用固定章节；动态对象（`pan`、`snapshot`、chart result）按“变量绑定”原样传回，不得裁剪、重建或猜字段。

## 打包

`make build-archive` 只生成：

```text
dist/liki.tar.gz
```

`dist/index.json` 只包含 `liki` 一个 entry。旧 `liki-*` 包不再发布。
