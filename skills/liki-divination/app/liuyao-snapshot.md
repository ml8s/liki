---
name: app-liuyao
description: 六爻问卦 — snapshot 生成、结构化 answer 与追问
依赖域: liuyao
---

# 六爻问卦

## 依赖的领域知识

[必读] `domains/liuyao/yongshen.md` + `yuejian.md` + `jixiong.md` + `patterns.md` + `liushou.md` + `yingqi.md`

## 流程

| 步骤 | 动作 | 产物 |
|---|---|---|
| 1 | 确认单一目标、事项和视角 | 可排盘的问题 |
| 2 | 调 `liuyao_snapshot`：传 `question`、`matter` 或 `yong_shen`，必要时传 `mode/rounds/yaos` | immutable snapshot |
| 3 | 调 `liuyao_ask(snapshot, message)` | 结构化 answer |
| 4 | 解释 focus、evidence、topic_guidance、timing_plan、condition_rules | 结论、阻碍与应期 |
| 5 | 追问继续传同一 snapshot | 不重排上下文 |

## 输入

- 默认 `mode=auto`。
- 用户手动摇币时用 `mode=coins`，传六组三枚“正/反”，顺序为初爻到上爻。
- 高级复现可用 `mode=yaos`，传六个爻值 6-9。
- 普通问事传 `matter`；高级用户才传 `yong_shen`；两者互斥。

## 输出模板

```text
结论：一句话条件性倾向。
用神：名称、爻位、旺衰、月破 / 旬空 / 墓库。
动爻：动变关系与作用路径。
应期：只解释 answer 引用的候选。
建议：一个现实动作或核查条件。
边界：传统六爻视角，不承诺现实结果；不构成医疗、法律、财务或其他专业建议。
```
