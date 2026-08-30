<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-命理_Skill-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki 灵机 — 命理 Skill</strong><br>
  按命理师的专业标准构建：排盘走天文历算引擎，断语附经典出处，结论可验证<br>
  八字 · 紫微 · 六爻 · 奇门 · 黄历择日 · 风水 · 起名
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-官方网站-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img src="https://img.shields.io/badge/CI-全绿-4a9e6b?style=flat&logo=githubactions&logoColor=white&labelColor=30305c"></a>
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
</p>

---

## 30 秒了解

安装后，你的 AI 助手获得 4 个命理技能：

| 技能 | 能问什么 | 试试这样问 |
|------|---------|-----------|
| **liki-bazi** 命理 | 婚姻、事业、财运、健康、学业、性格、六亲、两人合盘、全盘命书 | `算八字，1990-05-20 12:00 北京出生，男` |
| **liki-naming** 起名 | 新生儿起名、成人改名、公司命名、英文起中文名、名字评估 | `宝宝起名，2024-06-10 广州出生，男，姓陈` |
| **liki-divination** 问卦 | 六爻问成败应期、奇门问方向时机、黄历选吉日 | `这件事能成吗？什么时候有结果` |
| **liki-fengshui** 风水 | 八宅命卦布局、玄空飞星、流年风水 | `我家风水怎么样` |

**专业标准意味着什么**：

- 排盘由天文历算引擎计算（真太阳时/节气秒级精度），AI 不编数字
- 判断依据 701 条断语真值表，每条附《渊海子平》《子平真诠》等经典原文
- 160 道命理师大赛真题独立评测，答案与评测过程隔离

## 安装

```bash
npx skills add ml8s/liki          # 一次安装全部 4 个技能
```

只装某一个：

```bash
npx skills add ml8s/liki --skill liki-bazi      # 命理（八字+紫微）
npx skills add ml8s/liki --skill liki-naming    # 起名
npx skills add ml8s/liki --skill liki-divination # 问卦
npx skills add ml8s/liki --skill liki-fengshui  # 风水
```

**装完后，直接这样开始**：

```
帮我出一份命书，1990-05-20 12:00 北京出生，男
我和她合不合？我 1992-03-15 生，她 1994-08-20 生
2026 年我的事业和财运怎么样？
```

## 使用手册

### 第一次使用

**准备**：出生日期（公历）、出生时间（尽量精确到分）、出生城市、性别。

**直接发**——出生信息和问题写在一条消息里即可：

> 帮我看看婚姻，女，1992-03-15 14:30 广州出生

信息不全也没关系：只知道"上午"或"不知道几点"，技能会追问或进入考时流程（见下方 FAQ）。

### 对话方式

技能像命理师一样与你对话——**一次问一个主题**，得到结构化分析后可以继续追问：

| 你想… | 这样说 |
|------|-------|
| 追问依据 | `为什么？` `这个结论的依据是什么？` |
| 问具体年份 | `2026年呢？` `未来三年怎么样？` |
| 换主题 | `那财运呢？` `健康方面呢？`（同一命盘，不用重新排） |
| 问对方 | `我们俩合不合？`（提供双方出生信息） |
| 要完整报告 | `帮我出一份命书`（全盘：婚姻/事业/财运/健康/学业/六亲） |

### 各技能详细指南

#### liki-bazi 命理（八字 + 紫微双盘同参）

按人生领域直接问——技能自动排盘、查断语、给出结论+依据+应期：

- **婚姻**：什么时候结婚？会不会离婚？对方什么样的人？
- **事业**：适合什么行业？创业还是打工？起伏在哪几年？
- **财运**：财源类型？哪年得财、哪年破财？
- **健康**：哪个脏腑薄弱？哪些年份要注意？
- **学业**：能读到什么学历？考试运如何？
- **性格 / 六亲**：我是什么性格？父母/子女缘分如何？

**输出格式**：结论先行，后附依据。每个结论可追溯到具体命理步骤和经典原文。

#### liki-naming 起名

> 宝宝起名，2024-06-10 广州出生，男，姓陈

流程：八字定用神 → 补五行 → 三才五格筛选 → 推荐名（附出处与依据）。

也支持：改名、公司命名、英文起中文名、自选名字评估。

#### liki-divination 问卦

问具体一件事的吉凶成败和时间——**越具体越准**：

- **六爻**：`这个项目能成吗？` `什么时候有结果？`
- **奇门**：`往哪个方向发展好？` `现在该不该签？`
- **黄历**：`哪天搬家/签约/开业好？`

输出：卦象依据 → 一句话判断 → 应期。

#### liki-fengshui 风水

