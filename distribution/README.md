# Liki distribution

`distribution/` 只保存分发适配信息，不保存第二份方法论或专家身份。

## 目录

```text
liki-expert/
  profile.json         # 内部 AgentRuntime profile
  deployment.yaml      # 内部 agent 配置
  host-adapter.md      # 仅内部部署使用；公开包不得包含

variants/
  free.json            # 直连 engine + counsel
  pro.json             # 通过 liki-mcp 付费网关访问

channels/
  skillhub.json        # SkillHub 免费包 / pro 包规则
  workbuddy.json       # WorkBuddy 免费包 / pro expert 包规则

metadata/
  common.json          # 跨渠道共享描述、示例与 Agent 名
  skillhub.json        # SkillHub 差异字段：slug、展示名

workbuddy/
  liki-free/           # WorkBuddy 免费专家静态资产与 skill-meta.json
  liki-pro/            # WorkBuddy pro 静态资产：plugin 元数据、支付 overlay、头像、README、skill-meta.json
```

## 权威源

- 唯一专家身份和公共 instruction：[`skills/liki/SKILL.md`](../skills/liki/SKILL.md)
- 唯一方法论源：[`skills/liki/references/`](../skills/liki/references/)
- 渠道目录只保存 SKU、MCP、平台差异元数据和渠道 overlay。
- 市场展示字段（display_name / description_zh / examples / agent_name）不写进源 `skills/liki/SKILL.md`，由 `distribution/metadata/common.json`、渠道差异文件和 WorkBuddy `skill-meta.json` 注入。
- `author` / `license` 不重复配置，构建时直接从源 `SKILL.md` frontmatter 派生。
- 渠道包只生成各自平台的字段方言：SkillHub 带 `slug` / `displayName` / `summary`，WorkBuddy 带 `agent_created`；两者都不再输出重复的嵌套 `metadata`。

## 构建

```bash
make distribution-package
```

版本由 Makefile 的 `LIKI_PACKAGE_VERSION` 固定；每次对外发布前递增。产物：

```text
dist/workbuddy/liki-free.zip
dist/workbuddy/liki-pro.zip
dist/skillhub/liki-free.zip
dist/skillhub/liki-pro.zip
```

`liki-free` 直连 `engine-mcp` 与 `counsel-mcp`。
`liki-pro` 走 `liki-mcp` 付费网关，并包含 SkillPay 支付处理说明。

## 校验

```bash
make check-content
```

会校验：

- zip 内容与 variant / channel 配置一致；
- free 包不含支付流程；
- pro 包不包含 `is_calibrated` / `sync_ui`；
- 方法论卡没有第二份源码镜像。
