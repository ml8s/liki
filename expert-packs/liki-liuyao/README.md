# Liki 六爻专家包

> **兼容提示**：生产系统只发布全功能 `liki-experts` 装配镜像（部署配置在仓库根
> `agents/`）。本目录的 `SKILL.md` 是本专家身份的唯一手写源（内部部署与市场包共用），
> 同时作为 CodeBuddy 市场包兼容层。

**`liki-liuyao`** — 六爻专家。起卦由 Go 引擎确定性计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-liuyao/
├── SKILL.md                 # 手写：专家身份唯一源（元数据 + 角色正文）
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── skills/
│   └── liuyao/
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引（ADK skill 入口）
│       └── references/
│           └── *.md             # 方法论卡（吉凶/应期/用神/六亲/模式 等 6 卡）
├── .mcp.json                # 连接器：engine-mcp（起卦）+ counsel-mcp（判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-mcp", "counsel-mcp"]` — 起卦与装卦能力由 engine-mcp 提供；不可变快照与追问能力由 counsel-mcp 提供。

## 校验要点（提交前核对）

- `expertType: "agent"`；`agentName` 与 `agents/*.md` 文件名一致。
- `tags` 固定 3 个；`quickPrompts` 固定 3 个；`quickPrompts[0]` 与 `defaultInitPrompt` 一致。
- `displayDescription.zh` 40-50 字。
- `categoryId` = `12-IndustryConsultant`。
- 头像 512×512 PNG、≤500KB、漫画/插画风格。
- 不添加任何系统 tools（权限由平台统一分配 + 连接器提供）。

## 待办

- [ ] `avatars/expert.png` 当前为占位（纯色+文字）；上架前替换为正式漫画/插画风头像（素材授权要求见 `expert-packs/README.md`）。
- [ ] 用官方 `expert-manager` 技能逐项校验后上架。