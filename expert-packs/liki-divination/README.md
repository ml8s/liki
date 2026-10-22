# Liki 占卜专家包

> **兼容提示**：本目录是专家的 CodeBuddy/WorkBuddy 市场发布形态；包根 `SKILL.md` 为身份唯一
> 手写源，`skills/divination/`（域 skill 与 references）由身份与方法论卡生成、逐字节一致，勿手改。

**`liki-divination`** — **Liki 占卜专家**（六爻 + 奇门 + 黄历择日）。起卦、排局与择日事实由确定性工具计算，断语来自规则真值表，依据可回溯。

## 结构

```text
liki-divination/
├── SKILL.md                 # 手写：专家身份唯一源（元数据 + 角色正文）
├── .codebuddy-plugin/
│   └── plugin.json          # 专家核心配置与市场展示
├── avatars/
│   └── expert.png           # 头像：512×512 PNG ≤500KB，漫画/插画风格
├── skills/
│   └── divination/
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引
│       └── references/              # 生成：本域平铺卡（按 `divination-` 前缀过滤的逐字节视图）
│           ├── divination-ENTRY.md  # 域内路由（锁定所问 → 方法卡）
│           ├── divination-TOOLS.md  # 工具链、变量绑定与输出契约
│           ├── divination-app-*.md  # 场景卡（锁定问题 / 问结果 / 问策略 / 问日期）
│           ├── divination-liuyao-*.md  # 六爻方法论卡（用神/吉凶/应期/月建/六兽等）
│           ├── divination-qimen-*.md   # 奇门方法论卡（八门/九星/八神/格局/专占等）
│           └── divination-huangli-*.md # 黄历择日卡
└── README.md
```

## 依赖

- 运行依赖是对已上架连接器的声明（`plugin.json.dependencies.connectors`：排盘引擎 + 断语引擎，
  ID 单一源 `workbuddy/connector-ids.json`），包内不自带 MCP；六爻/奇门/黄历的工具都在
  这两个聚合连接器内，引导卡片与其余三个专家共用同一次连接。

## 说明

- **一个专家覆盖三种占法**：六爻、奇门与黄历择日同属「占卜」——都回答「这一件事怎么样」；
  黄历只输出日期适配事实，与奇门择吉连用，不拆成独立专家。
- 领域卡只写本体系方法（`divination-liuyao-*` 不写奇门规则，`divination-qimen-*` 不写六爻规则）；
  锁定所问、跨方法选择与输出契约落在 `divination-app-*` 与 `divination-TOOLS.md`。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
