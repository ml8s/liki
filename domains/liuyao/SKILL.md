---
name: liki-liuyao
description: 六爻占卜模块（组件）。起卦、装卦、用神、月建日建、应期。确定性计算走引擎，断语由前端查表+LLM。
---

# 六爻占卜

## 路由

- **何时使用**：用户问吉凶/事之成败/应期（「这事能成吗」「什么时候有结果」）
- 问方向/时机类问题走奇门

## 知识索引

### 📋 方法（引擎确定性计算）

| 文件 | 功能 | 说明 |
|------|------|------|
| `liuyao.qigua` RPC | 起卦（摇卦） | **可不依赖时间**（用户心念/随机起卦）；也可先问时间 |
| `liuyao.chart` RPC | 装卦：六亲/六神/世应 + 每爻旺衰/月破/动爻生克 + 应期 | `solar_time` + `yaos`（**6 爻数组，取值 6-9：6 老阴/7 少阳/8 少阴/9 老阳，6/9 发动**）；`yong_shen` 可选（默认世爻） |

### 📖 断语（符号→现实，LLM 查表翻译）

| 文件 | 用途 | 依据 |
|------|------|------|
| domains/liuyao/duanyu/yongshen.md | 占事 → 用神取用 | 《增删卜易》 |
| domains/liuyao/duanyu/yuejian.md | 月建日建旺衰 → 影响 | 《增删卜易》 |
| domains/liuyao/duanyu/jixiong.md | 用神状态×动爻关系 → 吉凶判定链 | 《增删卜易》 |
| domains/liuyao/duanyu/yingqi.md | 用神状态 → 应期 | 《增删卜易》 |
| domains/liuyao/duanyu/liushou.md | 六神断事（青龙喜/白虎凶伤…） | 《卜筮正宗》 |

## 技术流程

1. **路由判断**：用户问吉凶时使用六爻。
2. **收集参数**：问题 + 出生信息（可选）。
3. **起卦**：调 `liuyao.qigua` 摇卦 → 得 `yaos`（用户可要求用自己摇的卦）。
4. **装卦**：调 `liuyao.chart`（solar_time + yaos，可选 yong_shen）→ 得卦名/每爻状态/用神/应期。
5. **断卦（前端）**：
   - 按 `yongshen.md` 选定用神六亲 → 在 `lines[]` 中定位该爻（`liu_qin` 匹配）
   - 读该爻 `wang_shuai`/`yue_po`/`dong_self`/`dong_sheng`/`dong_ke`/`xun_kong`/`ri_chen_relations` → 按 `jixiong.md` 判定链组织断语（`chart.xun_kong` 为日柱旬空地支，`lines[].xun_kong` 为该爻是否值空）
   - 叠加六神：用神爻/动爻所临 `liu_shou` 按 `liushou.md` 断事之色彩（青龙喜/白虎凶伤/玄武暗昧…）
   - 应期按 `yingqi.md` 核对 chart 的 `ying_qi`
   - **不产出抽象评级档位**，用符号关系（用神旺衰+动爻生克）表述吉凶
6. **解读**：按 jixiong.md 输出吉凶结论，附依据链。

📖 搜索 domains/liuyao/duanyu/yongshen.md → 读取用神表
📖 搜索 domains/liuyao/duanyu/yuejian.md → 读取月建表
📖 搜索 domains/liuyao/duanyu/yingqi.md → 读取应期表

## 边界规则

- 起卦可用时间也可不用时间（心念/随机），**不强制用户提供出生时间**
- 用神选择是前端推理（按所问），chart 不因用神不同而重排
- 多动爻取最关键的一动；用神持世=近身（`shi_ying=="世"`），临应=在外
