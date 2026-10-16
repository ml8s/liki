# Liki 占卜专家包

> **兼容提示**：生产系统只发布全功能 `liki-experts` 装配镜像（部署配置在仓库根
> `agents/`）。本目录的 `SKILL.md` 是本专家身份的唯一手写源（内部部署与市场包共用），
> 同时作为 CodeBuddy 市场包兼容层。

**`liki-divination`** — 占卜专家（六爻 + 奇门 + 黄历择日）。起卦、排局与择日事实由确定性工具计算，断语来自规则真值表，依据可回溯。

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
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引（ADK skill 入口）
│       └── references/      # 生成：根 references/divination/ 子树的逐字节镜像
│           ├── ENTRY.md     # 域内路由（锁定所问 → 方法卡）
│           ├── TOOLS.md     # 工具链、变量绑定与输出契约
│           ├── app/         # 场景卡（锁定问题 / 问结果 / 问策略 / 问日期）
│           └── domains/
│               ├── liuyao/  # 六爻方法论卡（用神/吉凶/应期/月建/六兽等）
│               ├── qimen/   # 奇门方法论卡（八门/九星/八神/格局/专占等）
│               └── huangli/ # 黄历择日卡
├── .mcp.json                # 连接器：engine-liuyao + engine-qimen + engine-huangli（排盘/历法）+ counsel-liuyao + counsel-qimen（判断）
└── README.md
```

## 依赖

- 运行依赖由包内 `.mcp.json` 声明（`dependencies.mcpServers: "./.mcp.json"`）：六爻起卦、奇门排局与
  黄历查日由 `engine-liuyao` / `engine-qimen` / `engine-huangli` 提供，不可变快照与追问能力由
  `counsel-liuyao` / `counsel-qimen` 提供。

## 说明

- **一个专家覆盖三种占法**：六爻、奇门与黄历择日同属「占卜」——都回答「这一件事怎么样」；
  黄历只输出日期适配事实，与奇门择吉连用，不拆成独立专家。
- 领域卡只写本体系方法（`domains/liuyao/` 不写奇门规则，`domains/qimen/` 不写六爻规则）；
  锁定所问、跨方法选择与输出契约落在 `app/` 与 `TOOLS.md`。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
