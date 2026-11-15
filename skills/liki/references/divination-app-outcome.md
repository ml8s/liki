# 问结果

> 场景：用户核心目标是事件结果或应期。

## 依赖的领域知识

[必读] `references/divination-liuyao-yongshen.md` + `references/divination-liuyao-yuejian.md` + `references/divination-liuyao-jixiong.md` + `references/divination-liuyao-patterns.md` + `references/divination-liuyao-liushou.md` + `references/divination-liuyao-yingqi.md`

## 流程

| 步骤 | 动作 | 产物 |
| --- | --- | --- |
| 1 | 确认事项与视角（`matter`；`relationship` 还需 `perspective`） | 起卦参数 |
| 2 | 调用六爻起卦方法 六爻起卦方法，只传 `origin` 与 `casting_context`（时间、经度、focus），创建唯一 event；禁止自动重试 | event |
| 3 | 调用 六爻排盘方法 传入同一 event，取得盘面与 `chart.analysis` | chart |
| 4 | 调用 六爻解读方法 传入同一 event，取得 evidence、应期候选、条件候选与受控解读 | reading |
| 5 | 结合自己持有的 `question` 解释 chart 与 reading；追问沿用同一 event 与已返回断语，不重新起卦 | 结论、阻碍与应期 |

## 起卦输入

- 起卦地点必须来自用户确认的地点事实：用户给城市时先用城市检索能力，多候选时等待确认；未给地点先询问，不得用默认地点代替（精度要求）。
- 默认自动起卦（线上安全随机数）；纯文字对话无法实际摇硬币，先告知用户线上起卦没有实地摇卦的仪式性和传统局限，由用户确认接受后再继续。
- 起卦方法失败时，不得自动重发请求或重新起卦；先告知用户未取得有效卦象，用户明确确认后重新起卦。
- 六爻排盘方法 / 六爻解读方法 只消费同一 event，绝不创建新卦；两者都不读 `question`，领域化解释由持有问题的调用方完成。
- 起卦只传 `origin` 与 `casting_context`；`question` 不进入 MCP 参数。返回的 event 不含 question / origin / focus。
- 用户手动摇币时，按当前 schema 声明的 coins 形态传入六组三枚“正 / 反”；顺序保持初爻到上爻，组内按用户原始顺序原样传递，不自行换算爻值或判断动爻。
- 复现历史卦可传 `origin.mode=yaos` 的六个爻值 6-9。
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
应期：只解释 Reading 引用的候选。
建议：一个现实动作或核查条件。
```

### 示例

```text
结论：面试结果有利的倾向较高，但须在应期内确认。
用神：官鬼（事象）持世，临月令旺相。
动爻：五爻妻财动，生官鬼，外部助力明确。
应期：Reading 引用甲午日（3 天内）为首选窗口。
建议：准备面试材料，3 天内出结果概率较高。
边界：传统六爻视角，不承诺现实结果。
```
