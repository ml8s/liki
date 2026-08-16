<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-命理师的_Skill-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — 命理师的 Skill</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-官方网站-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
</p>

---

**Liki** 是命理师的 Skill：在 AI 助手里完成**八字、紫微、六爻、奇门、黄历择日、八宅风水、玄空风水、起名**等 8 个领域的专业分析。它把传统命理实现为**流程可执行、结论可复验、断语可溯源**的工程化工具：排盘由天文历算引擎计算，断语由 46 张真值表（590 条，含经典原文）驱动，流程每步填表，结论用已发生事件校准。

## 快速开始

一次安装全部 4 个 skill：

```bash
npx skills add ml8s/liki
```

| skill | 领域 | 单装 |
|-------|------|------|
| `liki-bazi` | 命理（八字+紫微「八紫」双盘同参） | `--skill liki-bazi` |
| `liki-divination` | 问卦（六爻/奇门/黄历择日） | `--skill liki-divination` |
| `liki-fengshui` | 风水（八宅/玄空） | `--skill liki-fengshui` |
| `liki-naming` | 起名（八字用神+三才五格） | `--skill liki-naming` |

单装某个：`npx skills add ml8s/liki --skill liki-bazi`

安装后在 AI 助手中直接对话即可；各 skill 能做什么、怎么问，见下方「功能特性」。

## 功能特性

### liki-bazi — 命理（八字 + 紫微双盘同参）

覆盖人生主要领域，每个结论附命理依据与应期年份：

- **婚姻**：何时结婚、感情走向、会不会离婚、对方是什么样的人
- **事业**：适合什么行业、创业还是打工、事业起伏年份
- **财运**：财源类型、得财/破财年份
- **健康**：哪个脏腑薄弱、易感疾病、风险年份
- **学业**：能读到什么学历、考试运
- **性格 / 外貌 / 六亲（父母子女）/ 合盘（两人关系）**

示例：`算八字，1990-05-20 12:00 北京出生，男`；`帮我出一份命书`（全盘报告）。

### liki-naming — 起名

八字用神定五行 → 三才五格 → 候选字。支持新生儿起名、改名、公司命名、英文起中文名、自选名字评估。

示例：`宝宝起名，2024-06-10 广州出生，男，姓陈`

### liki-divination — 问卦

- **六爻**：问吉凶成败、应期（"这个事能成吗""什么时候有结果"）
- **奇门**：问方向、时机决策（"往哪个方向好""现在该不该做"）
- **黄历择日**：选吉日（"哪天搬家/结婚/开业好"）

示例：`明天适合搬家吗`

### liki-fengshui — 风水

- **八宅**：命卦、人宅相配、门主灶布局
- **玄空**：元运飞星、旺山旺向、流年风水

示例：`我家风水怎么样`

## 实现机制