- **八宅**：`我的命卦是什么？` `大门/厨房/卧室怎么布局？`
- **玄空**：`现在的元运我家旺不旺？` `2026年流年风水注意什么？`

### 常见问题

**不知道出生时辰怎么办？**
提供 2-3 个候选时辰 + 3-5 件人生大事及年份，技能会逐盘核验反推最可能的时辰（附置信度）。若证据不足，会明确要求补充时辰范围或验证事件，不会自动代选默认时辰。

**需要联网吗？**
排盘计算通过 liki.hk 引擎完成，需要网络。不可达时技能会明确告知，不会退回"AI 凭感觉编"。

**我的出生数据会被存储吗？**
不会。技能明确约定：不在对话之外存储出生信息、不索要真实姓名；排盘数据仅在你的对话上下文中使用。

**结果该怎么理解？**
每个结论附命理依据与经典出处，可自行检验。命理为传统文化视角，不构成医疗、法律、金融或重大人生决策依据。

**怎么更新？**
技能启动时自动做版本检查，提示更新时重跑：`npx skills add ml8s/liki -y`。

## 为什么可信

- **计算不靠 AI 编** — 八字/紫微排盘由 Go 天文历算引擎完成：真太阳时校正、夏令时、经纬度时区、VSOP87D 秒级节气。模型只做解读，不推算排盘数据。
- **断语有出处** — 46 张断语真值表共 701 条，每条附经典原文列（《渊海子平》《子平真诠》《滴天髓》《三命通会》《紫微斗数全书》等）。
- **双体系交叉验证** — 八字定主、紫微复核，冲突时显式列证裁决。
- **流程可查** — 每步分析填「输出：□」表，未填不得进入下一步；结论可回溯到具体某一步。
- **独立评测** — 160 道命理师大赛真题（MingLi-Bench），答案隔离、自动判分、数据公开（`tests/`）。

---

## 开发者

### 架构

```
skills/liki-bazi
├── SKILL.md    ← 规则层（流程骨架 + 强制规则）
├── app/        ← 流程层（10 卡：婚姻/事业/财运/…）
├── domains/    ← 知识层（bazi 16 + ziwei 8 篇）
└── tools/      ← 工具层（6 个 Python 工具 + 46 张断语 csv + 2 张因子表）
repo root
├── engine/     ← Go JSON-RPC 天文历算引擎（8 领域）
├── tests/      ← 评测体系（160 题分组 + 答案隔离）
└── scripts/    ← 构建 / 分发索引
```

调用链：SKILL.md 路由到 app 卡 → 卡调用 6 个 Python 工具（`agent_cli.py` 内部编排 RPC 排盘、因子求值与 csv 断语匹配）→ 按 domains 知识解读 → 按卡内模板输出。RPC 方法对 liki-bazi 的 LLM 不可见。

### 引擎镜像

镜像随 GitHub Release 自动发布（CI 全量测试 → 构建推送 + 冒烟）：`docker pull ghcr.io/ml8s/liki-engine:latest`。源码构建：`cd engine && docker compose -f deploy/docker-compose.yml up -d --build`。
### 快速开始

```bash
make hooks        # 安装 git hooks（首次）
make test-all     # 全量：skills 单测 + engine（lint/vet/race/集成/冒烟）+ 全链集成
make check        # 断语表 schema + 文档契约 + 版本一致性
make build-archive # 打包 4 skill + 生成分发索引与归档摘要
```

### 设计原则

- 分层单一职责：根=规则、app=流程、domains=知识、tools=工具
- 单一数据来源：LLM 工具契约以 `tools/skill-tools.json` 为准、断语以 csv 真值表为准
- 双体系交叉：八字定主、紫微复核，冲突显式列证
- 日期版本（CalVer）：VERSION=日期戳 + CHANGELOG，启动做版本自检，里程碑按需 git tag
- 评测驱动：独立判分、答案隔离、数据公开

贡献指南见 [CONTRIBUTING.md](./CONTRIBUTING.md)，版本历史见 [CHANGELOG.md](./CHANGELOG.md)。设计参考了 [mingli-skills](https://github.com/weizeW/mingli-skills)、[bazi-skill](https://github.com/jinchenma94/bazi-skill)、[iztro](https://github.com/SylarLong/iztro)、[MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) 等开源项目。

liki-bazi 的因子契约与完整因子清单见 [docs/FACTOR_MODEL.md](./docs/FACTOR_MODEL.md)。

## 协议与声明

MIT。命理结论为传统文化视角，仅供参考，不构成医疗诊断、法律建议、金融投资预测或重大人生决策；请保持理性。
