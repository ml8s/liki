# Liki 起名专家包

> **兼容提示**：本目录是专家的 CodeBuddy/WorkBuddy 市场发布形态；包根 `SKILL.md` 为身份唯一
> 手写源，`skills/naming/`（域 skill 与 references）由身份与方法论卡生成、逐字节一致，勿手改。

**`liki-naming`** — **Liki 起名专家**。用神来自八字裁决，字库、音韵与五行匹配在本地确定性完成，取名在其返回范围内做语义与文化筛选，依据可回溯。

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
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引
│       └── references/            # 生成：本域平铺卡（按 `naming-` 前缀过滤的逐字节视图）
│           ├── naming-ENTRY.md    # 域内路由
│           ├── naming-app-*.md    # 场景卡（通用起名 / 外国人起名 / 自选名评估）
│           ├── naming-bazi-*.md   # 起名用神与考时精简卡
│           └── naming-qiming-*.md # 字库与选字卡
└── README.md
```

## 依赖

- 运行依赖是对已上架连接器的声明（`plugin.json.dependencies.connectors`：排盘引擎 + 断语引擎，
  ID 单一源 `workbuddy/connector-ids.json`），包内不自带 MCP；取用神用排盘引擎、取字与评估
  用断语引擎，引导卡片与其余三个专家共用同一次连接。

## 说明

- 起名依赖八字用神（`naming-bazi-*` 的 yongshen/calibration），但产出是名字，属独立起名术数，独立成专家。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
