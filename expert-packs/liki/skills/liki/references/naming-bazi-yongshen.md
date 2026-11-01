# 起名用神计划

## 数据流

出生信息明确时，先取得八字用神裁决，再直接复用其中的五行角色计划：

```text
出生信息
  → 排盘
  → 用神裁决能力
  → selected_yongshen.selected
  → 取字能力 / 起名校验能力.selected_yongshen
```

八字域负责用神裁决。起名域只消费计划中的：

```text
yong / xi / ji
```

不接收状态、冲突、证据、规则、策略来源等八字审计字段，也不重新裁决用神。

## 计划形状

```json
{
  "yong": "木",
  "xi": ["水"],
  "ji": ["金"]
}
```

约束：

- `yong` 必须是具体五行。
- `xi`、`ji` 是五行数组。
- 用、喜、忌互斥。

当八字结果是 `resolved` 或 `partial`，且 `selected.yong` 是具体五行时，才能传给起名工具。若结果是 `needs_review`、`unresolved`，或 `selected.yong = 无候选`，省略 `selected_yongshen`，只做字库、禁用字与音律检查。

## 组合策略

| 字位 | 可用角色 | 说明 |
| --- | --- | --- |
| `first_role` | `yong` / `xi` | 第一字位 |
| `second_role` | `yong` / `xi` | 第二字位；省略表示单名 |

默认：

```text
单名：yong
双名：yong + yong
```

可回退或按音义选择：

```text
yong + xi
xi + yong
```

双名至少一个字位是 `yong`。生成阶段不提供 `ji` 字池。

若 `xi` 有多个五行，`取字能力` 会为每个 `xi` 五行返回独立字池，便于语义、音韵与出处筛选。

## 起名校验能力

`起名校验能力` 使用同一个 `selected_yongshen` 返回对象。工具内部只读取 `selected.yong / xi / ji`，对名字每个字的实际五行逐一分类：

```text
yong     命中用神五行
xi       命中喜神五行
ji       命中忌神五行
neutral  不属于上述三类
```

check 不预设「用+用」或「用+喜」，只如实返回每个字的角色。例如：

```text
梓涵 → 梓=yong，涵=xi
梓林 → 梓=yong，林=yong
梓鑫 → 梓=yong，鑫=ji
```

`valid` 只表示结构合法性；五行角色不改变 `valid`。

## 输出

```text
八字依据：用神裁决能力.status / selected
选字策略：first_role + second_role
逐字事实：char、wuxing、wuxing_role
```

不输出分数、排名、置信度或目标命中汇总。
