---
name: app-foreign
description: 外国人起中文名 — 英文姓转音近中国姓，并按八字用神与五行选字
依赖域: bazi,qiming
---

# 外国人起中文名

> 场景：英文起中文名。通用起名/改名 → `naming.md`；自选名字评估 → `selfcheck.md`。

## 依赖的领域知识

[必读] - bazi: `domains/bazi/yongshen.md`「用神方法论」+ `domains/bazi/calibration.md`「时辰判定」
[必读] - qiming: `domains/qiming/ziku.md`「字库与选字」

## 📖 流程

第0步：选姓
  从英文姓的读音选择音近中国姓，并说明音、义依据。
  输出：□ 英文姓____ 中国姓____

第1步：时辰判定 + 排盘
  有出生信息时调用 `bazi.chart` 和 `bazi.fullchart`；无出生信息时明确跳过五行评估。
  输出：□ 时辰____ 八字已排____

第2步：定用神/喜神/忌神
  有排盘时按 `domains/bazi/yongshen.md` 定夺；无排盘时询问用户期望五行，用户选择“不限”则合并五个 `qiming.pick` 单字池。
  输出：□ 用神____ 喜神____ 忌神____ 期望五行____

第3步：取字
  有用神时双字名默认「用+用」，候选不足回退「用+喜」；无排盘且用户选择五行时按该五行取字；选择“不限”时合并五个单字池。
  LLM 只在返回 `chars` 内过滤。
  输出：□ 过滤后 first____ 过滤后 second____

第4步：组名
  调用 `qiming.compose`，只传字数组。
  单名只传 `first`；双名传 `first` 和 `second`。
  LLM 删除谐音不雅、难读难写、与选姓重复的名字。
  输出：□ 候选名____

第5步：评估
  调用 `qiming.check`，`given_names` 不含姓。
  选定的中国姓仅用于最终展示与谐音判断，不传入 qiming.check。
  结合 `valid/errors/phonetic/wuxing` 与英文母语者的发音便利度终选。
  输出：□ 推荐名____ 拼音____ 寓意____ 音韵____

## 红线（强制）

- 候选字只能来自 `qiming.pick`。
- `qiming.compose` 只传字；不传五行、笔画或拼音属性。
- `qiming.check` 的 `given_names` 不含姓。
- 无出生信息时不臆造用神，评估只依据字库事实、音韵、字义和文化适配。
- 推荐依据只使用五行、字义、出处和音韵。

## 📖 输出模板

用户英文时中英双语输出；用户中文时按中文输出。

### 一、选姓
`Surname: 史 (Shǐ) — from "Smith", chosen for its sound.`

### 二、用神
有排盘时说明宜补五行；无排盘时说明未评估五行。

### 三、候选名字

| 中文名 | 拼音 | 五行 | 寓意/出处 | 音韵 |
|---|---|---|---|---|
| [名字] | [pinyin] | [五行] | [中文一句 + 英文一句] | [声调/谐音] |

### 四、推荐
`I recommend [name] ([pinyin]) — it is easy to pronounce and matches the favored elements.`
