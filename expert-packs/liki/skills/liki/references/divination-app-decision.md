# 问决策

> 场景：用户核心目标是策略、方向或行动时机。

## 依赖的领域知识

[必读] `references/divination-qimen-yongshen.md` + `references/divination-qimen-bamen.md` + `references/divination-qimen-jiuxing.md` + `references/divination-qimen-bashen.md` + `references/divination-qimen-yingqi.md` + `references/divination-qimen-quarter.md`
[专占] `references/divination-qimen-lost-property.md` + `references/divination-qimen-thief-capture.md` + `references/divination-qimen-missing-person.md` + `references/divination-qimen-capture-escape.md`

## 流程

| 步骤 | 动作 | 产物 |
| --- | --- | --- |
| 1 | 确认地点与事项 | 排盘参数 |
| 2 | 使用 奇门一次性问测 一次完成排盘、投影和解读；参数按 契约 自举（普通问事事项 / 显式用神 / 专占） | Reading |
| 3 | 解释 Reading evidence 中的用神落宫、门星神、生克、空亡、马星和应期 | 方向、态势与时机 |
| 4 | 追问基于已返回 Reading；超出 scope 时先确认并新建问测 | 延续当前问测上下文 |

## 常用参数

用户没有明确高级方法要求时，按当前 schema 声明的默认方式调用；schema 若表达 default，就使用 default。若 schema 未表达默认但存在必选方法组合，先选择 schema 描述中的常规定义，不臆造高级参数；仍无法判定时询问用户。高级组合只在用户明确要求时按 schema 传入。不能从“更细”“传统”“准确”推断高级方法。

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

- 普通问事用事项用神（契约 二选一）。
- 缺地点时先询问城市或经度；使用城市检索能力时，多候选或未命中必须让用户确认，确认后按 schema 组装地点事实。
- 失物 / 偷盗 / 走脱 / 画像专占不接受事项或用神。
- 走失专占必须用走失问事。
- 金函玉镜改列九宫与十二日神，不输出局数、用神、值符值使与应期。

## 输出模板

```text
结论：一句话方向 / 时机 / 决策倾向。
用神：符号、落宫、旺衰、空亡 / 马星。
态势：支持与阻碍来源。
时机：Reading 引用的应期候选。
建议：一个现实动作或核查条件。
边界：传统奇门视角，不承诺现实结果。
```

### 示例

```text
谈判窗口偏有利。用神：开门落坎宫（旺）。时机：甲申时最优。建议：下午主动约谈。
```
