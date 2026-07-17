# AGENTS.md — liki.hk 开发规则

这些规则适用于本仓库的所有 AI agent 和人类开发者。

## 架构边界

- **Skill 提示词 ⇄ 引擎分离**：`SKILL.md` 不写方法名、参数列表（由 `rpc.discover` 的 schema 驱动）。引擎升级不影响 Skill，Skill 优化不依赖引擎发版。
- **引擎是私有仓库**：`ml8s/liki-engine` 不在本仓。本仓只管理提示词、报告模板、参考文件、文档。
- **网站是独立仓库**：`ml8s/liki-web` 负责 HTML/部署/CI。`web/skills/` 下的文件由 `sync-skills.sh` 从本仓同步，不要在那边手动改。

## 铁律

1. **禁手算**：所有排盘结果必须通过引擎 RPC 获取。禁止 AI 凭训练知识编造四柱、用神、大运、神煞等数据。
2. **结论先行**：所有分析输出第一句必须给明确判断，不得以"可能/或许/有几种可能"开头。
3. **不写方法名**：子 SKILL.md 描述工作流步骤，不列方法名和参数。方法 schema 由 `rpc.discover` 自动发现。
4. **体系边界不可混**：风水不用八字用神解释，紫微不串八字十神，八宅命卦和八字日主是两个独立体系。
5. **品牌统一**：品牌用 `liki.hk`，对话中用 `liki（灵机）`作为友好称呼。页面标题、meta 描述、对外文案一律用 `liki.hk`，不用"灵机平台""Skills 合集""子包"等过时表述。

## 升级版本

每次修改必须同步四处：

1. `VERSION` 文件 ← 版本号
2. `SKILL.md` ← `version:` 字段
3. `CHANGELOG.md` ← 新增条目
4. `README.md` + `README.en.md` ← 版本号 + 如有品牌相关也同步更新

版本号规则：bug 修→patch，流程变→minor，架构变→major。

## 文件结构与编码

- SKILL.md 中文为主，术语保持原文，不写方法名和参数
- 报告模板放在子目录下（如 `mingli/report-chart.md`），字段只引用 RPC schema 中存在的
- `references/` 目录存放经典摘要参考文件，不打 `.tar.gz`，随包分发
- 同步到 `web/skills/` 时排除 `.git`、`.git-*`、`README.md`、`*.tar.gz`

## 身份与品牌规范

完整规范见 [`docs/Identity.zh-Hans.md`](./docs/Identity.zh-Hans.md)。
快速索引：

- **品牌**：`liki.hk`
- **品类**：AI命理Skills
- **一句话**：liki.hk 提供 AI命理Skills，覆盖八字、紫微、起名、问卦、风水。
- 不用词：平台、合集、子包、套件、Suite、collection
- 对话内友好称呼：`liki（灵机）`
