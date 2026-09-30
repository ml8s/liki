# Liki 风水专家包

**`liki-fengshui`** — 风水专家（八宅 + 玄空两流派）。排盘由 Go 引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-fengshui/
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── agents/
│   └── fengshui-expert.md   # Agent 定义（frontmatter + 系统提示词）
├── skills/
│   └── fengshui/
│       ├── SKILL.md         # 专家人设 + 方法论入口
│       └── references/
│           └── *.md             # 方法论卡（八宅游星/门主灶、玄空飞星/元运、流派裁决 等 5 卡）
├── .mcp.json                # 连接器：engine-mcp（八宅+玄空排盘）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp"]` — 八宅/玄空排盘走 engine-mcp（bazhai_chart/bazhai_layout/xuankong_chart/xuankong_liunian），无独立判断层。

## 说明

- **八宅与玄空是同一风水的两个流派**（《八宅明镜》《沈氏玄空》），合并在本专家内跨流派裁决，不拆成两个专家。
- 校验要点与 `liki-bazi` 一致。

## 待办

- [ ] `avatars/expert.png` 当前为占位，上架前替换为正式漫画/插画风头像。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。