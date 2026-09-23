---
name: app-personality
description: 性格描写 — 五行基础性格、十神修正、身强/身弱正反
依赖域: bazi,ziwei
---

# 性格分析

> 工具报文：只使用 `natal/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: liki-bazi/skills/bazi/wangshuai.md「身弱五行属性降级表」
[必读] - bazi: liki-bazi/skills/bazi/shishen.md「十神组合场景化」

[按需] - ziwei: liki-ziwei/skills/ziwei/yingqi.md（仅问特定年份状态时读取）
[必读] - ziwei: liki-ziwei/skills/ziwei/gexing.md「紫微性格分析方法」

> **本卡的 `analyze_natal` / `analyze_periods` 调用均传 `topics=["personality"]`，仅保留性格域断语。**

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 基础性格 | 分别发送 `TOOLS.md §2` 的 `analyze_natal`、`analyze_natal`；读取 `liki-bazi/skills/bazi/wangshuai.md` | 日主、身强弱、基础特征 |
| 2 | 组合修正 | 发送 `TOOLS.md §2 analyze_natal`；读取 `liki-bazi/skills/bazi/shishen.md` | 十神组合与修正方向 |
| 3 | 紫微合参 | 分别发送 `TOOLS.md §2` 的 `analyze_natal(topics=["personality"])`、`analyze_natal(topics=["personality"])`、`analyze_natal(topics=["mental"])`；特定年份发送 `§4.1 analyze_periods` 或用户指定场景 | 主星、四化、执念与消耗点 |
| 4 | 外貌 / 体型 | 读取 `liki-ziwei/skills/ziwei/xiangmao.md`，与八字旺衰互证 | 体型倾向与证据强弱 |

## 边界条件

| 异常场景 | 处理方式 |
|---------|---------|
| 用户问外貌 | 日主五行+旺衰定体型倾向（木高瘦、土敦实、金方正、水丰腴、火中等）|

## 输出模板

### 示例

```text
结论：月令正印主内向沉稳，食伤透干添表达力，外冷内热。
八字：日主戊土，月令寅木藏甲木（七杀）；食神伤官双透。
紫微：命宫天相（庙），外圆内方，稳重中带变通。
依据：月令主面为偏印（xg_m03）；食伤为辅面增表达（xg_507）。
```
