# 问结果

> 场景：用户核心目标是事件结果或应期。

## 依赖的领域知识

[必读] `references/domains/liuyao/yongshen.md` + `references/domains/liuyao/yuejian.md` + `references/domains/liuyao/jixiong.md` + `references/domains/liuyao/patterns.md` + `references/domains/liuyao/liushou.md` + `references/domains/liuyao/yingqi.md`

## 流程

| 步骤 | 动作 | 产物 |
| --- | --- | --- |
| 1 | 确认事项与视角 | 起卦参数 |
| 2 | 调六爻起卦工具创建快照；起卦方式按输入规则自举（参数见工具 schema） | immutable snapshot |
| 3 | 调六爻追问工具出断语；`snapshot` 原样绑定 `$LIUYAO_SNAPSHOT` | 结构化 answer |
| 4 | 解释 focus、evidence、topic_guidance、timing_plan、condition_rules | 结论、阻碍与应期 |
| 5 | 追问继续传同一 snapshot | 不重排上下文 |

## 起卦输入

- 默认自动起卦（线上安全随机数）；纯文字对话无法实际摇硬币，先告知用户线上起卦没有实地摇卦的仪式性和传统局限，由用户确认接受后再继续。
- 用户手动摇币时，按初爻到上爻给出六组三枚“正 / 反”，原样传入。
- 高级复现可传六个爻值 6-9。
- 普通问事用事项用神，高级用户才传显式用神。

## 边界条件

| 场景 | 处理 |
| --- | --- |
| 手动摇币 | LLM / Python 不得换算爻值或判断动爻；爻值与动爻以工具返回为准 |

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
