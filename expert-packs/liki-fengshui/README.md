# Liki 风水专家包

> **兼容提示**：本目录是专家的 CodeBuddy/WorkBuddy 市场发布形态；包根 `SKILL.md` 为身份唯一
> 手写源，`skills/fengshui/`（域 skill 与 references）由身份与方法论卡生成、逐字节一致，勿手改。

**`liki-fengshui`** — **Liki 风水专家**（八宅 + 玄空两流派）。排盘由确定性引擎计算，断语来自规则真值表，依据可回溯。

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
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引
│       └── references/      # 生成：方法论文卡镜像（逐字节一致）
│           ├── ENTRY.md     # 域内路由
│           ├── app/         # 场景卡（家居风水，含八宅 / 玄空合参与裁决）
│           └── domains/
│               ├── bazhai/  # 八宅方法论卡（游星 / 门主灶）
│               └── xuankong/ # 玄空方法论卡（飞星 / 元运）
└── README.md
```

## 依赖

- 运行依赖是对已上架连接器的声明（`plugin.json.dependencies.connectors`：排盘引擎 + 断语引擎，
  ID 单一源 `workbuddy/connector-ids.json`），包内不自带 MCP。本专家只用排盘引擎的工具，
  断语引擎工具不参与，故规则卡须写明结论由本域卡给出。

## 说明

- **八宅与玄空是同一风水的两个流派**（《八宅明镜》《沈氏玄空》），合并在本专家内跨流派裁决，不拆成两个专家；
  领域卡只写本流派方法，跨流派裁决落在 `references/app/fengshui.md`「八宅 / 玄空合参与裁决」节。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
