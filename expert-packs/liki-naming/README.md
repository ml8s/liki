# Liki 起名专家包

**`liki-naming`** — 起名专家。用神与字库属性由 Go 引擎确定性计算，LLM 只在引擎返回范围内做语义/音韵/文化筛选，依据可回溯。

## 结构

```text
liki-naming/
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── agents/
│   └── naming-expert.md     # Agent 定义（frontmatter + 系统提示词）
├── skills/
│   └── naming/
│       ├── SKILL.md         # 专家人设 + 方法论入口
│       └── references/
│           └── *.md             # 方法论卡（起名流程/外国人/自选评估/用神取字/字库 等 7 卡）
├── .mcp.json                # 连接器：engine-mcp（八字排盘）+ counsel-mcp（起名判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp", "counsel-mcp"]` — 用神取字走 engine-mcp（bazi_chart/fullchart），起名走 counsel-mcp（qiming_surname/pick/char/compose/check）。

## 说明

- 起名依赖八字用神（`domains/bazi` 的 yongshen/calibration），但产出是名字，属独立起名术数，独立成专家。
- 校验要点与 `liki-bazi` 一致。

## 待办

- [ ] `avatars/expert.png` 当前为占位，上架前替换为正式漫画/插画风头像。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。