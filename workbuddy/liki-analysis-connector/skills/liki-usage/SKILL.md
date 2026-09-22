---
name: liki-usage
display_name: Liki 命理工具使用
display_name_en: Liki Tool Usage
description: 调用 Liki 连接器的命理计算工具：八字、紫微、六爻、奇门、黄历、风水与起名。适用于排盘、看命、占卜、择日、起名等请求。
description_zh: 指导如何正确调用 Liki 连接器提供的八字、紫微、六爻、奇门、黄历、风水与起名 MCP 工具，包括参数要求、调用顺序、示例与错误恢复。
description_en: Guides how to correctly call the Liki connector's metaphysics MCP tools — BaZi, Zi Wei Dou Shu, Liu Yao, Qi Men, Huang Li, Feng Shui and naming — including parameters, call order, examples and error recovery.
version: 1.0.0
author: Q.Sui
---

# Liki 命理工具使用

> 连接器工具通过 MCP `tools/call` 调用，工具名以 `_` 分隔（如 `bazi_chart`）。所有结果由 Go 天文历算引擎确定性计算，断语来自规则真值表，可回溯。命理结论是传统文化视角的条件性解读，不构成医疗、法律或投资建议。

## 通用原则

1. **先排盘再分析**：分析类工具（`*_fullchart`、`*_liunian`、`*_liuyue`、`*_liuri`、`*_liushi`、`*_daxian`、`*_bond`、`liuyao_chart` 等）都需要 `chart` 或 `casting` 收据参数，必须先调用对应的排盘工具取得结果，再原样传入，不得手写或编造盘面。
2. **真太阳时优先**：涉及时间计算的工具建议先用 `tianwen_time`（参数 `time` 为 RFC3339，`longitude` 为出生地经度）取得真太阳时 `solar` 字段，再把该值传给排盘工具。城市经度可用 `city_coords` 查询。
3. **时间用 RFC3339**：所有时间参数格式为 `YYYY-MM-DDTHH:MM:SS+08:00`；日期格式为 `YYYY-MM-DD`。
4. **参数不足就问**：缺出生日期、时间、城市或性别时先向用户确认，补齐后再调用，不要凭猜测传参。
5. **错误即失败恢复**：工具返回 `isError` 时，错误消息会给出原因（如参数非法、图表缺失、digest 不匹配）。按消息修正后重试；校验类错误（digest 不符）说明盘面被改过，重新排盘。

## 时间与地点

| 工具 | 用途 | 关键参数 |
| --- | --- | --- |
| `tianwen_time` | 计算真太阳时，返回公历/真太阳/农历三套时间 | `time`(RFC3339 必填)、`longitude`(经度必填，北京≈116.4) |
| `city_coords` | 城市名查经纬度（Nominatim 服务，可能稍慢） | `city`(中英文均可) |
| `time_now` | 当前服务端时间，避免时间幻觉 | 无参数 |

## 八字（BaZi）

1. `bazi_chart`：排最小命盘。参数 `solar_time`（真太阳时，来自 `tianwen_time.solar`）、`gender`（male/female）。得到四柱、纳音、大运。
2. `bazi_fullchart`：补全十神/藏干/神煞/自坐/空亡/用神三派等。参数 `chart` = `bazi_chart` 的返回。
3. `bazi_bond`：八字合盘原始事实，不做合婚评级。参数 `a.chart`、`b.chart` 均为各自 `bazi_chart` 结果。
4. 流年/流月/流日/流时：`bazi_liunian`(year+chart)、`bazi_liuyue`(year+month+chart)、`bazi_liuri`(year+month+day+chart)、`bazi_liushi`(year+month+day+hour+chart)。
5. `bazi_xiaoyun`：小运列表。参数 `chart`、`count`(默认12)。

> 解读要点：流年工具返回十神、神煞、伏吟反吟与 `atomic_facts`；结合大运（`chart.da_yun`）与命局五行为用户做条件性解读，说明依据，不打包票。

## 紫微斗数（Zi Wei Dou Shu）

