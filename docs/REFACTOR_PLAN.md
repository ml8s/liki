# liki-bazi 分层激进重构计划（执行版）

> 目标：消除"生活域"在命理层的错位。app = 唯一用户价值层；csv-rule = 纯命理本体。
> 用户已定：**中文命理词作 rule**、**脚本自动迁移**、紫微按宫位、八字按命理层次。

## 一、最终中文 rule 全集

### 紫微（按宫位）12 + 1
`命宫` `官禄` `财帛` `疾厄` `夫妻` `子女` `迁移` `田宅` `父母` `兄弟` `仆役` `福德` ＋ `格局`(跨宫综合)

### 八字（按命理层次）
`十神` `格局` `旺衰` `用神` `大运` `合会` `神煞` `调候` `五行` `六亲` `出身` `外貌`

### 流年（保留 8，本期不改命名）
`yearly_*`（yearly_婚姻/事业/财运/健康/学业/六亲/子女 + yingqi）—— 流年域作**后续阶段**，避免本期范围爆炸。

## 二、现有 csv → 目标中文 rule 映射（脚本迁移基准）

### 紫微侧
现有 ziwei/*.csv 断语 → 按**因子列宫位前缀**自动归位到目标宫位 csv（已核：十二宫各 15-35 断语，分布均匀）：
- ziwei/career.csv → `官禄`（含我此前错补的 `命宫` 断语归 `命宫`）
- ziwei/wealth.csv → `财帛`
- ziwei/health.csv → `疾厄`
- ziwei/marriage.csv → `夫妻`
- ziwei/personality.csv → `命宫`
- ziwei/qianyi.csv → `迁移`
- ziwei/tianzhai.csv → `田宅`
- ziwei/zinv.csv → `子女`
- ziwei/family.csv → 拆 `父母`/`兄弟`/`仆役`
- ziwei/ziwei.csv（通用表, 37 行）→ 按因子宫位前缀分到各宫 + 跨宫/福德归 `福德`/`命宫`
- ziwei/yingqi.csv + ziwei/zhiye.csv 跨宫应期/职业 → 属"应期"层面，本期 `命宫`/`官禄` 或保留聚合
- ziwei/geju.csv → `格局`（跨宫，不按宫拆）＋ 各宫格局若纯单宫则归该宫

### 八字侧（按因子键语义→层次）
现有 bazi/*.csv → 按**因子键（十神/用神/格局等）**归位：
- `shensha.csv`→`神煞`；`geju.csv`→`格局`；`tiaohou.csv`→`调候`；`dayun.csv`→`大运`；`zuhe.csv`(合会冲刑部分)→`合会`；(十神组合)→`十神`
- `health.csv`(五行部分)→`五行`；(十神为用/为忌部分)→`十神`
- `wealth/career/study/marriage/personality/family/zinv/zhiye/waimao/chushen/qianyi/tianzhai` → 按断语主体十神语义归 `十神`/`六亲`/`旺衰`/`用神`/`外貌`/`出身`
  - 用神喜忌类 → `用神`；强弱 → `旺衰`；六亲宫位刑 → `六亲`；外貌 → `外貌`；出身 → `出身`

## 三、同步改造清单（P1 每步后跑 make check + pytest）

| 改造点 | 动作 |
|---|---|
| `duanyu.py` | `ALL_DUANYU_RULES`/`_NATAL_RULES`/`_YEARLY_RULES`/`_BAZI_ONLY/_ZIWEI_ONLY`/`_CURRENT_DAYUN_RULES` 改为中文命理域白名单 |
| `tools/skill-tools.json` | `query.rule.enum` 改为中文全集；`yearly_range.rules` 保留 |
| `app/*.md`（12 卡） | `query(rule=…)` 改按命理域拼接（如 wealth 卡 → 十神+用神+旺衰+大运 + 财帛）|
| 断语 id | **保留不变**（唯一性标签；check_docs 只查存在性） |
| 测试 | test_agent_cli/test_duanyu 等 rule 引用改中文名 |
| `docs/FACTOR_MODEL.md` | 只补因子归位说明；因子全集不变（439） |

## 四、分阶段执行

- **P0 定稿**：本文件获批（含中文全集）
- **P1 紫微迁移**：写迁移脚本，把紫微断语按宫位前缀重组到中文宫位 csv；删旧生活域 ziwei csv；改 duanyu/skill-tools.json/app/测试的紫微侧 rule 引用；跑 check+pytest
- **P2 八字迁移**：八字断语按因子键归位中文层次 csv；删旧生活域 bazi csv；改八字侧 rule 引用；跑 check+pytest
- **P3 清理定稿**：删别名/映射垫层、重审单侧域、domains/bazi 生活域 md 重组、app 依赖域改写
- **P4 全量验证**：make check + pytest + pre-push + 抽样端到端 + 计数核对（711 断语 / 439 因子 / id 不变）

## 五、风险与应对
- 八字"十神→生活域"语义本性 → 拆层次后 app 卡多域拼接，需 app 卡重写为聚合调用（这是最大改动）
- 断语 id 前缀是生活域缩写（cai_/hun_/shi_…）→ 保留 id 不变，仅更新 doc 引用说明
- 脚本迁移遗漏/错配 → 每批迁移后人工核对行数与宫位计数；check_schema 拦截孤儿/死列
- 测试大量引用旧 rule → 同步改测试，先改契约再迁移（P1/P2 顺序保证）

## 六、执行进展（截至当前）

### 已完成
1. **紫微按宫位迁移（P1a 完成）**：284 断语 → 13 中文宫位 csv（命宫/官禄/财帛/疾厄/夫妻/子女/迁移/田宅/父母/兄弟/仆役/福德/格局），id 无损（0 丢 0 多），make check 全绿。
2. **紫微合并 5 组重复**：迁移聚合同域后暴露的跨表重复（命宫星/四化同约束）已合并，断语 711→706。
3. **八字按层次迁移（P2a 完成）**：312 断语 → 12 中文层次 csv（十神/格局/旺衰/用神/大运/合会/神煞/调候/五行/六亲/出身/外貌），id 无损，make check 全绿。
4. **备份**：紫微 /tmp/ziwei_backup（20）、八字 /tmp/bz_backup（18），均可回退。

### 现状风险/待办
- ⚠️ **契约未跟上**：duanyu `ALL_DUANYU_RULES`/`_NATAL_RULES` 等仍为旧 rule（career/health...），skill-tools.json enum 未改；`query(rule)` 对紫微/八字新中文域会失败（2 测试挂，过渡态）。
- ⚠️ **暴露 18 组真重复**（迁移聚合后 check_schema 逮住）：同一命理信号在不同生活域曾各写一条，语义相同，待合并（如 sh_102≈xg_301 华盖、sh_103≈qy_101 驿马）。
- ⚠️ **待命理人工复核 50 条**：脚本归位偏差标注（婚姻断语含大运辅助归大运、部分用神断语归十神等），需命理审确认。
- ⚠️ **app 12 卡需聚合重写**：生活域不再是一种 rule，app 卡改为按多个命理域 query 聚合（如 wealth 卡 → 十神+用神+旺衰+大运+财帛）。
- ⚠️ **domains 需重组**：bazi 生活域 md（career/wealth/study/family/health）重组成命理层次方法论；与已数据化内容去重。

### 后续待执行
- P1b/P2b：duanyu 白名单改中文命理域 + skill-tools.json enum 改 + 测试引用改
- 合并 18 组重复
- P3：app 卡多域聚合 + domains 重组
- P4：全量验证（pre-push + 计数 + 端到端）
