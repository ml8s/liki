# Liki 风水专家包

> **兼容提示**：生产系统只发布全功能 `liki-experts` 装配镜像（部署配置在仓库根
> `agents/`）。本目录的 `SKILL.md` 是本专家身份的唯一手写源（内部部署与市场包共用），
> 同时作为 CodeBuddy 市场包兼容层。

**`liki-fengshui`** — 风水专家（八宅 + 玄空两流派）。排盘由 Go 引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-fengshui/
├── SKILL.md                 # 手写：专家身份唯一源（元数据 + 角色正文）
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── skills/
│   └── fengshui/
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引（ADK skill 入口）
│       └── references/
│           └── *.md             # 方法论卡（八宅游星/门主灶、玄空飞星/元运、流派裁决 等 5 卡）
├── .mcp.json                # 连接器：engine-mcp（八宅+玄空排盘）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp"]` — 八宅排盘与布局、玄空排盘与流年能力由 engine-mcp 提供，无独立判断层。

## 说明

- **八宅与玄空是同一风水的两个流派**（《八宅明镜》《沈氏玄空》），合并在本专家内跨流派裁决，不拆成两个专家。
- 校验要点与 `liki-bazi` 一致。

## 待办

- [ ] `avatars/expert.png` 当前为占位（纯色+文字）；上架前替换为正式漫画/插画风头像（素材授权要求见 `expert-packs/README.md`）。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。