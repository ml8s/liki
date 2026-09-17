---
name: app-health
description: 健康诊断 — 脏腑薄弱、易感疾病、健康建议
依赖域: bazi,ziwei
---

# 健康分析


> 工具报文：只使用 `bazi/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
## 依赖的领域知识

[必读] - bazi: bazi/domains/bazi/wuxing-health.md「五行所属」
- bazi: bazi/domains/bazi/hehui.md「冲宫位表」+「冲吉凶表」（按需——论冲害细节时读取）
- bazi: bazi/domains/bazi/dayun.md「应期决策表」（按需——应期走疾厄宫闭环）
- bazi: bazi/domains/bazi/tiaohou.md「调候用神」（按需——冬夏寒燥失衡时读取）

[必读] - ziwei: bazi/domains/ziwei/yingqi.md「健康紫微应期」
## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
|---|---|---|---|
| 1 | 脏腑倾向 | 分别发送 `TOOLS.md §3.1` 的 `query.五行`、`query.旺衰`、`query.合会`、`query.用神`；读取 `bazi/domains/bazi/wuxing-health.md` | 过旺 / 过弱五行与易病方向 |
| 2 | 限运触发 | 分别发送 `TOOLS.md §3.1` 的 `query.大运`、`query.大限`；读取合会冲宫与应期表 | 冲入宫位与触发年份 |
| 3 | 性质与走向 | 读取 `bazi/domains/bazi/hehui.md` | 事件性质、结果走向、严重程度 |
| 4 | 具体细节 | 分别发送 `TOOLS.md §3.1` 的 `query.疾厄`、`query.福德`；用户问具体年份、应期或流年时发送 `§4.1 yearly_range.health` | 疾厄宫、福德宫、四化与流年信号 |

健康输出为倾向和关注方向；重大病灾需不同层证据闭环，并先列较轻替代解释。

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 用户问"我有什么病"而非倾向 | 说明命理只能给倾向，不能诊断；给出对应脏腑建议 |
| 原局五行平衡无明显过旺过弱 | 输出"原局五行均衡，无明显薄弱脏腑" |
| 冲刑入多个宫位 | 按年>月>日>时优先级处理 |

## 输出模板

| 输出 | 内容 |
|---|---|
| 结论 | 脏腑倾向与关注方向 |
| 依据 | 五行过旺过弱、合会冲宫、疾厄 / 福德宫 |
| 应期 | 触发年份与引动因子 |
| 建议 | 生活方式提示；标注非医学诊断 |
