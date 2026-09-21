---
name: app-marriage
description: 婚姻分析 — 何时结婚、婚姻质量、感情走向
依赖域: bazi,ziwei
---

# 婚姻分析

> 工具报文：只使用 `natal/TOOLS.md`；`pan` / `snapshot` 等动态对象按“变量绑定”原样传回。
>
## 依赖的领域知识

[必读] - bazi: natal/domains/bazi/shishen.md「女命婚姻——官杀混杂判断」

- bazi: natal/domains/bazi/gongwei.md「宫位论」（按需——论宫位细节时读取）
[必读] - bazi: natal/domains/bazi/dayun.md「应期决策表」
- bazi: natal/domains/bazi/family.md「六亲——配偶」（按需——官杀线已覆盖配偶星主线）

[必读] - ziwei: natal/domains/ziwei/yingqi.md「婚姻紫微应期」

> **本卡所有 analyze_natal / analyze_periods 调用必须传 `topics=["marriage"]`**，只保留 婚姻 域断语，排除跨域噪声。

## 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
| --- | --- | --- | --- |
| 1 | 配偶星 | 男看财星、女看官杀；读取 `natal/domains/bazi/shishen.md` | 星名、清浊、取清状态 |
| 2 | 夫妻宫 | 发送 `TOOLS.md §2 analyze_natal`；读取 `natal/domains/bazi/gongwei.md` | 日支冲刑合害与化用 / 化忌 |
| 3 | 婚姻状态 | 分别发送 `TOOLS.md §2` 的 `analyze_natal`、`analyze_periods(time_scope.type="decade")`、`analyze_periods(time_scope.type="decade")` | 已婚 / 单身 / 离异 / 婚缘迟 |
| 4 | 应期 | 读取 `natal/domains/bazi/dayun.md`；发送 `TOOLS.md §3 analyze_periods` | 首选年、备选年、引动层 |
| 5 | 具体细节 | 发送 `TOOLS.md §3 analyze_periods` | 紫微夫妻宫、四化、桃花信号 |

## 边界条件

| 异常场景 | 处理方式 |
| --------- | --------- |
| 用户未婚但问离婚 | 先问是否已有稳定对象，有则分析当前关系，无则分析命局倾向 |
| 男命问感情但原局无财星 | 食伤为财源，查食伤状态（食伤生财为隐性妻星） |
| 女命问婚姻但原局无官杀 | 查财星（财生官杀为隐性夫星），或大运引动 |
| 已离婚问再婚 | 查七杀是否清透，大运有无正官/正财出现 |
| 用户只提供了一个人的八字但想看合盘 | 提示需要双方出生信息，引导到 natal/app/compatibility.md |

## 输出模板

### 示例

```text
结论：配偶星得地有根，夫妻宫安静，婚姻中稳。
依据：配偶星透干有根（hun_101）；夫妻宫无冲合刑害（hun_421）。
应期：大运配偶星窗口 2027-2031，流年丙午（2026）配偶星透干，首选 2026-2027。
建议：未验证时段标注置信度；确证前不写成必然。
```
