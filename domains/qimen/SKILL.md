---
name: liki-qimen
description: 奇门遁甲模块（组件）。排盘、八门、九星、八神、用神。确定性计算走引擎，断语由前端查表+LLM。
---

# 奇门遁甲

## 路由

- **何时使用**：用户问方向/时机/决策（「往哪个方向好」「现在该不该做」）
- 问吉凶成败/应期类问题走六爻

## 知识索引

### 📋 方法（引擎确定性计算）

| 文件 | 功能 | 说明 |
|------|------|------|
| `qimen.chart` RPC | 排盘：九宫三盘（天/人/神）+ 值符值使 + 日时干落宫/生克/空亡马星影响 + 格局 | 输入时间（kind 可选 shi/ri/yue/nian） |

### 📖 断语（符号→现实，LLM 查表翻译）

| 文件 | 用途 | 依据 |
|------|------|------|
| domains/qimen/duanyu/yongshen.md | 事件→用神取用（干/门/星）+ 十干克应 | 《烟波钓叟歌》 |
| domains/qimen/duanyu/bamen.md | 八门吉凶表 | 《烟波钓叟歌》 |
| domains/qimen/duanyu/jiuxing.md | 九星吉凶表 | 《烟波钓叟歌》 |
| domains/qimen/duanyu/bashen.md | 八神断事（六合婚姻/白虎凶伤…） | 《烟波钓叟歌》 |

## 技术流程

1. **路由判断**：用户问方向/时机时使用奇门。
2. **收集参数**：问题 + 时间 + 方位（可选）。
3. **确认 TimeSet**：时家奇门按**问事时刻**定局（`kind` 默认 shi）——确认时刻准确（时差校正、干支无误），时辰错则局数错。
4. **排盘**：调 `qimen.chart`（solar_time，kind 默认 shi）→ 得九宫/八门/九星/八神/日时干落宫/生克/空亡马星影响/格局。
5. **断事（前端）**：
   - 定用神：按所问查 `yongshen.md`「事件→用神」表（事业→开门、求财→戊+生门、婚姻→乙/庚、健康→天芮+死门、诉讼→伤门、文书→景门等），再在 `pan.gong_wei[].tian_pan_gan` 中找用神干落宫
   - 读该宫的门/星/神 + `ri_shi_sheng_ke`/`kong_wang_affected`/`ma_xing_affected` → 按 `bamen.md`/`jiuxing.md`/`bashen.md` 组织断语
   - 格局/十干克应/应期**直接读 chart 的 `patterns`/`gan_interaction`/`men_interaction`/`ying_qi`**（引擎已含名称+含义+吉凶，无需查表）
   - **不产出抽象评级档位**，用符号关系（门星神吉凶+生克+空亡马星）表述
6. **输出**：按报告模板（结论先行 + 依据链）输出。

📖 搜索 domains/qimen/duanyu/yongshen.md → 读取用神取用表
📖 搜索 domains/qimen/duanyu/bamen.md → 读取八门表
📖 搜索 domains/qimen/duanyu/jiuxing.md → 读取九星表

## 边界规则

- 值符宫/值使宫/日干宫为排盘固有（chart 直接读取），不用神重推
- 空亡/马星影响为确定性派生（chart 已给），前端只翻译
- 时家（shi）为默认；问长期事可考虑日家/月家/年家（kind）
