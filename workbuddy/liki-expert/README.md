# Liki 专家包（WorkBuddy）

> `liki-master` — 命理顾问专家。排盘由 Go 天文历算引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-expert/
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── agents/
│   └── liki-master.md       # Agent 定义（frontmatter + 系统提示词）
├── skills/
│   └── liki-usage/
│       └── SKILL.md         # 预加载技能：连接器工具使用手册
└── README.md
```

## 依赖

- `dependencies.connectors: ["liki"]` — 依赖已上架的 **Liki 连接器**。用户召唤本专家前，WorkBuddy 会先引导连接 Liki 连接器。
- 上架顺序：先提交并审核 Liki 连接器，再提交本专家。

## 校验要点（提交前核对）

- `expertType: "agent"`；`agentName` 与 `agents/*.md` 文件名一致。
- `tags` 固定 3 个；`quickPrompts` 固定 3 个；`quickPrompts[0]` 与 `defaultInitPrompt` 一致。
- `displayDescription.zh` 40-50 字。
- `categoryId` = `12-IndustryConsultant`。
- 头像 512×512 PNG、≤500KB、漫画/插画风格。
- 不添加任何系统 tools（权限由平台统一分配 + 连接器提供）。

## 待办

- [ ] 生成 `avatars/expert.png`（当前缺失，需 AI 绘图或设计稿）。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。
