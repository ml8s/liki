<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-命理师的_Skill-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — 命理师的 Skill（v3.8.1）</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-官方网站-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

**Liki** 是命理师的 Skill：在 AI 助手里完成**八字、紫微、起名、六爻、奇门、择日、风水**等 8 个领域的专业分析——不是"让 AI 自由发挥算命"，而是把传统命理做成**流程可执行、结论可复验、断语可溯源**的工程化工具。

- **给你的价值**：一个结论，四种保障——排盘不靠模型（天文历算引擎）、断语不靠发挥（46 张真值表 701 条断语全部带经典原文）、流程不许跳步（关卡制检查表）、结论可用你的人生事件复核（考时校准）。
- **给开发者的价值**：断语全部 CSV 真值表化（可读可改可评审）、评测体系随 skill 分发（160 题答案隔离自动判分可复现）、引擎与 skill 分层（liki-engine 开源计算引擎）。

## 你能用它做什么

| 场景 | 输入示例 | Liki 会怎么做 |
|------|---------|--------------|
| 看八字 | `1990-05-20 12:00 北京出生，男` | 真太阳时校正 → 排盘 → 强弱用神 → 格局 → 领域断语 |
| 问婚姻/事业/财运 | `帮我看看这几年事业运` | 主域识别 → 配偶星/官星查表 → 八字紫微交叉 → 流年应期候选 |
| 排紫微 | `1988-03-15 上海出生，女` | 农历排盘 → 十二宫四化 → 大限流年 |
| 宝宝起名 | `2024-06-10 广州出生，男，姓陈` | 用神定五行 → 五格三才 → 候选字 |
| 择日/占卜 | `明天适合搬家吗` | 黄历 + 八宅/玄空 → 吉凶判断 |
| 综合命书 | `帮我出一份命书` | 八字+紫微全流程 → 综合论断 + 双盘报告 |

**它和"AI 直接算命"的区别**：你问"1996 年发生了什么"，普通 AI 凭训练记忆编答案；Liki 会先调引擎排 1996 年流年盘，查真值表断语，再给出有命理依据的应期候选——**每一步都有据可查**。

## 为什么可以信任 Liki

可靠性不是口号，是五条硬机制：

