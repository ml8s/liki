---
name: app-health
description: 健康诊断 — 脏腑薄弱、易感疾病、健康建议
依赖域: bazi,ziwei
---

# 健康分析

> 工具报文：只使用 `natal/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: liki-bazi/skills/bazi/wuxing-health.md「五行所属」

- bazi: liki-bazi/skills/bazi/hehui.md「冲宫位表」+「冲吉凶表」（按需——论冲害细节时读取）
- bazi: liki-bazi/skills/bazi/dayun.md「应期决策表」（按需——应期走疾厄宫闭环）
- bazi: liki-bazi/skills/bazi/tiaohou.md「调候用神」（按需——冬夏寒燥失衡时读取）

[必读] - ziwei: liki-ziwei/skills/ziwei/yingqi.md「健康紫微应期」

> **本卡的 本命判断 / 应期判断 调用仅保留健康域断语。**

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 脏腑倾向 | 调 `liki-bazi` 专家 本命判断（health）+ `liki-ziwei` 专家 本命判断（health）；读取 `liki-bazi/skills/bazi/wuxing-health.md` | 过旺 / 过弱五行与易病方向 |
| 2 | 限运触发 | 调 `liki-bazi` 专家 `十年期应期判断` + `liki-ziwei` 专家 `十年期应期判断`；读取合会冲宫与应期表 | 冲入宫位与触发年份 |
| 3 | 性质与走向 | 读取 `liki-bazi/skills/bazi/hehui.md` | 事件性质、结果走向、严重程度 |
| 4 | 具体细节 | 调 `liki-bazi` 专家 `本命判断（健康）` + `liki-ziwei` 专家 `本命判断（心理）`；用户问具体年份、应期或流年时发送 `应期判断` | 疾厄宫、福德宫、四化与流年信号 |

健康输出为倾向和关注方向；重大病灾需不同层证据闭环，并先列较轻替代解释。

## 边界条件

| 异常场景 | 处理方式 |
| --------- | --------- |
| 用户问"我有什么病"而非倾向 | 说明命理只能给倾向，不能诊断；给出对应脏腑建议 |
| 原局五行平衡无明显过旺过弱 | 输出"原局五行均衡，无明显薄弱脏腑" |
| 冲刑入多个宫位 | 按年>月>日>时优先级处理 |

## 输出模板

### 示例

```text
结论：五行木弱金旺，肝胆与筋骨为薄弱环节。
依据：木弱（jk_111）；金旺克木（jk_104）；命局无水通关。
建议：注意肝功能与筋骨健康；四季变化时关注呼吸道（金旺应肺）。
```
