# Liki 本命专家包

> **兼容提示**：生产系统只发布全功能 `liki-experts` 装配镜像（部署配置在仓库根
> `agents/`）。本目录的 `SKILL.md` 是本专家身份的唯一手写源（内部部署与市场包共用），
> 同时作为 CodeBuddy 市场包兼容层。

**`liki-natal`** — 本命专家（八字 + 紫微两体系）。排盘与断语来自确定性工具，依据可回溯。

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
│       ├── SKILL.md         # 生成：身份正文 + 方法论卡索引（ADK skill 入口）
│       └── references/      # 生成：根 references/natal/ 子树的逐字节镜像
│           ├── ENTRY.md     # 域内路由（场景卡 + 按侧执行）
│           ├── TOOLS.md     # 工具链、断语输出契约与「合参与裁决」
│           ├── app/         # 场景卡（命书 / 婚姻 / 事业 / 财运 / 健康 / 学业 / 性格 / 六亲 / 合盘）
│           └── domains/
│               ├── bazi/    # 八字方法论卡（旺衰/用神/格局/十神/大运/合盘/考时等）
│               └── ziwei/   # 紫微方法论卡（主星/格局/十二宫/性格/应期/流年等）
├── .mcp.json                # 连接器：engine-bazi + engine-ziwei（排盘）+ counsel-bazi + counsel-ziwei（判断）
└── README.md
```

## 依赖

- `dependencies.connectors: ["engine-bazi", "engine-ziwei", "counsel-bazi", "counsel-ziwei"]` —
  排盘走 `engine-bazi` / `engine-ziwei`，判断走 `counsel-bazi` / `counsel-ziwei`。

## 说明

- **一个专家覆盖两体系**：八字与紫微同属「本命」，双侧各自排盘出断语，合参与裁决（含合盘、考时）
  在包内 `references/TOOLS.md`「合参与裁决」与各场景卡的「双盘合参」节完成，不拆成两个专家。
- 领域卡只写本体系方法（`domains/bazi/` 不写紫微规则，`domains/ziwei/` 不写八字规则）；
  跨体系内容一律落在 `app/` 场景卡与 `TOOLS.md`。
- 校验要点与上架前待办见 `expert-packs/README.md`（expertType / agentName / tags / quickPrompts / displayDescription / categoryId / 头像）。
