---
name: liki-liuyao
description: 六爻占卜模块（组件）。起卦、装卦、用神、月建日建、应期。
---

# 六爻占卜

## 知识索引

| 文件 | 功能 |
|------|------|
| domains/liuyao/duanyu/yongshen.md | 用神取用 |
| domains/liuyao/duanyu/yuejian.md | 月建日建 |
| domains/liuyao/duanyu/yingqi.md | 应期判断 |
| domains/liuyao/duanyu/jixiong.md | 吉凶判定（用神旺衰+动爻生克→吉凶） |

## 技术流程

1. **路由判断**：用户问吉凶时使用六爻。
2. **收集参数**：问题 + 出生信息（可选）。
3. **调用引擎**：liuyao.qigua 起卦 → liuyao.chart 装卦（chart 含用神状态/月破/日建/动爻/应期数据）。
4. **断卦（skill 自断）**：取用神（yongshen.md）→ 看月建日建旺衰（yuejian.md）→ 看动爻生克（jixiong.md）→ 判吉凶（jixiong.md）→ 断应期（yingqi.md）。
5. **解读**：按 jixiong.md 判定表输出吉凶结论，附依据链。

📖 搜索 domains/liuyao/duanyu/yongshen.md → 读取用神表
📖 搜索 domains/liuyao/duanyu/yuejian.md → 读取月建表
📖 搜索 domains/liuyao/duanyu/yingqi.md → 读取应期表
