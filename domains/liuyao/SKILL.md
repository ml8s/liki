---
name: liki-liuyao
description: 六爻占卜模块（组件）。起卦、装卦、用神、月建日建、应期。
---

# 六爻占卜

## 知识索引

| 文件 | 功能 |
|------|------|
| domains/liuyao/yongshen.md | 用神取用 |
| domains/liuyao/yuejian.md | 月建日建 |
| domains/liuyao/yingqi.md | 应期判断 |

## 技术流程

1. **路由判断**：用户问吉凶时使用六爻。
2. **收集参数**：问题 + 出生信息（可选）。
3. **调用引擎**：liuyao.qigua 起卦 → liuyao.chart 装卦 → liuyao.judgment 断卦。
4. **解读**：调 yongshen/yuejian/yingqi 匹配用神、月建、应期。

📖 搜索 domains/liuyao/yongshen.md → 读取用神表
📖 搜索 domains/liuyao/yuejian.md → 读取月建表
📖 搜索 domains/liuyao/yingqi.md → 读取应期表
