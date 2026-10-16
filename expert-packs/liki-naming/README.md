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
│       └── references/      # 生成：根 references/naming/ 子树的逐字节镜像
│           ├── ENTRY.md     # 域内路由
│           ├── app/         # 场景卡（通用起名 / 外国人起名 / 自选名评估）
│           └── domains/
│               ├── bazi/    # 起名用神与考时精简卡
│               └── qiming/  # 字库与选字卡
├── .mcp.json                # 连接器：engine-mcp（八字排盘）+ counsel-mcp（起名判断）
└── README.md
```

## 依赖

- 运行依赖由包内 `.mcp.json` 声明（`dependencies.mcpServers: "./.mcp.json"`）：八字排盘能力由 `engine-bazi` 提供（取用神）；姓氏匹配、取字、组名和评估能力由 `counsel-naming` 提供。

## 说明

- 起名依赖八字用神（`domains/bazi` 的 yongshen/calibration），但产出是名字，属独立起名术数，独立成专家。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
