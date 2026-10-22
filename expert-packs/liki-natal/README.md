# Liki 命理专家包

> **兼容提示**：本目录是专家的 CodeBuddy/WorkBuddy 市场发布形态；包根 `SKILL.md` 为身份唯一
> 手写源，`skills/natal/`（域 skill 与 references）由身份与方法论卡生成、逐字节一致，勿手改。

**`liki-natal`** — **Liki 命理专家**（八字 + 紫微两体系）。排盘与断语来自确定性工具，依据可回溯。

## 结构

```text
liki-natal/
├── SKILL.md                 # 手写：专家身份唯一源（元数据 + 角色正文）
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── skills/
│   └── natal/
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引
│       └── references/          # 生成：本域平铺卡（按 `natal-` 前缀过滤的逐字节视图）
│           ├── natal-ENTRY.md   # 域内路由（场景卡 + 按侧执行）
│           ├── natal-TOOLS.md   # 工具链、断语输出契约与「合参与裁决」
│           ├── natal-app-*.md   # 场景卡（命书 / 婚姻 / 事业 / 财运 / 健康 / 学业 / 性格 / 六亲 / 合盘）
│           ├── natal-bazi-*.md  # 八字方法论卡
│           └── natal-ziwei-*.md # 紫微方法论卡
└── README.md
```

## 依赖

- 运行依赖是对已上架连接器的声明（`plugin.json.dependencies.connectors`：排盘引擎 + 断语引擎，
  ID 单一源 `workbuddy/connector-ids.json`），包内不自带 MCP；引导卡片由连接器提供，
  与其余三个专家共用同一次连接。

## 说明

- **一个专家覆盖两体系**：八字与紫微同属「本命」，双侧各自排盘出断语，合参与裁决（含合盘、考时）
  在包内 `references/natal-TOOLS.md`「合参与裁决」与各场景卡的「双盘合参」节完成，不拆成两个专家。
- 领域卡只写本体系方法（`natal-bazi-*` 不写紫微规则，`natal-ziwei-*` 不写八字规则）；
  跨体系内容一律落在 `natal-app-*` 场景卡与 `natal-TOOLS.md`。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
