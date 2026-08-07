<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-命理师的_Skill-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — 命理师的 Skill（v2.4.0）</strong>
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

**Liki** 是命理师的 Skill：用工程方法把传统命理做成**流程可执行、结论可复验**的专业工具。在 AI 助手中完成八字、紫微、起名、六爻、奇门、择日、风水等分析，覆盖 **8 个独立领域**。

**它不是"让 AI 自由发挥算命"**，而是一套关卡制方法论：每个结论都必须先排盘（天文历算引擎，非模型推算）、再填完对应检查表、再经八字紫微双体系交叉验证、最后用命主已发生的人生事件考时校准——任一步缺失，结论自动降级。

## 领域

| 领域 | 说明 |
|------|------|
| 八字 | 排盘、用神、格局、合会冲刑、合盘、大运流年 |
| 紫微 | 十二宫、四化、三方四正、特殊格局、大限流年 |
| 起名 | 八字用神 + 五格三才。支持外国人英文定中文姓 |
| 六爻 | 起卦→装卦→断卦 |
| 奇门 | 排盘、断事、择吉 |
| 黄历 | 择日查询 |
| 八宅 | 命卦配门主灶 |
| 玄空 | 飞星盘、三元九运、流年飞星 |

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

## 为什么可以信任 Liki

可靠性不是口号，是四条硬机制：

**① 天文历算引擎，排盘不靠模型** — 八字/紫微排盘由开源计算引擎（[liki-engine](https://github.com/ml8s/liki-engine)）完成：真太阳时校正、夏令时、经纬度时区，全部按天文算法计算。模型只负责解读，**禁止自行推算排盘数据**（写不出 RPC 返回的大运干支/大限起止，结论无效）。

**② 关卡制执行主干，不允许跳步** — 所有题型走统一流程（Phase 0-8）：路由 → 时辰判定 → 排盘快照 → 强弱用神 → 领域查表 → 紫微交叉 → 考时自洽。每阶段有**填空式检查表**（□ 填空代替打勾），不填完不得进入下一阶段。

**③ 八字 + 紫微双体系交叉验证** — 两套体系独立排盘、互相复核：紫微对生死/无子/残疾类有**一票否决权**，一般冲突以八字用神为纲并**显式列出双方证据**裁决；紫微四化只作佐证不作门槛。

**④ 考时校准，用人生事实验证结论** — 排盘前用已发生事件校准时辰；结论前用命主 3-5 个已发生时间段反向验证（"该时段命理引动 → 应验何事"），≥2 段吻合结论才成立。这不是玄学话术，是**可验证的回环**。

## 评测：实事求是，可复现

Liki 在 [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench)（160 道命理师大赛真题）上建立了独立评测体系：**答案与题目分离、Docker 沙箱隔离运行（agent 物理读不到答案）、skill-up 自动判分**，每轮改动都有回归数据：

| 版本 | 正确率 | 说明 |
|------|--------|------|
| v1 | 36.9% | 初版基线 |
| v2 | 38.8% | +9 项规则 |
| v3 | 40.0% | +7 项护栏 |
| v4 | 40.0% | 执行主干重构（统一 Phase 0-8、去重复） |
| **2.4.0** | 待评测 | 第 3 轮系统性优化 |

40% 级正确率是**诚实的现状**：命理判断（尤其婚姻应期、六亲灾亡、学历档位）对当前轻量模型仍是难点，Liki 的价值在于流程纪律 + 每轮改动的可观测迭代。

**复现**（评测配置与判分脚本随 skill 分发；答案文件与题目分离，运行时自动移出容器挂载，agent 物理读不到答案）：

```bash
cd skills/liki
skill-up validate evals/eval-grouped-qwen.yaml    # 校验 32 个 case
bash evals/run-qwen.sh --parallelism 16          # 移答案→评测→恢复→自动判分
```

> 我们不宣称"准确率 100%"——任何不可复现的数字都不写进文档。评测体系本身（160 题、隔离、自动判分）就是专业度的证据。

## 架构

```
Liki（本仓库）     → Skill（流程定义 + 方法论文档）
liki-engine        → 天文历算 API（开源计算引擎）
[liki.hk](https://liki.hk)        → 官方网站，基于 engine + skill 构建
```

## 项目结构

```
├── SKILL.md    ← 执行主干（Phase 0-8 路由 + 共性规则 + 裁决）
├── app/        ← 应用层（13 卡：命盘/婚姻/健康/事业/财运/学业/性格/家庭/择日/占卜/合盘/起名/风水）
├── domains/    ← 领域层（8 域，38 份断语+方法文档）
│   ├── bazi/       ← 八字（断语 6 + 方法 9）
│   ├── ziwei/      ← 紫微（断语 10 + 方法 2）
│   ├── liuyao/     ← 六爻（断语 3）
│   ├── qimen/      ← 奇门（断语 2）
│   ├── huangli/    ← 黄历（断语 2）
│   ├── bazhai/     ← 八宅（断语 1）
│   ├── xuankong/   ← 玄空（断语 2）
│   └── qiming/     ← 起名（断语 2）
├── evals/      ← 评测体系（160 题分组 case、答案分离、判分脚本）
└── webapp/     ← Web 集成流水线
```

## 设计原则

- **域与应用分离** — 领域层放断语（符号→现实翻译表）与方法（分析流程），应用层放流程卡。修知识不碰流程，调流程不影响应用。
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
