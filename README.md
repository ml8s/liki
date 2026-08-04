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
  <a href="./README.en.md"><img src="https://img.shields.io/badge/English-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

**Liki** 是命理师的 Skill，基于精密计算引擎与系统化方法论构建，为命理师提供可靠、可复验的专业工具。在 AI 助手中完成八字、紫微、起名、六爻、奇门、择日、风水等全面命理分析。

覆盖 **8 个独立领域**，内置 **9 份方法论文档**，八字紫微**双体系交叉验证**，报告流程经 **generate → review → revise** 三阶段审查。

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

## 安装

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

## 架构

```
Liki（本仓库）     → Skill（流程定义 + 方法论文档）
liki-engine        → 天文历算 API（[开源计算引擎](https://github.com/ml8s/liki-engine)）
[liki.hk](https://liki.hk)        → 官方网站，基于 engine + skill 构建
```

## 项目结构

```
├── SKILL.md    ← 路由 + 共性规则
├── app/        ← 用户应用（13个：命盘/婚姻/健康/事业/财运/学业/性格/家庭/择日/占卜/合盘/起名/风水）
├── domains/    ← 领域层（8域）
│   ├── bazi/       ← 八字（断语6 + 方法9）
│   ├── ziwei/      ← 紫微（断语10 + 方法2）
│   ├── liuyao/     ← 六爻（断语3）
│   ├── qimen/      ← 奇门（断语2）
│   ├── huangli/    ← 黄历（断语2）
│   ├── bazhai/     ← 八宅（断语1）
│   ├── xuankong/   ← 玄空（断语2）
│   └── qiming/     ← 起名（断语2）
└── webapp/    ← Web 集成流水线
```

## 特性

- **域与应用分离** — 领域层放断语（符号→现实翻译表）与方法（分析流程），应用层放流程卡。修知识不碰流程，调流程不影响应用。
- **39 个领域文件**（断语28 + 方法11）— 旺衰/用神/应期/合化/冲/学历/财运/事业等判断从"AI理解"改为"查表"，输出不再飘忽不定。
- **填空式清单** — □ 填空代替打勾，不填完输出残缺，AI 无法跳步。
- **星动+宫动双验证** — 流年应期排队列机制，八字紫微交叉验证。
- **考时校准** — 用已发生的人生事件反向验证时辰。
- **天文历算** — 真太阳时/夏令时/经纬度时区校正。
- **语义版本管理** — VERSION + CHANGELOG。
- **MingLi-Bench**: 初答 35% → 关卡制 Gate 强制重测 73/73 全对（160 道命理师大赛真题，严格走流程、填完检查表才给结论）。

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
