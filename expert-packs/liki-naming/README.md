# Liki 起名专家包

> **兼容提示**：生产系统只发布全功能 `liki-experts` 装配镜像（部署配置在仓库根
> `agents/`）。本目录的 `SKILL.md` 是本专家身份的唯一手写源（内部部署与市场包共用），
> 同时作为 CodeBuddy 市场包兼容层。

**`liki-naming`** — 起名专家。用神与字库属性由 Go 引擎确定性计算，LLM 只在引擎返回范围内做语义/音韵/文化筛选，依据可回溯。

## 结构

```text
liki-naming/
├── SKILL.md                 # 手写：专家身份唯一源（元数据 + 角色正文）
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── skills/
│   └── naming/
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引（ADK skill 入口）
│       └── references/
│           └── *.md             # 方法论卡（起名流程/外国人/自选评估/用神取字/字库 等 7 卡）
├── .mcp.json                # 连接器：engine-mcp（八字排盘）+ counsel-mcp（起名判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp", "counsel-mcp"]` — 八字排盘能力由 engine-mcp 提供；姓氏匹配、取字、组名和评估能力由 counsel-mcp 提供。

## 说明

- 起名依赖八字用神（`domains/bazi` 的 yongshen/calibration），但产出是名字，属独立起名术数，独立成专家。
- 校验要点与 `liki-bazi` 一致。

## 待办

- [ ] `avatars/expert.png` 当前为占位（纯色+文字）；上架前替换为正式漫画/插画风头像（素材授权要求见 `expert-packs/README.md`）。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。