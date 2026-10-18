# 问决策

> 场景：用户核心目标是策略、方向或行动时机。

## 依赖的领域知识

[必读] `references/domains/qimen/yongshen.md` + `references/domains/qimen/bamen.md` + `references/domains/qimen/jiuxing.md` + `references/domains/qimen/bashen.md` + `references/domains/qimen/yingqi.md` + `references/domains/qimen/quarter.md`
[专占] `references/domains/qimen/lost-property.md` + `references/domains/qimen/thief-capture.md` + `references/domains/qimen/missing-person.md` + `references/domains/qimen/capture-escape.md`

## 流程

| 步骤 | 动作 | 产物 |
| --- | --- | --- |
| 1 | 确认地点与事项 | 排盘参数 |
| 2 | 使用奇门排盘能力创建快照；参数按 schema 自举（普通问事事项 / 显式用神 / 专占） | 不可变快照 |
| 3 | 使用奇门追问能力出断语；快照对象原样传递 | 结构化断语 |
| 4 | 解释 factors 中的用神落宫、门星神、生克、空亡、马星和应期 | 方向、态势与时机 |
| 5 | 追问继续传同一 snapshot | 不重排上下文 |

## 常用参数

用户没有明确要求，不传方法参数；默认时家转盘常规定局；刻家 / 流派 / 定局只在用户明确要求时传对应高级参数。不能从“更细”“传统”“准确”推断高级方法。

| 用户问题 | 排盘意图 |
| --- | --- |
| 钥匙丢了，能不能找到 | 失物专占 |
| 东西被偷 / 逃走的人 / 偷者画像 | 偷盗 / 走脱 / 画像专占 |
| 工作能不能升、财能不能求、婚姻如何 | 普通问事（事项用神） |
| 家人走失 | 走失专占 |
| 用置闰盘看现在 | 置闰定局 |
| 用洛书飞盘看这件事 | 飞盘流派 |
| 用十分钟刻家看看 | 十分钟三元刻家 |
| 用十二分钟十分局看看 | 十二分钟十分局刻家 |
| 看今天整体 / 用金函玉镜看今天 | 日盘金函玉镜 |

## 边界条件

- 普通问事用事项用神（schema 二选一）。
- 缺地点时先要城市或经度。
- 失物 / 偷盗 / 走脱 / 画像专占不接受事项或用神。
- 走失专占必须用走失问事。
- 金函玉镜改列九宫与十二日神，不输出局数、用神、值符值使与应期。

## 输出模板

```text
结论：一句话方向 / 时机 / 决策倾向。
用神：符号、落宫、旺衰、空亡 / 马星。
态势：支持与阻碍来源。
时机：answer 引用的应期候选。
建议：一个现实动作或核查条件。
边界：传统奇门视角，不承诺现实结果。
```

### 示例

```text
谈判窗口偏有利。用神：开门落坎宫（旺）。时机：甲申时最优。建议：下午主动约谈。
```
