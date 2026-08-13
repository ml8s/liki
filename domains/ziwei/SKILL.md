---
name: liki-ziwei
description: 紫微斗数模块（组件）。排盘、星曜、宫位、四化、格局。
---

# 紫微斗数

## 知识索引

| domains/ziwei/fangfa/geju.md | 特殊格局 | 星曜组合 | 格局名称+识别 |
| domains/ziwei/fangfa/yingqi.md | 紫微应期 | 领域+宫位 | 吉凶信号 |

### 📋 方法（分析流程）

| 文件 | 用途 | 依据 |
|------|------|------|
| domains/ziwei/fangfa/calibration.md | 紫微考时 | 人生事实 | 时辰校准 |
| domains/ziwei/fangfa/gexing.md | 性格合断 | 八字主+紫微辅 | 综合性格 |

### 📖 断语（符号→现实）

| 文件 | 用途 | 依据 |
|------|------|------|
| domains/ziwei/fangfa/zhuxing.md | 主星性格 | 命宫主星 | 性格基调 |
| domains/ziwei/fangfa/fuxing.md | 辅星作用 | 辅星名称 | 吉凶作用 |
| domains/ziwei/fangfa/gong12.md | 十二宫要点 | 宫位+星曜 | 解读要点 |
| domains/ziwei/fangfa/sihua.md | 四化规则 | 四化星+宫位 | 含义影响 |
| domains/ziwei/fangfa/liunian.md | 流年分析 | 流年命宫+四化+星 | 应年重点领域+吉凶信号 |
| domains/ziwei/fangfa/laiyin.md | 来因宫解读 | 命盘天干 | 人生课题方向 |
| domains/ziwei/fangfa/xiangmao.md | 断长相 | 命宫主星+辅星 | 外表特征描述 |

## 技术流程

### 命盘分析

1. 收集参数 → TimeSet
2. 排八字（仅四柱基础，不读八字方法论）
3. 排紫微盘（调用 `ziwei.chart`）
4. 排大限（调用 `ziwei.daxian`）
   - 大限数据在 RPC 返回的 `palaces[].decadal` 中：
     - `range` — 起止年龄，如 `[4, 13]`
     - `heavenlyStem` / `earthlyBranch` — 大限干支
     - `ages` — 流年年龄明细
   - **严禁自行以五行局起算年龄或以命宫干支顺逆推算大限**，必须使用引擎返回数据
5. 解读：调 `domains/ziwei/fangfa/` 各断语表 → 命宫→身宫→十二宫→四化→三方→格局→综合
6. 产出：结构化数据，由执行主干（SKILL.md Phase 0 路由）分发到对应 app 输出

📖 读取 `domains/ziwei/fangfa/` 下 9 张断语表 → 逐一搜索「清单」填写 □ 填空

### 合盘分析

1. 排双方紫微盘
2. 合盘（调用 `ziwei.bond`）
3. 解读：命宫对位→夫妻宫呼应→四化交集
4. 产出：结构化数据（双方），由执行主干分发到 app/compatibility 输出

📖 搜索 domains/ziwei/fangfa/gong12.md → 读取夫妻宫要点

## 体系边界

- 紫微不串八字十神解释命盘
- 「同名陷阱」：七杀/太阳/太阴在两体系含义不同，不可混用
