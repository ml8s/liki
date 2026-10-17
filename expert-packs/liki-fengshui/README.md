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
│       └── references/      # 生成：根 references/fengshui/ 子树的逐字节镜像
│           ├── ENTRY.md     # 域内路由
│           ├── app/         # 场景卡（家居风水，含八宅 / 玄空合参与裁决）
│           └── domains/
│               ├── bazhai/  # 八宅方法论卡（游星 / 门主灶）
│               └── xuankong/ # 玄空方法论卡（飞星 / 元运）
└── README.md
```

## 依赖

- 运行依赖是对已上架连接器的声明（`plugin.json.dependencies.connectors`：排盘引擎 + 判断层，
  ID 单一源 `workbuddy/connector-ids.json`），包内不自带 MCP。本专家只用排盘引擎的工具，
  判断层工具不参与，故规则卡须写明结论由本域卡给出。

## 说明

- **八宅与玄空是同一风水的两个流派**（《八宅明镜》《沈氏玄空》），合并在本专家内跨流派裁决，不拆成两个专家；
  领域卡只写本流派方法，跨流派裁决落在 `references/app/fengshui.md`「八宅 / 玄空合参与裁决」节。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
