---
name: liki-xuankong
description: 玄空风水模块（组件）。元运、九星飞布、旺山旺向。
---

# 玄空风水

## 知识索引

| 文件 | 功能 |
|------|------|
| knowledge/yuanyun.md | 三元九运表 |
| knowledge/feixing.md | 九星吉凶表 |

## 技术流程

1. 确定当前元运 → 调用 knowledge/yuanyun.md 查表
2. 调用引擎获取飞星盘数据
3. 解读每宫 → 调用 knowledge/feixing.md 匹配吉凶

📖 搜索 knowledge/yuanyun.md → 读取九运表
📖 搜索 knowledge/feixing.md → 读取九星表
