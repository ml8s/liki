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
3. **调用引擎**：qimen.chart 排盘，qimen.judgment 断事。
4. **解读**：调 bamen/jiuxing 匹配八门九星。

📖 搜索 domains/qimen/duanyu/bamen.md → 读取八门表
📖 搜索 domains/qimen/duanyu/jiuxing.md → 读取九星表
