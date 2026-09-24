# Liki 八字专家包

> `liki-bazi` — 八字专家。排盘由 Go 天文历算引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-bazi/
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── agents/
│   └── bazi-expert.md       # Agent 定义（frontmatter + 系统提示词）
├── skills/
│   └── bazi/
│       ├── SKILL.md         # 专家人设 + 方法论入口
│       └── *.md             # 方法论卡（旺衰/用神/格局/十神/大运/合盘/考时等 16 卡）
├── .mcp.json                # 连接器：engine-bazi（排盘）+ counsel-bazi（判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-bazi", "counsel-bazi"]` — 依赖排盘连接器与判断连接器（排盘 engine-bazi，判断 counsel-bazi）。

## 校验要点（提交前核对）

- `expertType: "agent"`；`agentName` 与 `agents/*.md` 文件名一致。
- `tags` 固定 3 个；`quickPrompts` 固定 3 个；`quickPrompts[0]` 与 `defaultInitPrompt` 一致。
- `displayDescription.zh` 40-50 字。
- `categoryId` = `12-IndustryConsultant`。
- 头像 512×512 PNG、≤500KB、漫画/插画风格。
- 不添加任何系统 tools（权限由平台统一分配 + 连接器提供）。

## 待办

- [ ] `avatars/expert.png` 当前为占位（纯色+文字），上架前替换为正式漫画/插画风头像。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。