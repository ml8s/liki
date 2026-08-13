# app — 用户价值层（场景卡索引）

每张卡 = 一类用户问题 → 领域信号翻译表 + 流程卡（调用 `domains/*` 方法与 `core` 规则引擎）。
frontmatter 声明 `依赖域`（命理概念层）。执行顺序由 SKILL.md「执行主干」统一控制，app 卡只定义领域内查表。

## 八字 + 紫微全流程（Phase 1-7）

| 卡 | 功能 | 依赖域 |
|----|------|--------|
| [mingshu.md](mingshu.md) | **命书/命盘综合解读**（用户无明确问题走此入口） | bazi,ziwei |
| [marriage.md](marriage.md) | 婚姻：何时结婚、婚姻质量、感情走向、离婚 | bazi,ziwei |
| [career.md](career.md) | 事业：职业方向、事业起伏、成就层次 | bazi,ziwei |
| [wealth.md](wealth.md) | 财运：财源类型、收入层次、风险提示 | bazi,ziwei |
| [health.md](health.md) | 健康：脏腑薄弱、易感疾病、健康建议 | bazi,ziwei |
| [study.md](study.md) | 学业：学历层次、学习能力、考试运 | bazi,ziwei |
| [personality.md](personality.md) | 性格：五行基础性格、十神修正、身强弱 | bazi,ziwei |
| [family.md](family.md) | 家庭六亲：父母、兄弟姐妹、子女、祖上 | bazi,ziwei |
| [compatibility.md](compatibility.md) | 合盘合婚：感情匹配、关系评估（双人） | bazi,ziwei |

## 子流程（Phase 8 专项）

| 卡 | 功能 | 依赖域 |
|----|------|--------|
| [divination.md](divination.md) | **占卜断事**（六爻+奇门）：问吉凶、决策、时机 | liuyao,qimen |
| [auspicious.md](auspicious.md) | 择日（黄历）：婚嫁搬家开业吉日 | huangli |
| [fengshui.md](fengshui.md) | 风水（八宅+玄空）：房屋布局、家宅吉凶 | bazhai,xuankong |

## 用神流程（Phase 1-4，无紫微交叉）

| 卡 | 功能 | 依赖域 |
|----|------|--------|
| [naming.md](naming.md) | 起名改名（八字用神 + 姓名学） | qiming,bazi |

## 功能维度速查（用户诉求 → 入口）

| 用户诉求 | 入口 |
|---------|------|
| 应期断事（何时结婚/升迁/发财/生病） | 对应主卡（marriage/career/wealth/health…）Phase 7 考时 |
| 命理报告（综合解读/无明确问题） | `mingshu.md` |
| 合盘（两人关系） | `compatibility.md` |
| 起名/改名 | `naming.md` |
| 问吉凶/占卜 | `divination.md` |
| 择吉日 | `auspicious.md` |
| 看风水 | `fengshui.md` |
