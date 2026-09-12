# Domain oracle fixtures

这些 JSON 是领域硬事实的独立数据集，不从实现表自动生成。新增或修改用例时必须先写明经典口径、流派与边界；禁止为了通过实现反向改 expected。

## 维护规则

1. fixture 与实现分开评审；不能由实现导出后直接覆盖。
2. 每个文件保留 `basis` / `sources` / `school`（如有流派差异）。
3. 修改 expected 必须给出经典依据或多个独立实现交叉验证。
4. 只放硬事实与条件候选，不放绝对吉凶承诺。
5. Go / Python 测试只读取本目录，不复制实现内部常量当 expected。

## 覆盖

| 文件 | 覆盖 |
|---|---|
| `bazi_core.json` | 十神 100、藏干 12、纳音 30 对、旬空 6、十二长生 120、十干禄 10、五虎遁 10 |
| `bazi_operator_lu_root.json` | 十干禄边界与禄根算子正反例 |
| `bazi_lu_roots.json` | engine 十干禄根输出：透干且禄位实见 / 禄位缺失边界 |
| `bazi_gender_shen_sha.json` | 勾绞煞与元辰的年干阴阳 × 男女双向口径 |
| `bazi_annual_shen_sha.json` | 流年病符 / 丧门 / 吊客 / 大耗的方向锚点 |
| `bazi_nayin_shen_sha.json` | 学堂 / 词馆的年命纳音同气正位与反例 |
| `bazi_classic_corrections.json` | 阴干羊刃与月刃格、天赦 / 四废四季边界、天罗地网年命纳音限定、#57 辛日申月真实命例 |
| `bazi_ten_god_states.json` | 十神显隐 / 通根 / 得令 / 组合旺弱、五行生克方向与大运通根边界 |
| `bazi_atomic_facts.json` | 官杀取清、财库、夫妻宫、日支神煞、柱位十神 / 长生 / 格神透干 / 柱刑原子事实 |
| `bazi_relation_groups.json` | 紧邻 / 隔位天干五合、地支六合、三合 / 三会 / 六冲 / 六害 / 三刑完整组与自刑成双边界 |
| `bazi_yongshen_structure.json` | 格局格神 / 克格神 / 生格神结构事实、扶抑强弱输入与关系投影、调候主辅神显隐与遭遇 |
| `bazi_liunian_atomic.json` | 流年生克、忌神、财坏印、三合 / 三会 / 三刑 / 半合原子事实 |
| `ziwei_core.json` | 四化 10、天魁 10、天钺 10、命主 12、身主 12、闰月边界 3 |
| `ziwei_pattern_semantics.json` | 六吉闭集、亮度分组、三方四正多四化与命宫 / 三方格局边界 |
| `liuyao_core.json` | 京房纳甲 8 宫、世应 8 序、可见 / 伏藏用神边界、真假空破与动爻生克冲突 |
| `liuyao_pattern_semantics.json` | 三墓来源、用神两现取舍、真假空破救应、卦变冲合 / 伏吟与三合去重边界 |
| `liuyao_dong_yao_priority.json` | 动爻直接作用与原忌神间接作用并列输出 |
| `../skills/liki-divination/tools/liuyao_timing_rules.json` | 六爻应期机制优先级、专题加权与视野策略 |
| `qimen_core.json` | 二十四节气 × 三元 72 局、值符 / 值使应期候选 |
| `qimen_specialized.json` | 庚格四柱遁甲、失物时干宫、天蓬 / 阴遁玄武、天网与地罗临时干宫 |
| `huangli_core.json` | 建除 12、十二月青龙起例 12、事项规则 10、日期锚点 |
| `fengshui_core.json` | 二十四山 24、八宅大游年 8、命卦锚点、流年飞星、三元九运、玄空四大局分类与位置事实 |
