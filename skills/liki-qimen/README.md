# Liki 奇门专家包

**`liki-qimen`** — 奇门专家（含黄历择日子功能）。排盘由 Go 引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-qimen/
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── agents/
│   └── qimen-expert.md      # Agent 定义（frontmatter + 系统提示词）
├── skills/
│   ├── qimen/
│   │   ├── SKILL.md         # 专家人设 + 方法论入口
│   │   └── *.md             # 方法论卡（八门/八神/九星/格局/专占 等 11 卡）
│   └── huangli/
│       └── *.md             # 黄历择日卡（2 卡，奇门子功能）
├── .mcp.json                # 连接器：engine-mcp（排盘+黄历）+ counsel-mcp（判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp", "counsel-mcp"]` — 排盘/黄历走 engine-mcp（qimen_chart/huangli_days），断语走 counsel-mcp（qimen_snapshot/qimen_query）。

## 说明

- 黄历择日（`huangli_days`）是简单 aux 工具，LLM 自举使用；择日卡就近放在本专家，与奇门择吉连用。
- 校验要点与 `liki-bazi` 一致（expertType/agentName/tags/quickPrompts/displayDescription/categoryId/头像）。

## 待办

- [ ] `avatars/expert.png` 当前为占位，上架前替换为正式漫画/插画风头像。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。