---
name: liki-lifebook
description: liki.hk 生命之书 — AI 综合命理报告。命理结论为传统文化视角，仅供参考，不构成专业建议。
---

version: 1.0.0

# liki.hk 生命之书

## 流程

1. 读取 `generate.md`，按指引生成 JSON 报告
2. 读取 `review.md`，审查已生成的报告
3. 如不通过（pass: false），读取 `revise.md` 并附上审查反馈，修正后重复 2-3 最多 3 轮
4. 输出最终 JSON
