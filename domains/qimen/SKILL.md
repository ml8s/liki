---
name: liki-qimen
description: 奇门遁甲模块（组件）。八门、九星、八神、用神、决策。
---

# 奇门遁甲

## 知识索引

| 文件 | 功能 |
|------|------|
| domains/qimen/duanyu/bamen.md | 八门吉凶 |
| domains/qimen/duanyu/jiuxing.md | 九星吉凶 |

## 技术流程

1. **路由判断**：用户问方向/时机时使用奇门。
2. **收集参数**：问题 + 时间 + 方位（可选）。
3. **调用引擎**：qimen.chart 排盘（chart 含八门/九星/八神/空亡/马星/门迫/天干关系/值符宫/值使宫/日干宫）。
4. **断事**：qimen.judgment（参数化——按事件类型取用神断主题宫/生克/格局/空亡马星影响）；或用 bamen.md/jiuxing.md 断八门九星吉凶。
   - chart 排盘固有：zhi_fu_xing_gong（值符宫）/ zhi_shi_men_gong（值使宫）/ ri_gan_gong（日干宫）直接读取
   - qimen.judgment 事件类型：general/career/wealth/relationship/study/travel/health/legal

📖 搜索 domains/qimen/duanyu/bamen.md → 读取八门表
📖 搜索 domains/qimen/duanyu/jiuxing.md → 读取九星表