1. `ziwei_chart`：排紫微盘。参数 `lunar`{year,month,day,leap,shichen}、`gender`。返回十二宫星曜、亮度、四化。
2. `ziwei_fullchart`：扩展杂曜/长生/博士/小限/将前/岁前。参数 `chart`。
3. `ziwei_daxian`：十年大限。参数 `chart`。
4. 流年/流月/流日/流时：`ziwei_liunian`(lunar_year+chart)、`ziwei_liuyue`(target_lunar+chart)、`ziwei_liuri`(target_lunar+chart)、`ziwei_liushi`(target_lunar+shi_zhi+chart)。`target_lunar` 为完整农历日期 {year,month,day,leap}。
5. `ziwei_bond`：紫微合盘原始事实。参数 `a`、`b` 均为 `ziwei.chart` 完整对象。

## 六爻（Liu Yao）

1. `liuyao_qigua`：起卦。`mode` 可选 `auto`（安全随机六次）、`coins`（六组三枚正/反）、`yaos`（六个爻值 6-9）。返回可审计的 casting 收据。
2. `liuyao_chart`：装卦与分析。只接受 `liuyao_qigua` 返回的完整 `casting` 收据 + `solar_time`。返回纳甲、六亲、六兽、用神、旺衰、应期。

> 用神默认世爻，可通过 `yong_shen` 指定（父母/兄弟/官鬼/妻财/子孙/世爻/应爻）。问卦先明确所问之事与是否已有卦象。

## 奇门（Qi Men）

`qimen_chart`：奇门排盘。参数 `solar_time`(必填)、`scope`(hour/day/quarter/month/year)、`school`(zhuanpan/luoshu_feipan/mingfa_feipan/jinhan_yujing)、`dingju_method`、`yong_shen` 等。条件约束见工具 schema 的 `allOf`，不满足组合会返回参数错误。

> 事象路由由上层完成：拿到盘面后结合用神符号（门/星/神/干）与 `ying_qi` 窗口做策略性解读，说明方向与时机的条件性。

## 黄历（Huang Li）

`huangli_days`：查连续 N 天黄历。参数 `start_date`(YYYY-MM-DD)、`count`(默认3，最多30)、`event`(可选：wedding/move/opening 等15种事项)。给 `event` 时返回建除 suitability。

## 风水（Feng Shui）

1. `bazhai_chart`：八宅命卦 + 四吉四凶方 + 出生年紫白。参数 `birth_year`、`gender`。不需要出生时辰。
2. `bazhai_layout`：八宅门主灶配合。参数 `ming_gua`/`door_gua`/`master_gua`/`stove_gua`。
3. `xuankong_chart`：玄空飞星。参数 `period_date`(宅运起盘日期)、`zuo_shan`/`xiang_shan`(0-23，坐向差180°)。
4. `xuankong_liunian`：玄空流年飞星。参数 `chart`(可选，xuankong_chart 返回，含 digest)、`year`。

## 起名（Naming）

1. `qiming_surname`：外国人中文姓候选。参数 `source_surname`(罗马字姓)。
2. `qiming_char`：查单个汉字五行/笔画/部首/拼音。
3. `qiming_pick`：按五行取字池。参数 `wuxing1`、`count`(1单名/2双名)。
4. `qiming_compose`：组名。参数 `first`(首字数组)、`second`(次字数组，可选)、`max_names`。
5. `qiming_check`：名字评估。参数 `given_names`、`yongshen`/`xishen`/`jishen`(五行)。

> 起名流程：先确定用神（可结合八字 `bazi_fullchart` 的用神三派）→ 按五行取字 → 组名 → 评估。候选名数量与字池受控，需与用户确认偏好后缩小范围。

## 安全边界

- 出生信息只在当前会话使用，不索要真实姓名，不存储。
- 医疗、法律、金融、安全等现实话题：仍按流程计算，但结论标注"传统文化视角"，建议咨询相应专业人员。
- 未授权不读取任何个人文件或外部数据；所有计算只依赖用户主动提供的信息。