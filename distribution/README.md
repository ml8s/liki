# Liki distribution

`distribution/` 只保存分发适配信息，不保存第二份方法论或专家身份。

## 目录

```text
liki-expert/
  profile.json         # 内部 AgentGo 发布 profile
  deployment.yaml      # 内部 deployment 配置与工具白名单
  host-adapter.md      # 仅内部宿主使用；公开渠道包不得包含

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
- 内部专家装配源：`liki-expert/profile.json`、`liki-expert/deployment.yaml`
- 内部宿主协议：`liki-expert/host-adapter.md`，只在 expert 镜像 instruction 中合并
- 渠道目录只保存 SKU、MCP、平台差异元数据和渠道 overlay。
- 市场展示字段（display_name / description_zh / examples / agent_name）不写进源 `skills/liki/SKILL.md`，由 `distribution/metadata/common.json`、渠道差异文件和 WorkBuddy `skill-meta.json` 注入。
- `author` / `license` 不重复配置，构建时直接从源 `SKILL.md` frontmatter 派生。
- 渠道包只生成各自平台的字段方言：SkillHub 带 `slug` / `displayName` / `summary`，WorkBuddy 带 `agent_created`；两者都不再输出重复的嵌套 `metadata`。

## 构建

```bash
make distribution-package
```

包版本由当前 commit 派生为 host 兼容 semver（非 CalVer），可通过 `LIKI_PACKAGE_VERSION` 显式覆盖。产物：

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
make distribution-check
```

会校验：

- zip 内容与 variant / channel 配置一致；
- free 包不含支付流程；
- pro 包不包含 `is_calibrated` / `sync_ui`；
- 方法论卡没有第二份源码镜像。

## Expert assembly

`assembly/Dockerfile` 使用显式指定的 AgentGo base image，并把生成的
`deployment.json`、instruction 和 `skills/` 复制进 `liki-expert` 镜像：

```bash
make build-deployment
make image-assembly AGENTGO_IMAGE=ghcr.io/ml8s/agentgo:v1.0.0
```

`assembly/` 与 `distribution/liki-expert/` 只用于专家镜像；它们不会进入
SkillHub 或 WorkBuddy 分发包。