**① 天文历算引擎，排盘不靠模型** — 八字/紫微排盘由开源计算引擎（[liki-engine](https://github.com/ml8s/liki-engine)）完成：真太阳时校正、夏令时、经纬度时区全部按天文算法计算。模型只负责解读，**禁止自行推算排盘数据**。

**② 断语表驱动，不靠模型发挥** — 46 张断语真值表（八字 26 + 紫微 20，共 **701 条断语**），每条带**经典原文列**（《渊海子平》《子平真诠》《滴天髓》《三命通会》《紫微斗数全书》等）——规则就是规则，可读、可改、可评审，不存在"参考"弱化。流年另有 7 张 yearly 断语表（应期/婚姻/六亲/财运/事业/健康/学业/子女，含值年神煞）。

**③ 关卡制执行主干，不允许跳步** — 所有题型走统一流程（Phase 0-8）：路由 → 时辰判定 → 排盘快照 → 强弱用神 → 领域查表 → 紫微交叉 → 考时自洽。每阶段有**填空式检查表**（□ 填空代替打勾），不填完不得进入下一阶段。

**④ 命理深度：上下文因子消除断语冲突** — 命理判断的难点是"同盘多断语互相矛盾"。Liki 用命理上下文因子系统性解决：
- **月令格神定性格主面**（《子平真诠》"月令为提纲，格神主性"）——性格题主面唯一，十神旺衰只作辅面
- **星宫同参**（《三命通会》）——配偶星透干（星吉）+ 夫妻宫被冲（宫凶）= 婚可成但波折
- **值年神煞**（《协纪辨方书》）——病符/丧门/吊客/大耗按太岁查，命局四柱逢即应
- 全断语表**零冲突验证**：扫描 19 域"同盘命中矛盾条目"→ 补上下文因子 → 全部 0 撞

**⑤ 考时校准，用人生事实验证结论** — 排盘前用已发生事件校准时辰；结论前用命主 3-5 个已发生时间段反向验证（"该时段命理引动 → 应验何事"），≥2 段吻合结论才成立。

## 快速开始

```bash
npx skills add ml8s/liki
```

然后在 AI 助手中直接对话：

> 算八字，1990-05-20 12:00 北京出生，男
> 排个紫微盘，1988-03-15 上海出生，女
> 宝宝起名，2024-06-10 广州出生，男，姓陈
> 明天适合搬家吗

也可生成综合命书报告：

> 帮我出一份命书

AI 助手会完成八字+紫微全流程分析，输出综合论断 + 八字报告 + 紫微报告。

## 评测：体系可复现，数据见发布帖

Liki 在 [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench)（160 道命理师大赛真题）上建立了独立评测体系：**答案与题目分离、Docker 沙箱隔离运行（agent 物理读不到答案）、skill-up 自动判分**，每轮改动可回归对比。

**评测数据（正确率/迭代记录）不在本 README 发布**——过程数据与实测数字见发布帖与 CHANGELOG（避免 README 沦为宣传页）。

**复现**（评测配置与判分脚本随 skill 分发；答案文件与题目分离，运行时自动移出容器挂载，agent 物理读不到答案）：

```bash
cd skills/liki
skill-up validate tests/eval-grouped-qwen.yaml    # 校验 32 个 case
bash tests/run-qwen.sh --parallelism 16          # 移答案→评测→恢复→自动判分
```

> 评测体系本身（160 题、答案隔离、自动判分、可复现）就是工程严谨度的证据；任何数字都附评测配置可复现，不虚标。

## 架构

```
┌─ 流程层  SKILL.md Phase 0-8（关卡制）＋ app/ 13 卡（按场景路由）
├─ 断语层  tools/ 46 张真值表（701 条断语＋经典原文）＋ 因子引擎（497 行因子定义）
├─ 工具层  tools/ 排盘/因子/断语查询 5 函数（full_paipan/make_factors/query…）
└─ 引擎层  liki-engine（开源 JSON-RPC 天文历算：八字/紫微/六爻/奇门/风水）
```

数据链路：**引擎排盘（天文历算）→ 因子生成（真值表）→ 断语查询（经典原文）→ 关卡流程 → 双盘交叉 → 考时校准**——每一环都可审计。

## 项目结构

```
├── SKILL.md    ← 执行主干（Phase 0-8 路由 + 共性规则 + 裁决）
├── app/        ← 应用层（13 卡：命盘/婚姻/健康/事业/财运/学业/性格/家庭/择日/占卜/合盘/起名/风水）
├── domains/    ← 领域层（8 域，断语+方法文档）
├── tools/      ← 推理机（排盘/因子/断语查询）
│   ├── bazi/       ← 八字断语表 26 张（含 yearly_* 流年 7 张）
│   ├── ziwei/      ← 紫微断语表 20 张
│   ├── factors/    ← 因子定义（本命 497 行 + 流年因子）
│   ├── paipan.py   ← 排盘（full_paipan/liunian）
│   └── duanyu.py   ← 断语查询（query/match）
├── tests/      ← 评测体系（160 题分组 case、答案分离、判分脚本）
└── webapp/     ← Web 集成流水线
```

## 设计原则

- **域与应用分离** — 领域层放断语（符号→现实翻译表）与方法（分析流程），应用层放流程卡。修知识不碰流程，调流程不影响应用。
- **断语表驱动** — 断语 CSV 真值表 + 经典原文列，规则可读可改可评审，不靠模型记忆。
- **上下文因子消冲突** — 月令格神定主面、星宫同参、值年神煞——补上下文因子消除断语矛盾（全域零撞验证）。
- **关卡制执行** — 填空式检查表强制每步落纸，模型无法跳步。
- **双体系交叉** — 八字定主、紫微复核，冲突显式列证裁决。
- **考时回环** — 结论用命主已发生事件验证，不吻合则重推。
- **语义版本管理** — VERSION + CHANGELOG，每版记录改动与评测数据。
- **诚实评测** — 独立自动判分、答案隔离、数据公开，不虚标成绩。

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

MIT。命理结论为传统文化视角，仅供参考，不构成专业建议。