**① 排盘：天文历算引擎** — 八字/紫微排盘由开源计算引擎（[liki-engine](https://github.com/ml8s/liki-engine)）完成：真太阳时校正、夏令时、经纬度时区按天文算法计算。模型只做解读，不自行推算排盘数据。

**② 断语：csv 真值表** — 46 张断语真值表（八字 26 + 紫微 20，共 **597 条断语**），每条带**经典原文列**（《渊海子平》《子平真诠》《滴天髓》《三命通会》《紫微斗数全书》等）。流年另有 7 张 yearly 断语表（应期/婚姻/六亲/财运/事业/健康/学业/子女，含值年神煞）。

**③ 流程：每步填表** — 根 SKILL.md 定全局骨架（排盘 → 因子 → 断语；应期走 liunian 链）与强制规则；每个领域按对应 app 卡流程逐步执行，每步填「输出：□」表，□ 未填不得进入下一步，结论回溯到已填的 □。

**④ 上下文因子消除断语冲突** — 同盘多断语互相矛盾是命理判断的难点，用上下文因子解决：
- **月令格神定性格主面**（《子平真诠》"月令为提纲，格神主性"）——性格题主面唯一，十神旺衰只作辅面
- **星宫同参**（《三命通会》）——配偶星透干（星吉）+ 夫妻宫被冲（宫凶）= 婚可成但波折
- **值年神煞**（《协纪辨方书》）——病符/丧门/吊客/大耗按太岁查，命局四柱逢即应
- 全断语表**零冲突验证**：扫描 19 域"同盘命中矛盾条目"→ 补上下文因子 → 全部 0 撞

**⑤ 考时校准** — 排盘前用已发生事件校准时辰；结论前用命主 3-5 个已发生时间段反向验证（"该时段命理引动 → 应验何事"），≥2 段吻合结论才成立。

## 开发者

面向开发者：八字/紫微断语全部 CSV 真值表化（可读可改可评审）；占卜/风水（六爻/奇门/八宅/玄空）断语为统一「翻译表」，确定性计算下沉引擎、断语由 LLM 按表翻译；评测体系随 skill 分发（160 题答案隔离自动判分可复现）；引擎与 skill 分层（liki-engine 开源计算引擎）。

### 评测：体系可复现，数据见发布帖

Liki 在 [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench)（160 道命理师大赛真题）上建立了独立评测体系：**答案与题目分离、Docker 沙箱隔离运行（agent 物理读不到答案）、skill-up 自动判分**，每轮改动可回归对比。

评测数据（正确率/迭代记录）见发布帖与 CHANGELOG，本 README 不发布。

**复现**（评测配置与判分脚本随 skill 分发；答案文件与题目分离，运行时自动移出容器挂载，agent 物理读不到答案）：

```bash
bash tests/run-qwen.sh --parallelism 16          # 移答案→评测→恢复→自动判分
python3 -m pytest tests/ -q --ignore=tests/test_integration.py   # 单测（因子/断语/agent_cli 分派）
```

> 评测配置（160 题、答案隔离、自动判分）随仓库可复现，评测数据见发布帖与 CHANGELOG。

### 架构

一个 skill = **文档层 + 工具层**，外部接**引擎**（天文历算）。文档层内部再按职责分三层：

```
skill ── 文档层 ── 根 SKILL.md      ← 规则（全局骨架 + 强制填表 + 路由 + RPC 说明 + 输出/交互/行为）
      │         ├─ app/             ← 流程（每领域排盘 → 查断语 → 输出，每步「输出：□」+ 输出模板）
      │         └─ domains/<域>/    ← 知识（方法论 + 断语翻译，域下平铺）
      └─ 工具层 ── tools/           ← 5 工具（schema + 实现）+ 断语 csv（46 张）+ 因子 factors.csv
外部 ── 引擎 ── liki-engine         ← 开源 JSON-RPC 天文历算（八字/紫微/六爻/奇门/风水）
```

调用关系：**根 SKILL.md 读 app 卡 → app 卡调工具（tools/ 内部：RPC 排盘 + csv 断语匹配）→ 工具返回断语 → 按 domains/ 知识解读 → 按 app 卡输出模板成稿**。csv 是工具内部数据（agent 不读），不单独成层。工具层是**可选组成**——有确定性计算需求的 skill 才有（liki-bazi 有，divination/fengshui/naming 无工具层，直接 RPC + 文档翻译）。

### 项目结构

```
仓库根（liki-skills 工程区——npx skills 安装时不装）
├── skills/                     ← 4 个独立 skill（`npx skills add ml8s/liki` 一次装全部）
│   ├── liki-bazi/              ← 命理（八字+紫微「八紫」双盘同参）
│   │   ├── SKILL.md            ← 规则层（流程约定 + RPC 说明 + 输出/交互/行为）
│   │   ├── app/                ← 流程层（9 卡：命书/婚姻/事业/财运/健康/学业/性格/家庭/合盘）
│   │   ├── domains/            ← 解读层（bazi 16 篇 + ziwei 8 篇，域下平铺）
│   │   ├── tools/              ← 工具层（skill-tools.json + 5 工具 + 断语 csv + 因子）
│   │   │   ├── skill-tools.json ← 工具 schema（parameters + result_schema）
│   │   │   ├── bazi/           ← 八字断语表 26 张（含 yearly_* 流年 7 张）
│   │   │   ├── ziwei/          ← 紫微断语表 20 张
│   │   │   ├── factors/        ← 因子定义（本命 + 流年）
│   │   │   ├── paipan.py       ← 排盘（full_paipan/liunian）
│   │   │   └── duanyu.py       ← 断语查询（query/match）
│   │   └── VERSION / content.sha256  ← 版本 + 内容指纹（自检）
│   ├── liki-divination/        ← 问卦（六爻/奇门/黄历择日）
│   ├── liki-fengshui/          ← 风水（八宅/玄空）
│   └── liki-naming/            ← 起名（八字用神 + 三才五格）
├── tests/      ← 评测体系（160 题分组 case 挂 liki-bazi、答案分离、skill-up script judge）
├── scripts/    ← 构建脚本（build-archive.sh 打 4 包）
└── webapp/     ← Web 集成流水线（liki-bazi 部署附带）
```

### 设计原则

- **分层单一职责** — 根=规则、app=流程、domains=知识、tools=工具；各层不串层（流程只归 app，规则只归根，知识只归 domains）。
- **单一数据来源** — 参数/返回字段以 rpc.discover（或 skill-tools.json result_schema）为准，断语结论以 query（csv 真值表）为准；域文档不重复这三者，只写 rpc/csv 没有的（业务映射、判断链、约束规则、体系隔离）。
- **双体系交叉** — 八字定主、紫微复核，冲突显式列证裁决。
- **语义版本管理** — VERSION + CHANGELOG，每版记录改动与评测数据。
- **评测** — 独立自动判分、答案隔离、数据公开。

## 参考

设计参考了以下开源项目：

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — 四维交叉验证框架，提示词工程化方法论参考
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — 经典摘要参考文件设计、历史事件校准机制
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — 八字+紫微综合印证模式
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — 结论先行输出设计
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — MCP 全栈命理，Agent 友好设计参考
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — 紫微斗数排盘引擎
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — 三层计算架构
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — 多流派互审与结论强度标注理念
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — 命理评测参照
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — 置信度体系、人格画像推断

## 协议

MIT。命理结论为传统文化视角，仅供参考，不构成医疗诊断、法律建议、金融投资预测或重大人生决策；请理性看待。
