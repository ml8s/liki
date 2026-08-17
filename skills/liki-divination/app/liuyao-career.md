---
name: app-liuyao-career
description: 六爻·事业 — 职业方向、升迁、创业、求职
依赖域: liuyao
---

# 六爻·事业占断

## 依赖的领域知识
[必读] - domains/liuyao/yongshen.md「用神取用」
[必读] - domains/liuyao/jixiong.md「吉凶判定」
[必读] - domains/liuyao/patterns.md「特殊格局」

## 用户问法 → 流程侧重
| 用户问题 | 侧重步骤 |
|---------|---------|
| 求职能否成功 | 官鬼状态 + 世爻承接 |
| 能否升迁 | 官鬼旺衰 + 动爻方向 |
| 创业是否可行 | 官鬼+妻财综合 |
| 在职是否稳定 | 官鬼+世爻关系 |

## 📖 断卦方法

### 1. 用神取用
- 官鬼爻为主用神（功名、官运、事业）
- 父母爻为辅用神（文书、合同、手续）
- 世爻为自己（承接能力）
- 应爻为单位/环境

### 2. 断语查询
调用 Python 工具层查询 career.csv 断语表：
```python
from tools.liuyao.duanyu import query, evaluate_factors

factors = evaluate_factors(chart, yong_shen)
results = query('career', factors)
```

### 3. 输出模板

#### 一、断语
Python 工具层输出的断语

#### 二、建议
根据断语给出建议
