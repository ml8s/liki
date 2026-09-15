---
name: app-personality
description: 性格描写 — 五行基础性格、十神修正、身强/身弱正反
依赖域: bazi,ziwei
---

# 性格分析


> 工具报文：只使用 `bazi/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
## 依赖的领域知识

[必读] - bazi: bazi/domains/bazi/wangshuai.md「身弱五行属性降级表」
[必读] - bazi: bazi/domains/bazi/shishen.md「十神组合场景化」

[按需] - ziwei: bazi/domains/ziwei/yingqi.md（仅问特定年份状态时读取）
[必读] - ziwei: bazi/domains/ziwei/gexing.md「紫微性格分析方法」

## 📖 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
|---|---|---|---|
| 1 | 基础性格 | 分别发送 `TOOLS.md §3.1` 的 `query.五行`、`query.旺衰`；读取 `bazi/domains/bazi/wangshuai.md` | 日主、身强弱、基础特征 |
| 2 | 组合修正 | 发送 `TOOLS.md §3.1 query.十神`；读取 `bazi/domains/bazi/shishen.md` | 十神组合与修正方向 |
| 3 | 紫微合参 | 分别发送 `TOOLS.md §3.1` 的 `query.命宫`、`query.身宫`、`query.福德`；特定年份发送 `§4.1 yearly_range.career` 或用户指定场景 | 主星、四化、执念与消耗点 |
| 4 | 外貌 / 体型 | 读取 `bazi/domains/ziwei/xiangmao.md`，与八字旺衰互证 | 体型倾向与证据强弱 |

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 用户问外貌 | 日主五行+旺衰定体型倾向（木高瘦、土敦实、金方正、水丰腴、火中等）|

## 📖 输出模板

| 输出 | 内容 |
|---|---|
| 结论 | 性格底色与外在气质 |
| 依据 | 日主五行、身强弱、十神组合、命身福德宫 |
| 外貌 / 体型 | 双盘证据并列后的倾向 |
| 建议 | 沟通与自我调节方向 |
