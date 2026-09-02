# 命理金样例

本目录是命理规则的**正确性证据**，不是普通测试数据。

## 目标

每个高风险命理规则都必须有三类样本：

| 类型 | 目的 |
|---|---|
| positive | 规则应当成立 |
| negative | 规则必须不成立 |
| boundary | 边界条件，如半合、缺支、无根、保存 pan 复用 |

当前覆盖因子层和流年原语。

## 文件

```text
cases.json
```

顶层结构：

```json
{
  "version": 1,
  "cases": []
}
```

## Case 类型

### 1. factor_snapshot

验证 `factors.csv` / `factors_liunian.csv` 生成的因子快照。

```json
{
  "id": "sanhe_full_group",
  "rule": "三合必须三方齐备",
  "kind": "positive",
  "category": "liunian_relation",
  "mode": "factor_snapshot",
  "shushi": "bazi",
  "input": {
    "base": {},
    "gender": "male",
    "chart": {},
    "current_year": 0
  },
  "expect_factors": {
    "流年配偶星合会": 1
  },
  "basis": "申子辰三方齐备，流年申合日支子。"
}
```

`expect_factors` 是精确断言：只检查列出的因子，不允许未列因子干扰本例意图。

### 2. flow_operator

验证单个流年原语的边界规则。

```json
{
  "id": "sanhe_half_group",
  "kind": "negative",
  "category": "liunian_relation",
  "mode": "flow_operator",
  "input": {
    "op": "流年合",
    "args": ["配偶星"],
    "base": {},
    "gender": "male",
    "chart": {},
    "ctx": {}
  },
  "expect": 0,
  "basis": "申子只是半合，缺辰不成三合局。"
}
```

## 硬约束

1. 每个样本必须有 `rule`、`kind`、`basis`。
2. `kind` 只能是 `positive` / `negative` / `boundary`。
3. `expect_factors` / `expect` 不允许为空。
4. 禁止在样本里写“吉”“凶”“好命”等应用结论；这里只证明命理规则。
5. 修改真值表或算子后，必须先跑本目录测试。
