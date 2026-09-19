---
name: app-liuyao
description: 六爻问卦 — snapshot 生成、结构化 answer 与追问
依赖域: liuyao
---

# 六爻问卦

> 工具报文：只使用 `divination/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] `divination/domains/liuyao/yongshen.md` + `yuejian.md` + `jixiong.md` + `patterns.md` + `liushou.md` + `yingqi.md`

## 流程

| 步骤 | 动作 | 产物 |
| --- | --- | --- |
| 1 | 确认单一目标、事项和视角 | 可排盘的问题 |
| 2 | 发送 `TOOLS.md §1 liuyao_snapshot` 报文；按输入规则选择 `mode / rounds / yaos` | immutable snapshot |
| 3 | 发送 `TOOLS.md §2 liuyao_ask` 报文，`snapshot` 原样绑定 `$LIUYAO_SNAPSHOT` | 结构化 answer |
| 4 | 解释 focus、evidence、topic_guidance、timing_plan、condition_rules | 结论、阻碍与应期 |
| 5 | 追问继续传同一 snapshot | 不重排上下文 |

## 输入

- 默认 `mode=auto`。
- 用户手动摇币时用 `mode=coins`，传六组三枚“正/反”，顺序为初爻到上爻。
- 高级复现可用 `mode=yaos`，传六个爻值 6-9。
- 普通问事传 `matter`；高级用户才传 `yong_shen`；两者互斥。

## 边界条件

| 场景 | 处理 |
| --- | --- |
| 手动摇币 | 六组三枚“正 / 反”必须按初爻到上爻原样传入 |
| `matter` / `yong_shen` | 只能传一项；普通问事优先 `matter` |
| 追问与新事件 | 原事件复用 snapshot；新事件、新时间或新决策新建 snapshot |
| 高风险事项 | 正常执行，附 `safety_advisory` |

## 输出模板

```text
结论：一句话条件性倾向。
用神：名称、爻位、旺衰、月破 / 旬空 / 墓库。
动爻：动变关系与作用路径。
应期：只解释 answer 引用的候选。
建议：一个现实动作或核查条件。
```

### 示例

```text
结论：面试结果有利的倾向较高，但须在应期内确认。
用神：官鬼（事象）持世，临月令旺相。
动爻：五爻妻财动，生官鬼，外部助力明确。
应期：answer 引用甲午日（3 天内）为首选窗口。
建议：准备面试材料，3 天内出结果概率较高。
边界：传统六爻视角，不承诺现实结果。
```
