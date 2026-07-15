# liki.hk 品牌定位

版本：9.0 (中文版)
状态：生效
更新日期：2026-07-15

---

# 品牌身份

项目的公开品牌身份是：

> **liki.hk**

所有面向公众的渠道统一使用此身份，包括：

- 官网
- GitHub
- README
- 文档
- LLMS.txt
- Open Graph
- Schema.org
- AI 搜索
- 社交媒体

---

# 定位

项目的核心定位是：

> **liki.hk — AI命理Skills**

此定位是所有公开渠道的唯一来源。

---

# 规范定义

**liki.hk 提供 AI命理Skills，覆盖八字、紫微、起名、问卦、风水。**

当前包含四个 Skill：

- Mingli（命理）
- Qiming（起名）
- Wengua（问卦）
- Fengshui（风水）

未来可能增加新的 Skill，但不改变品牌身份或核心定位。

---

# 术语

## AI Skills

面向用户的命理功能模块。

- Mingli
- Qiming
- Wengua
- Fengshui

## Metaphysics Engine

底层无状态确定性计算引擎。

只负责精确计算，不做推理解读。

## LLM

推理层。

负责理解用户意图、选择 Skill、调用引擎、解读计算结果、生成结构化回复。

---

# 产品模型

```
用户
  │
  ▼
AI Skills（面向用户的功能模块）
  ├── Mingli
  ├── Qiming
  ├── Wengua
  ├── Fengshui
  └── ...
```

用户只跟 AI Skills 交互。AI Skills 是面向用户的产品。

---

# 架构

liki.hk 将确定性计算与 AI 推理分离。

```
                User
                  │
                  ▼
             AI Skills
                  │
                  ▼
        Metaphysics Engine
                  │
                  ▼
         Deterministic Results
                  │
                  ▼
         LLM Interpretation
                  │
                  ▼
          Structured Response
```

## Metaphysics Engine

引擎是 liki.hk 生态的独立组件，提供无状态确定性计算。

职责包括：

- 天文计算
- 历法转换
- 真太阳时校正
- 八字排盘
- 紫微斗数排盘
- 风水计算
- 黄历计算
- 其他规则计算

引擎特质：

- 无状态
- 确定性
- 结果可重现
- 只做计算
- 不做推理或解读
- 可独立开发和部署

AI Skills 通过定义好的接口与引擎交互，LLM 负责推理和解读。

## LLM

LLM 负责：

- 理解用户意图
- 收集必要参数
- 选择对应 AI Skill
- 调用 Metaphysics Engine
- 解读计算结果
- 生成结构化回复

此架构确保确定性计算始终准确、可重现、不受 LLM 幻觉影响。

---

# 呈现规范

所有公开渠道应一致呈现：

品牌身份：

```
liki.hk
```

核心定位：

```
liki.hk — AI命理Skills
```

规范定义：

```
liki.hk 提供 AI命理Skills，覆盖八字、紫微、起名、问卦、风水。
```

README、官网、文档、LLMS.txt、Schema.org 和社交媒体应使用统一的品牌身份、定位和定义。

---

# 仓库

GitHub 仓库：

```
ml8s/liki
```

安装命令：

```bash
npx skills add ml8s/liki
```

仓库名称和安装命令是技术标识符。

公开品牌身份始终为：

```
liki.hk
```

---

# 设计原则

## 一个品牌

```
liki.hk
```

## 一个定位

```
liki.hk — AI命理Skills
```

## 一个定义

```
liki.hk 提供 AI命理Skills，覆盖八字、紫微、起名、问卦、风水。
```

## 一个产品

用户只与 AI Skills 交互。

## 一个引擎

AI Skills 由 Metaphysics Engine 提供算力。

## 一个架构

确定性计算与 AI 推理分离。

---

# 愿景

通过可信的 AI 推动命理的理解与应用。

每一次解读都应建立在确定性的、可重现的计算基础之上。
