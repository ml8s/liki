---
name: liki-xuankong
description: 玄空风水模块（组件）。元运、飞星盘、旺山旺向、流年飞星。确定性计算走引擎，断语由前端查表+LLM。
---

# 玄空风水

## 路由

- **何时使用**：用户问流年风水/飞星/旺山旺向/房屋坐向吉凶（与八宅按问法分流：玄空主「元运飞星·流年」，八宅主「人宅相配·门主灶」）
- 命卦/门主灶类问题走八宅，不走玄空

## 知识索引

### 📋 方法（引擎确定性计算）

| 文件 | 功能 | 说明 |
|------|------|------|
| `xuankong.chart` RPC | 排盘：元运 + 山向飞星盘 + 四大局（旺山旺向/双星会坐/双星会向/上山下水）+ 收山出煞 | 输入坐向（0-23） |
| `xuankong.liunian` RPC | 流年飞星：chart（可选）+ year → 飞星盘 + 宅盘凶星落宫对照 | **触发条件**：用户问流年/某年风水才调 |

### 📖 断语（符号→现实，LLM 查表翻译）

| 文件 | 用途 | 依据 |
|------|------|------|
| domains/xuankong/duanyu/feixing.md | 玄空九星吉凶表 | 《沈氏玄空》 |
| domains/xuankong/duanyu/yuanyun.md | 三元九运表（当运旺衰） | 玄空起例 |

## 技术流程

1. **收集参数**：房屋坐向（`zuo_shan`/`xiang_shan`，0-23），时间（当前年份）。
2. **确认 TimeSet**。
3. **排盘**：调 `xuankong.chart`（solar_time + zuo_shan + xiang_shan）→ 得元运/山向星/四大局（wang_shan 旺山、wang_xiang 旺向、shan_xing 双星会坐、xiang_xing 双星会向、xia_shui 上山下水）/收山出煞（shou_shan 收山=坐宫山星当令、chu_sha=拨水入零堂零神到向，理气部分；完整"出煞"判定需实际峦头砂水，断语时勿当作已验砂水）。
4. **（按需）流年**：用户问流年/某年吉凶 → 调 `xuankong.liunian`（chart + year）→ 得流年飞星 + 凶星落宫对照；未问流年则跳过此步。
5. **断语**：读 `domains/xuankong/duanyu/feixing.md` + `yuanyun.md`，把 chart/liunian 的确定性结果翻译成断语（星→五行→吉凶→应事→方位建议），**不产出抽象评级档位**。
6. **输出**：按报告模板（结论先行 + 依据链）输出。

📖 搜索 domains/xuankong/duanyu/feixing.md → 读取九星表
📖 搜索 domains/xuankong/duanyu/yuanyun.md → 读取九运表

## 边界规则

- 只使用玄空自有体系，不以八字日主/用神解释风水
- **飞星九星（一白~九紫）与八宅游年九星（生气/天医/…）是两套独立命名体系，不混用**
- 流年飞星叠加必须用 `xuankong.liunian` 的确定性结果，不得自行推算入中星
