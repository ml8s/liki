# Unified Skill 包结构

Liki 对外只有一个可安装 skill：`liki`。它通过标准 MCP 提供能力，是纯指令 + 支付/反馈的客户端。

## 目录契约

```text
skills/liki/
├── SKILL.md              # 唯一 skill 入口（MCP 调用 + 流程 + 边界）
├── VERSION.txt           # 唯一分发版本
├── FAQ.md                # 唯一运行失败 / 反模式恢复入口
├── .mcp.json             # MCP 依赖声明（liki-analysis + liki-engine）
├── requirements.txt      # Python 依赖清单
├── aipay.py              # 支付履约
├── feedback.py           # 唯一 feedback sender
└── feedback.schema.json  # feedback-v1 contract
```

规则层（因子/断语/应期/考时）与排盘引擎在 `liki-core` 仓库，skill 分发包不含领域目录。

## 分层

| 层 | 职责 |
|---|---|
| `SKILL.md` | MCP 前置、领域路由到 MCP 工具、硬边界、Aipay、输出契约、Feedback |
| `.mcp.json` | 声明 `liki-analysis` + `liki-engine` 端点（客户端启用时自动连接） |
| `aipay.py` / `feedback.py` | 支付履约与反馈 |

## 硬规则

1. 全包只能有一个 `SKILL.md`。
2. 根入口保持轻量，不写具体命理流程（流程由 MCP 工具 + SKILL.md 描述承载）。
3. 所有能力通过 MCP 工具调用：`liki-analysis`（判断层）+ `liki-engine`（排盘/起名/风水）。
4. MCP 不可用时明确说明降级，不自行用模型记忆补结果。
5. `VERSION.txt`、`FAQ.md`、`requirements.txt`、`aipay.py`、`feedback.py`、`feedback.schema.json` 不允许重复。
6. 规则层（因子/断语/应期/考时）在 `liki-core` 仓库，skill 不携带。

## Release model

Liki uses dual versions:

- Runtime / compatibility contract: CalVer in `VERSION.txt`; updated by normal engineering bumps.
- Product release identity: SemVer Git tag, for example `v5.0.0`; updated only on an intentional release.
- SkillHub package: `SKILL.md` frontmatter `version`; mirror the SemVer only when preparing a SkillHub release.

Do not merge these into one value. See [RELEASE_MODEL.md](./RELEASE_MODEL.md).

## 打包

`make build-archive` 只生成：

```text
dist/liki.tar.gz
```

`dist/index.json` 只包含 `liki` 一个 entry。旧 `liki-*` 包不再发布。