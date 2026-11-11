# WorkBuddy 提交

付费版 Liki 专家包通过 SkillHub.cn 分发；MCP 端点为 liki-mcp（SkillPay 付费网关）。

## 结构

| 路径 | 内容 |
| --- | --- |
| `.mcp.json` | 指向 liki-mcp 付费网关（`https://mcp.tokflux.com/mcp`） |
| `SKILL.md` | 方法论卡索引（生成物） |
| `expert-packs/liki/agents/liki.md` | 部署身份（从身份源投影） |
| `avatars/expert.svg` / `expert.png` | 品牌资源 |
| `.codebuddy-plugin/plugin.json` | 平台插件元数据 |

免费版 `skills/liki/` 直连 ENGINE + COUNSEL，不经过 liki-mcp。

## 打包

```sh
make workbuddy-package
make check
```

`liki-expert.zip` 上传 SkillHub.cn。
