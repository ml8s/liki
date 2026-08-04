---
name: liki-bazhai
description: 八宅风水模块（组件）。命卦、东四西四、门主灶分析。
---

# 八宅风水

## 知识索引

| 文件 | 功能 | 说明 |
|------|------|------|
| `bazhai.minggua` RPC | 命卦+东西四命 | 确定吉利方位 |
| domains/bazhai/duanyu/youxing.md | 游年九星+门主灶 | 吉凶应事+空间判断 |

## 技术流程

1. **收集参数**：出生年份+性别（定命卦），或完整出生信息（排盘）。
2. **确认 TimeSet**。
3. **调用引擎**：bazhai.minggua 命卦查询 → bazhai.chart/judgment 门主灶论断。
4. **解读**：调 `bazhai.minggua` 引擎命卦结果 + `domains/bazhai/duanyu/youxing.md` 断语 → 输出。

📖 搜索 `bazhai.minggua` RPC → 读取命卦表
📖 搜索 domains/bazhai/duanyu/youxing.md → 读取九星表

## 边界规则

- 只使用八宅自有体系，不以八字日主/用神解释风水
- 命卦（八宅）和日主（八字）是两个独立体系，各自独立使用
