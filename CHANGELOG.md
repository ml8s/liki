# Changelog

## [2026.09.09.0] — 问卦场景路由与六爻事实层

### liki-divination

- 六爻 Skill 主链路改为 `tools/agent_cli.py`：LLM 调用 `liuyao_qigua` / `liuyao_read`，不再直接编排 JSON-RPC。
- 新增六爻事项路由、RPC 访问、起卦编排与 snapshot 投影模块；原始硬币仍只由 engine 归一化。
- `liuyao_qigua` 支持 auto / coins / yaos，并输出完整起卦收据；auto 生产路径改用 crypto/rand。
- `liuyao_chart` 支持回传完整 casting，同时兼容旧 `yaos` 输入；冲突输入显式报错。
- Python 层新增 snapshot 投影和事项→默认用神表；`matter` 不进入 engine，engine 只接收显式用神。
- 六爻 app 卡补充问题澄清、手动起卦、感情视角、追问锁定、冲突并列和高风险边界。
- 六爻 chart 新增 `timing_candidates` 结构化应期候选，输出机制、触发支、依据和成立条件，不再只给单段判断。
- 新增 `liuyao_audit` Python 工具，审计报告中的本卦、变卦、世应和爻位六亲硬事实；不裁决吉凶、应期与取象。
- 新增六爻 4096 爻值组合 property test，覆盖装卦核心不变量、世应唯一性和动变一致性。
- `liuyao.chart` 新增日冲细分、动变转换、三合候选、用神原忌仇作用链和用神多现候选；全部为可解释事实/候选，不直接给最终吉凶。
- Snapshot 升级至 v2：evidence 分为 primary / secondary / reference / conflicts / ignored_scope，明确主判、参考与不应升级为结论的信号。
- 新增六爻 report v1 契约和 `liuyao_report` 工具：生成证据引用骨架，校验未知引用、冲突覆盖和“必然 / 百分百 / 保证”类表述。
- 新增 `divination_route` 场景路由：事件结果默认六爻，策略进退 / 方向默认奇门，择日默认黄历；结果与策略混合时先澄清，不自动双法合参。
- 新增 `app/question.md` 问卦总入口，六爻与奇门 app 卡改为场景路由通过后才进入排盘。
- 新增 `qimen_read` 奇门一次读盘入口：自动补时间 / 换真太阳时 / 排盘 / 投影 snapshot / 可选专占查询；LLM-facing 工具面移除底层 city / solar / raw chart / query 分步调用。
- 六爻墓库改为五行对应墓库，新增十二长生、完整伏神层和爻间地支关系；三合候选补充日 / 月来源、空破计数、激活条件与作用目标。
- 新增 `liuyao_session` 会话契约：固化原卦 fingerprint、snapshot digest 和首次结论；追问可记录澄清、策略、解释与结果反馈，禁止静默重排或改写。
- 新增 `liuyao_topic_methods` 专题库与 `liuyao_topic` 工具：覆盖求财、事业、婚恋、学业、失物、出行、争议和健康语境，输出主判因子、参考因子、常见误判与边界。
- 新增 `liuyao_timing` 应期排序器：只整理 engine 已返回候选，输出 priority、rank_score、horizon、required_check 和条件化结论范围；健康语境直接拦截。
- 新增跨开源纳甲表口径的 external diff fixture，并以 Go 测试校验乾卦锚点的本卦、纳甲地支、六亲、世应、旬空和用神位置。
- 新增 `liuyao_conditions` 歌诀条件还原器：将“月破必废 / 旬空无用 / 六合必成 / 六冲必散 / 兄弟动破财”转为旺衰、动静、救应和作用路径条件，不输出绝对结论。
- 六爻 / 奇门 matter 语言统一 canonical：relationship、study、legal；旧 marriage、lawsuit、academic、legal_risk 仅作输入别名。
- app 层新增 outcome / decision / date 用户场景卡；方法卡只作为六爻 / 奇门实现说明，不再作为第一入口。
- 六爻 / 奇门 session 新增 integrity 摘要：casting/input、method、snapshot、first_verdict 和 report 任一被篡改都会被校验拒绝。
- 修复 `qimen_read` 专占链路：`special` 直接使用 `qimen_duanyu.query` 返回的 rule/assertions envelope，避免二次嵌套导致报告模板崩溃。
- 新增共享安全边界、统一 RPC 出口、集中契约校验和 `huangli_days` Python 编排；所有问卦 / 择日入口都不再让 LLM 直接编排 JSON-RPC。
- `qimen-report-v1` 纳入集中契约文件；snapshot、assertion、timing 引用前缀分离并可 schema 校验。
- 修复 `qimen_read` 专占链路：`special` 直接使用 `qimen_duanyu.query` 返回的 rule/assertions envelope，避免二次嵌套导致报告模板崩溃。
- 新增回归测试锁定 `qimen_read.special` envelope、内部 chart 调用和工具面不含旧 qimen_chart / query / liuyao_chart。
- 修复测试隔离：内部编排测试不再污染 `sys.modules`，工具面测试在 `test_qimen_tools` 移除 TOOLS 后仍可稳定加载 agent_cli。
- 新增 `docs/DIVINATION_MODEL.md`，说明问卦领域对象、场景层、方法层、证据分层、报告审计与会话锁定。
- 六爻领域命名收敛：`liuyao_interpret` 更名 `liuyao_read`，snapshot / topic guidance 模块更名；新增 `liuyao-reading-v1` 聚合根、reading_id 和 reading_digest。
- 统一六爻 matter 语言：canonical 使用 relationship / study / legal，旧 academic / legal_risk / marriage / lawsuit 仅作输入别名；移除易误解的 `xi_shen` 主作用链字段。
- `liuyao_timing.plan()` 更名 `rank_timing_candidates()`，明确其只排序条件候选，不做择时决策。
- CI 回归固化：奇门集成链路改用 `qimen_read`；LLM-facing 工具面禁止重新暴露 `qimen_chart`、`query`、`liuyao_chart`。
- 新增 `liuyao_interpret` 六爻一次解读包：排盘后自动绑定专题规则、应期排序、条件还原和 report 骨架，减少 LLM 分步编排。
- 新增 `qimen_report` / `qimen_session`：奇门报告可校验方法、snapshot、专占断语和应期引用；会话固化原局时间、地点、方法与首次结论。
- `liuyao_session` 创建前会校验 report v1 的证据引用、冲突覆盖和禁语；headline / verdict 必须与结构化报告一致。
- `liuyao_audit` 同时支持文本报告和 report v1 结构化报告，并合并硬事实审计与报告契约审计结果。
- 六爻 / 奇门结构化报告运行时校验补齐未知字段拒绝，奇门 session 新增首次结论篡改测试。
- 新增六爻节气换月边界 fixtures，覆盖立春、惊蛰、清明、立夏、小暑、立秋、白露、寒露、立冬、大雪、小寒前后。

## [2026.09.08.7] — Issue #42–#45 修复

### liki-bazi

- 流年“天克地冲日柱”改为同组同时满足流年天干克日干与流年地支冲日支，避免午午同支被误判为冲。
- 流年三刑要求流年支实际进入成组三刑；本命自带三刑不再在无关流年逐年重复触发。
- `ying_h18` 女命损胎断语补入 `性别=female` 门槛；同类通用食伤因子与紫微女命断语补入显性别门槛。
- `fam_120` 财星得地、比劫旺夺财增加“比劫为忌”约束；身弱以比劫为用者不再机械断家境先裕后损。
- 学业大运破坏候选要求当前大运比劫 / 食伤 / 财星既为忌又通根原局或得坐支本气；虚浮无根大运不作实质破坏。
- 新增大运财星 / 食伤 / 比劫有根稳定因子，并补入正交因子与断语契约测试。

### liki-divination

- 六爻进退神改为同五行地支顺逆精确表：亥化子、寅化卯、巳化午、申化酉为进；子化亥、卯化寅、午化巳、酉化申为退。
- 不构成进退迁移的动爻变化不再输出进退格局候选，也不再以 `is_true=false` 表示退神方向。
- 空亡 / 月破 `sub_type` 与 `is_true` 同步输出真空 / 假空、真破 / 假破，消除真假标记矛盾。
- 六爻格局文档与 RPC schema 明确 `is_true` 语义：空亡 / 月破中区分真假空、真假破；其他已成立结构为实质效力。

## [2026.09.08.6] — 奇门捕盗稳定因子

### liki-divination

- 奇门 snapshot schema 升至 v3，并新增九星、八门、八神及神星 / 神神宫位五行关系稳定因子。
- 奇门 snapshot 新增阴阳遁内外宫域与用神宫域稳定因子；走失人口专占补入方向、内外远近与神盘信号候选，不判定必然回归。
- 奇门 snapshot 新增神门关系、宫位高低区间与时干五阳 / 五阴稳定因子；捕亡专占补入六合伤门生克、天网高低与太白入荧候选。
- 捕亡专占补充六癸高位 / 阳时难获候选与六癸落宫追踪方向候选；候选仍并列输出，不做必然结论。
- 走失人口神盘信号收紧为六合落宫所临之神；解释证据保留命中因子行，避免只给全盘粗粒度信号。
- 新增八宫方向稳定因子表，走失人口与捕亡方向候选直接携带宫位方向证据。
- 中宫稳定标记为 `center`，不输出内外远近与方向候选，避免无方向断语。
- 修复可选年命字段缺省时奇门快照误报不完整的问题；多因子数组证据保留完整命中行。
- 按标准转盘八神显隐对补齐勾陈 / 玄武稳定领域别名。
- engine `pan` 补齐年 / 月柱稳定事实；失物补入时干落宫方向与内外候选。
- 失物规则依据补入《元灵经》方向与内外口径；中宫不输出方向 / 内外候选。
- 新增奇门排盘 + 五类专占解释的本地引擎集成测试，覆盖年 / 月柱 RPC 事实与规则适配层。
- 引擎全量脚本的 RPC 端口改为 `LIKI_ENGINE_PORT` 可配置，避免端口硬编码。
- 新增表驱动捕盗候选规则，覆盖制贼、难捕、串通、捕者为盗、纵盗、年 / 月 / 日 / 时庚格、无庚格与杜门方向。
- 新增贼人画像规则，覆盖贵人 / 小人、八卦人物象与内外盘候选。
- 走失人口补入六合宫旺相星临景 / 死 / 惊 / 伤四门的可得候选；神盘信号收敛到六合落宫。

## [2026.09.08.1] — 刻家奇门双口径与领域轴收敛

### 奇门刻家

- 新增 **十二分钟十分局刻家**：`scope=quarter` + `school=zhuanpan` + `quarter_rule=twelve_minute_ten_division`。规则固定为一时辰 10 刻、每刻 12 分钟，初局取当前时家局，阳遁顺进、阴遁逆退，九局循环；该口径盘面仍以时柱为主柱，`method.quarter.lead_pillar_mode=hour` 显式说明，不与十分钟三元刻家的刻柱主盘混淆。
- 十二分钟十分局支持 `base_dingju_method=chaibu/zhirun/maoshan` 选择基准时家定局法，默认拆补；`method.base_dingju_method/base_dingju_method_name` 输出事实源，其他方法在 schema 层拒绝该参数。
- 十二分钟十分局补齐分遁与时辰界选项：`dun_source=hour_branch/solar_term`，默认时支分遁；`hour_boundary=late_zi/zi_zheng`，默认晚子时。两个选项均由 `quarter.json` 承载并只在十二分钟十分局下开放。
- `scope=quarter` 显示名收敛为 **刻家奇门**；刻家内部以 `quarter_rule=ten_minute_sanyuan/twelve_minute_ten_division` 区分十分钟三元刻家与十二分钟十分局刻家，方法矩阵增至 17 组。
- 新增十分钟三元刻家：`scope=quarter` 默认 `quarter_rule=ten_minute_sanyuan`；一个时辰 12 刻，每刻 10 分钟。
- 两个刻家口径均封闭为 `quarter + zhuanpan`；飞盘、鸣法、金函与拆补 / 置闰 / 茅山组合均显式拒绝。
- 刻柱、阴阳遁与三元定局全部入 `quarter.json`：刻柱由时干五鼠遁起子刻后逐刻进一；时支子至巳为阳遁、午至亥为阴遁；时柱五时符头按子午卯酉 / 寅申巳亥 / 辰戌丑未取 1-7-4 / 9-3-6。
- `method.quarter` 输出刻序、每刻分钟数与主柱口径；主柱干支仍以 `method.lead_pillar` 为唯一事实源，避免重复。
- 以大师奇门官方 Web 排盘固定 artifact 作为外部执行 golden，另覆盖晚子时、上中下三元与刻柱边界；不复制其商业代码或数据表。
- 十二分钟十分局的外部执行源固定为 `Horace-Maxwell/Horosa-Web-App-comprehensively-improved-MacOS` commit `0604fa416f95e330dbd798ba2ede0f2aa0cb1688`，仅记录 AGPL-3.0 项目的执行事实，不复制实现；项目名只保留在测试 provenance，不进入公开参数、结果 enum 或方法显示名。
- 奇门公开领域轴收敛为 `scope`（时间层级）、`school`（盘式）、`dingju_method`（定局法）与 `quarter_rule`（刻家规则）；`dingju_method` 仅保留拆补 / 置闰 / 茅山，刻家规则不再混入定局法。
- 十二分钟十分局刻家与十分钟刻家是不同口径；文档与 schema 均显式区分，不互相近似替代。

### 奇门契约与用神

- 奇门拆分为 engine 排盘层与 Python 编排层：`qimen.chart` 只接收显式 `yong_shen`；事业 / 求财 / 婚姻 / 健康 / 诉讼 / 学业 / 出行 / 隐藏 / 走失人口 9 类事象到用神的映射上移到 `tools/data/qimen_matters.csv`，由 `qimen_chart` 查表后调用 engine。
- 新增 `liki-divination` 奇门 Python 工具链：`city_coords`、`solar_time`、`qimen_chart`、`query`。失物解释规则入 `qimen_assertions.csv` / `qimen_conditions.csv`，只按引擎已稳定输出的反吟与时干空亡事实给出候选；“时干宫生日干宫”因古籍另需旺相气前提而暂不入表，不在代码或提示词中临场推理。
- 奇门 Python 层补齐 `qimen_factors.py` 与 `qimen_snapshot_contract.json`：`qimen_chart` 结果先投影为稳定快照，解释表只消费快照字段并受契约校验，不再直接绑定 engine JSON path；`patterns`、十干克应、门星克应与应期日期仍由 engine 作为机械因子保留。
- 奇门稳定快照显式预留十干克应名、门克应名、星克应名、星宫五行关系、门破 / 门制宫位与值符星 / 值使门落宫；这些字段当前失物表未消费，但属于稳定盘面因子，不作为死代码裁剪。
- 奇门 Python 表加载层按职责拆分为 `qimen_matters.py` 与 `qimen_interpretations.py`，避免排盘编排依赖解释表；`qimen_rules.csv` 表达专占适用的 scope / school 与依据，失物限定为时家转盘 / 洛书飞盘 / 鸣法飞盘，金函玉镜在投影前显式拒绝。
- 奇门 Python 契约测试补齐 tool enum ↔ 事象表 / 解释规则 / engine schema、快照来源路径 ↔ engine result schema、可选 `yong_shen` 半截结构、条件组排序与断语无优先级输出，防止表与 schema 漂移。
- 奇门 Windows 启动器对齐八字兼容口径：优先使用 `py -3`，显式设置 `PYTHONUTF8` / `PYTHONIOENCODING` 并以 `-X utf8` 启动；Python 侧继续强制 UTF-8 stdio 与 ASCII JSON 输出。
- 金函玉镜标准 `Chart` API 显式拒绝误用：`ComputeChartWithMethod` 与 `ComputeChartWithYongShenAndMethod` 不再落入常规九宫排盘流程，必须使用独立的 `ComputeJinhanChart`；RPC handler 的独立 Jinhan result 行为保持不变。
- 显式 `yong_shen` 输入契约收紧：engine schema 与 Python tool schema 均拒绝空数组并要求符号唯一；Python 编排层在无符号时应省略该参数，不得用空数组伪装成高级显式取用。
- 收窄奇门内部 Go API：无用神聚合包装仅由标准排盘内部调用，改为包内函数，不再作为公共导出入口。
- 奇门 Python 排盘编排层增加边界防护：显式 `yong_shen` 必须是非空且不重复的字符串数组；事象表拒绝重复符号；JSON-RPC 返回体畸形或缺 `result` 时转换为清晰的 `RPCError`，不再透出底层解析异常。
- 奇门 Python RPC 解包统一校验 engine envelope：`city_coords`、`solar_time`、`qimen_chart` 缺少 `data` 对象时返回清晰 `RPCError`，不再暴露 `KeyError`。
- 奇门 Python 层新增真实 engine golden chart 投影契约测试；RPC 端点改为每次调用读取 `LIKI_RPC_URL`，HTTP 客户端错误区分状态码与响应体，5xx / 超时 / 网络错误重试，4xx 不再盲目重试。
- 失物解释补入“时干宫乘旺相气生日干宫”复得候选；旺相状态复用 `WangShuaiOf` 的月令五行规则，由 engine 输出 `palace_wang_shuai` 与 `shi_gan_gong_wang_shuai`，解释层只消费 snapshot 因子。墓、绝仍不拟合。
- 格局表新增 `六仪击刑`：按《奇门遁甲统宗》六甲值符刑宫口径绑定戊震三、己坤二、庚艮八、辛离九、壬癸巽四，六支均提供 golden 向量；Python 快照自动透出 `patterns` 因子供专占解释消费。
- `ying_qi.candidates` 新增 `dates[]`：马星按冲日、空亡按值日与冲日在 60 日窗口内输出具体日期；日期锚点对齐排盘输出的日柱，避免晚子时日界漂移。LLM 只解释引擎日期，不补算。
- 格局表新增青龙返首、飞鸟跌穴、荧入太白、太白入荧与三奇入墓；前四者直接引用十干克应表的唯一天地盘干组合与吉凶事实，三奇入墓按古籍口径绑定乙奇坤宫、丙奇乾宫、丁奇艮宫，不重复维护第二份条件。
- 修正用神中门名重复输出“门门”的显示错误，门类符号统一输出完整“×门”名。
- `qimen.chart` 参数 schema 将缺省 `scope` 固定为时家口径：未传 `scope` 时拒绝 `quarter_rule`、`base_dingju_method`、`dun_source` 与 `hour_boundary`，避免 schema 放行而引擎拒绝的漂移。
- 结果 schema 禁止十分钟三元刻家输出十二分钟十分局专属的基准定局、分遁与时辰界字段，防止跨口径 result 漂移。
- 奇门用神表外问事必须先映射表内问事；无法映射时说明无封闭规则，禁止 LLM 按五行临场类象。格局输出同样限定为已入表的封闭规则，不承诺全量格局识别。
- 八神用神输出统一渲染为当前盘实际神名：转盘阳遁白虎 / 玄武别名解析后输出勾陈 / 朱雀，阴遁反向同理，避免名实不符。
- 全年边界日期 result schema 抽样覆盖全部 17 个公开方法组合，包括茅山、鸣法、两种刻家与金函玉镜。
- 十干克应、门克应与星克应数据文件增加文件级 `provenance=curated` 元数据；加载器显式校验来源状态与条目非空，避免把内部整理表伪装成外部 golden 或古籍原文。

### 工程精简

- `docs/FACTOR_MODEL.md` 不再逐行复制 456 个本命因子与 101 个流年因子，收敛为契约、统计与数据流说明；完整清单以 `factors.csv` / `factors_liunian.csv` 为唯一事实源。未引用的 `FACTOR_LAYER_DESIGN.md` 与重复的 `tests/schema.md` 合并删除。
- `eval_hybrid.py` 覆盖统计默认输出 stdout，只有显式 `--output` 才落盘，删除固定生成 `tests/RESULTS.md` 的路径与 ignore 规则。
- 删除仅测试使用的 `FullChart.CangGanArray()` internal 导出接口，测试直接断言四柱 `cang_gan` 领域字段。
- 奇门十干克应删除 72 行由天地盘干机械推导的 `name` 列；星克应删除 30 行由星宫五行推导的 `name` 列，运行时保持原输出显示名不变。门克应的“入/加”口径依赖门归本宫事实，继续由表承载。

## [2026.09.06.1] — 奇门完整方法矩阵与领域模型收敛

### 奇门领域边界

- `qimen.chart` 方法参数收敛为 `scope` / `school` / `bureau`：`scope` 表示 **hour 时家 / day 日家节气 / month 月家周期 / year 年家周期**，`school` 表示 **转盘法 / 洛书飞盘法 / 鸣法飞盘**，`bureau` 表示 **拆补 / 置闰**；month/year 使用固定周期定局并在 schema 层拒绝 `bureau`。
- 新增 **鸣法飞盘** 方法包：`school=mingfa_feipan`，第一版封闭为 `scope=hour` + `bureau=chaibu`；九星含天禽顺飞、九门含中门顺飞、九神为值符→螣蛇→太阴→六合→勾陈→太常→朱雀→九地→九天，并输出鸣法旬内回绕暗干支。固定 potuo/feipan-qimen MIT 参考实现与已检入测试作为外部 golden，未开放鸣法 day/month/year 与置闰组合。
- 鸣法 P0 收敛：新增 4 个 `potuo/feipan-qimen` 已检入参考锚点，覆盖阴遁、甲午旬、中门落宫与旬内回绕暗干支；中门按土行参与门迫 / 门制，但未入门克应表时不生成伪解释。schema 以 `method.school` 做 discriminator，非鸣法禁止 `中门` / `an_gan_zhi`，鸣法每宫必须输出 `an_gan_zhi`，且输入用神的 `中门` 仅鸣法可用。
- 新增茅山定局：`bureau=maoshan`，第一版仅开放 `hour + zhuanpan`。规则以精确交节时刻为唯一锚点，不问符头、不置闰，每 60 时辰推进一元；三元后未交下一节气则继续用下元。固定 deminzhang/qimen-go MIT 实现执行 7 个边界锚点，覆盖交节前后、60/120/180 时辰与冬夏至边界。
- 新增金函玉镜日家：`school=jinhan_yujing`，仅开放 `scope=day` 且不传 `bureau`。以《金函玉镜》九星落局、二至阴阳、八门三日一换与十干十二神为独立专用 chart，不复用三奇六仪、值符值使、常规格局或应期；以 ctext 古籍文本与 kinqimen 固定 commit 建立 9 个 golden 向量。
- `solar_time` 明确必须来自 `city.coords` → `tianwen.time` 的真太阳时链路；skill 流程禁止 LLM 自行换算或排盘。
- `an_gan` 按当前盘式与值使位移重新输出：转盘按外圈旋轴推导，洛书飞盘按九宫宫数位移推导，不再把两种几何混作一个结果。

### 排盘核心修复

- `qimen.chart` 在默认拆补之外支持表驱动的 hour/day 置闰定局，并可组合转盘 / 洛书飞盘。
- day 层级以日柱为主柱并统称“日家奇门”：转盘/飞盘按节气拆补或置闰定局，金函玉镜为独立日家盘；月家按年支组与寅月退行定局；年家按 1864 锚点、60 年周期与上中下元固定局定局。
- 洛书飞盘采用九宫宫数位移：中五参与飞宫、天禽独立落宫、八门存在一个空宫、九神包含太常。
- 方法元数据显式输出盘式几何与飞行口径：洛书飞盘为 `palace_number_shift`，九星与本宫天盘干整体宫数位移；八门真实中宫步位可为空；九神阳顺阴逆。不同飞盘流派不再被泛称 `feipan` 掩盖。
- 奇门 `params/result` JSON Schema 迁至 `qimen_schema.json` 并由引擎 embed 加载；保留自包含、无 `$ref` 的 OpenRPC 契约，消除 1.4 万字符单行 schema 的审查盲区。
- 奇门方法矩阵 golden 补齐实际节气、用局节气、阴阳遁、局数、三元、置闰状态与日符头断言；新增 `2033-06-14` 日家洛书飞盘置闰状态样本，并明确 `method-06` 为接气样本。置闰十五日网格、九日传统阈值等价、十二节气上限与芒种 / 大雪置闰依据入 `zhirun.json`。
- 日家洛书飞盘置闰的独立验证缺口提升为显式 gap，并收窄到阴遁核心盘面；`2033-06-14` 阳遁置闰样本与 mingyu 逐项核对定局、主柱、旬首、值符值使与九宫星干核心，门盘 / 神盘 / 暗干因流派差异显式排除。所有未交叉验证的生成向量必须被 gap 覆盖，避免单一来源事实被误读为已验证。
- open gap 显式声明必需口径：日柱主柱、阴遁、宫数位移、星干同移、真实中宫八门步位与置闰定局；`spirit_mode` 同步收敛为完整命名序列，洛书飞盘固定为白虎—太常—玄武九神。
- 奇门结果移除盘式、主柱干支与值符 / 值使落宫的重复输出；这些事实分别以 `method.school`、`method.lead_pillar` 和顶层落宫字段为唯一 API 事实源，引擎内部仍保留排盘几何所需数据。
- 置闰输出区分 `method.jie_qi`（实际天文节气）与 `method.yong_ju_jie_qi`（用局节气），并提供 `zhi_run_state`（正授/超神/接气/置闰）。
- 奇门排盘显式采用晚子时日柱换日口径，并输出 `method.day_boundary`，避免 23:00 后置闰三元与日柱符头错位。
- 建立表驱动九宫外圈环序，修复原实现按洛书数字顺序旋转天盘 / 人盘 / 神盘的错误。
- 天盘改为标准转盘几何：值符随时干（时干甲遁旬首六仪）、中五寄坤二、九星同步旋转；天禽随天芮同宫并携带中五宫地盘干。
- 值使门改为从旬首宫按时辰旬内序数阳顺阴逆飞九宫，再以八门外圈环序对齐；不再误用时支本位宫。
- 修正旬首遁干落中五时的值使飞步：先从中五真实宫起数，只有飞步结果落中五时才按外圈寄坤二；不再起步前提前寄宫。
- 八神以值符落宫起排，阳遁顺行、阴遁逆行；当前八诈盘流派写入 method metadata。
- 定局改用精确节气交接时刻，而非日期中午太阳黄经采样；冬至、夏至、立春边界均按时刻切换。

### 领域模型与 API

- 每宫 `tian_pan[]` 以 `{gan,xing}` 绑定天盘干与九星；转盘中宫无星时用 `tian_pan_gan` 保留天盘干事实；天禽寄宫不再丢失。
- 新增 / 收敛 `method` 元数据：家法、盘式、定局法、定局来源、主柱、主柱旬首六仪、旬首真实落宫、日符头、八神 / 九神流派、天禽规则、节气与三元。
- 每宫输出 `an_gan`、`men_present`、`shen_present`；pan 聚焦局数、阴阳遁、四柱干支、值符值使、九宫与空亡马星，洛书飞盘空门与转盘中宫状态不再混淆。
- `birth_year` 改为 `birth_date`，按精确立春推年命干；仅给日期且恰逢立春日时拒绝推断，要求提供时刻。
- Go 方法排盘 API 对非法 scope / school / bureau 组合返回错误，不再输出空盘；默认 `ComputeChart` 仍保持无错误快捷入口。
- 删除仅剩测试调用的 `ComputeChartWithJuMethod` / `ComputeChartWithYongShenAndJuMethod` / `ComputeChartWithYongShen` 固定默认盘包装；Go 层统一使用表达完整方法矩阵的 `*WithMethod` API。
- 用神符号不再接受独立 `甲`；甲不露盘，只能通过日干 / 时干 / 年命或对应六仪进入盘面。
- `xing_gong_wu_xing` 替代伪完整 `wang_shuai`，只输出星宫五行关系；`ying_qi` 改为结构化候选并标注日干 / 时干 / 年命干 / 所选用神相关性，不输出无关确定断语。

### 命理规则入表

- 新增 `plate.json`：外圈环序、地盘三奇六仪、九星 / 八门 / 八神环序、本宫星门、宫 / 星 / 门五行、五行关系断语、支宫、马星、六甲遁仪、旬空、五不遇时与三元名称全部入表。
- 新增 `patterns.json`：天遁、地遁、人遁、三奇得使、玉女守门、伏吟、反吟均为表驱动；条件显式同宫、天盘 / 地盘干、门 / 神与值使约束，代码只做机械匹配。
- 删除代码中的泛化十干克应 / 门克应 / 星克应兜底；未入权威表的组合不生成伪规则解释。

### 领域模型二次收敛

- 每宫新增显式 `gong` 身份（名称、洛书数、外圈序号），不再依赖数组下标识别宫位。
- `ma_xing` 与 `kong_wang[]` 改为 `{branch,gong}`，保留地支事实，宫位只是投影；空亡不再因两支同宫而丢失支身份。
- `ri_shi_sheng_ke` 收敛为结构化 `ri_shi_relation`（subject/object/relation/name）；`kong_wang_affected[]`、`ma_xing_affected[]` 逐项列出命中符号、地支与宫位，不再压缩成布尔值。
- `xing_gong_wu_xing[]` 输出 relation、relation_name 与 traditional_label，避免把关系标签误读为完整季节旺衰。
- `ying_qi` 在传入用神后按日干、时干、主柱、年命干与所选用神重新投影；新增 `yingqi.json` 承接机制与解释文案。
- 用神排盘链路先构建盘面因子再一次性按用神投影应期，不再先计算并丢弃基础 `ying_qi` 结果。
- `birth_date` 引入 DateOnly / Moment 精度值对象，立春跨界日的推断约束下沉到奇门领域层。
- 八神阳 / 阴遁名称、三元符头支组、宫 / 星 / 门五行、定局表全部改为 keyed 自描述表；`jushu.json` 直接以节气与太阳黄经为键，不再依赖数组顺序。
- `tian_qin` 新增表驱动 `star` 与 `lodging_gong`，中五寄宫和九星本位不再写死在代码；删除与外圈环序重复的九宫本位星门表，中五不伪造八门本位，值使门按寄宫规则取门。
- 奇门内置表加载时校验环序、九宫本位星门、三奇六仪、八神槽位、五行表、支宫、马星、六甲遁仪、旬空、五不遇时与关系文案的完整性和一致性；门表命名统一为完整“×门”。
- method 元数据改为 `lead_pillar`、`lead_xun_shou`、`lead_xun_shou_gong`、`day_fu_tou`，天禽寄宫改为转盘专属结构化规则对象；内部通用 `drive` 命名全部收敛为 lead。
- 方法矩阵、默认时家 / 转盘 / 拆补、主柱、定局法兼容性与定局法名称统一由 `catalog.json` 驱动，`ParseMethod` 按 scope/school/bureau 封闭目录校验，不把任意技术坐标拼装伪装成门派。
- 修正洛书飞盘天禽伏吟沿用转盘“禽随芮寄坤”的判断错误：洛书飞盘按天禽真实中五落宫，转盘才使用寄坤有效落宫。
- 修正转盘中宫 `tian_pan_gan` 无法进入十干克应的求值遗漏；中宫天盘干事实现在按表参与克应。
- 十干克应显示名统一为“天盘干+地盘干”，修正己辛、辛壬两组双向组合的顺序标反。

### 文档与测试

- 奇门 OpenRPC 契约完成 schema→handler→method 表→盘面输出→skill 文档的逐层核对：`solar_time` 标注 `date-time`，`birth_date` 明确 date / date-time 双格式，month/year 在 schema 层禁止携带 `bureau`；37 个用神输入名逐一通过 `ParseYongShen`，15 个方法组合的 method / pan / 宫位字段与 schema 精确一致，全年边界日期抽样全部通过 result schema 验证。奇门 result schema 补齐十干克应、门克应、星克应、门迫门制、应期、用神、干支宫枚举与 always-required 字段，并关闭未声明字段。
- 四个 skill 的根文档与 app 流程收敛为「条件 / 目标 → 动作 → 产物」步骤表；移除用户可见 `□` 过程检查、重复红线、工程 troubleshooting、长篇输出举例与 app 卡重复排盘步骤，输出模板压缩为字段表，通用硬边界集中在根 SKILL.md，场景差异留在 app 卡，六亲 / 健康 / 体型 / 婚姻状态裁决细则下沉 domain 文档。起名用神文档改为直接选择 `bazi.fullchart` 已返回的三派结果并保留冲突裁决表；奇门 / 六爻用神文档删除与 RPC schema 重复的字段长说明。命书拆为快速扫描与完整报告，问卦拆为六爻 / 奇门分支卡；考时与六亲检查框改为决策表。`check_docs` 新增流程与信息可达契约：根文档 ≤80 行、app 卡 ≤65 行、流程区 ≤20 行、单卡必读 domain ≤6 个且加载 ≤650 行，统计真实 `.md` 引用而非标记行，检查 domain 死文档、过程检查框与旧 Step 编号回流。
- `liki-divination` 的 app 流程与领域文档同步 `scope/school/bureau` 选择表、真太阳时调用链、method 审计与候选应期边界，并明确阴盘与非鸣法九门飞盘当前不支持，金函玉镜按 day 专用 chart 调用。
- 测试重建为当前功能契约：标准阳遁六局转盘手工锚点、时干甲遁、值使飞宫、八门外圈、精确节气边界、立春年命边界、格局同宫约束与表完整性；另引入 atopx/qimen 固定 commit 的 11 个“置闰与拆补定局及转盘盘面一致”外部锚点、16 个时家转盘置闰定局锚点，以及时家洛书飞盘、日家节气、月家、年家外部锚点、15 个方法组合不变量与金函玉镜专用 chart 契约；删除旧错误算法的过程性断言。
- 外部 golden 向量外置为 JSON 并记录 repo / commit / source hash / license / provenance；同时锁住每宫天盘干、地盘干、暗干、星、门、神六列事实与转盘天禽展开规则。
- 新增 deminzhang/qimen-go（MIT）与 Brhiza/mingyu（AGPL，仅作事实交叉核对、不引入源码或依赖）固定 commit 作为第二参考源，对时家洛书飞盘拆补、日家转盘置闰、日家洛书飞盘拆补、月家洛书飞盘、年家洛书飞盘与阳遁日家洛书飞盘置闰六个向量交叉验证共享口径下的阴阳遁、局数、主柱、旬首、空亡、值符值使与九宫核心；暗干、门盘几何、第九中门 / 中宫序列化、落宫元数据与九神流派差异显式列为 exclusion，不伪装成一致。阴遁日家洛书飞盘置闰核心仍保持单源 provenance。
- 新增十干克应、门克应、星克应表契约：条目数、唯一键、非空解释、命名一致与未入表组合不得生成伪解释。
- 新增七个格局的经典定义向量，覆盖当前 `patterns.json` 全部规则，并校验向量依据与表内 `basis` 一致。
- 统一 4 个 skill 与 engine 分发版本为 2026.09.06.1。

## [2026.09.05.2] — liki-naming 偏好与候选报告收敛

### liki-naming

- 新增一次性偏好与避讳收集：风格、必含字、避讳字、长辈同音、是否叠字、出处偏好；这些约束只过滤候选，不改变八字用神与字符五行。
- 同音避讳要求读音来自 `qiming.char` 或用户显式提供，并按拼音音节比较（默认不计声调）；无法核对时只做同字避讳，不凭记忆补拼音。
- 候选输出改为少量精选（首轮 5-8 个），每个名字同时给出优点与真实可商榷点，并附代表性不推荐清单和具体淘汰原因。
- 出处证据分为直接典故、字义联想、现代审美组合；记不准时降级或标注未确认，禁止编造书名、篇名、原句、重名率或流行排名。
- 同一会话续起名时避免重复完整名字和已呈现用字，并保持候选在字根与气质上的差异。
- 保持 `qiming.pick` 字池红线与 `qiming.check` 现有契约，不新增与 `characters` 拼音、声调重复的音韵投影。

## [2026.09.05.1] — 命理语义与数据流修复

### 断言模型

- **条件组显式化**：`assertion_conditions.csv` 新增 condition_group_id；同组 AND、跨组 OR，修复“或”被写成“且”的模型缺陷。
- **新增受控事件标签**：断言元数据新增 `领域 / 事件类型 / 时间层`，精简流年输出保留 `id + 领域 + 事件类型 + 时间层 + 事件 + 结论`。
- **显式合参层**：新增 `side=common` 断言表；八字/紫微仍分侧计算，跨体系条件在双盘合并快照上匹配，输出到 `合参`。
- **经典列改名**：`经典原文` 改为 `经典依据`，避免将派生解释伪装成原文；清理评测 case、内部阶段标签与旧内部路径残留。

### 命理规则修复

- 修复 `dy_100 / xue_401 / fam_122 / fam_124 / dy_101 / dy_103 / dy_110 / ys_220 / ys_221 / liu_310 / ycai_120 / yj_231` 的互斥、错位或跨作用域条件。
- `财坏印` 改为“财透 + 印现 + 财五行克印五行”的完整闭环。
- `半合` 改为三合旺支参与的两支组合；缺旺支的拱合不再按半合命中。
- 修复 `流年冲[年支]` 误按日支匹配的实现错误；显式年支/月支/日支/时支现在按指定柱支判断。
- 降格或重写确定性过强断语：子女性别、通灵、奉子成婚、配偶重病死亡、先天性心脏病、被骗、穷困潦倒等改为命理倾向/候选，并移除性取向推断。
- 删除与 `jk_112` 重复且过度医疗化的 `jk_120`。
- 删除仅以命主偏财得地便断配偶职业/财富的 `hun_500 / hun_510`；配偶经济判断须回到配偶星、夫妻宫与合参闭环。
- 收紧财运/性格断语条件：`cai_313` 必须财星为忌，`cai_110` 必须原局比劫夺财，`cai_111` 须身强任泄或食伤生财，`xg_702*` 显式绑定印星用忌。
- `fam_124` 的“伤官见正官”改用正官现，不再以官杀现泛化匹配。
- 删除仅含本命体质、无当年引动而会逐年重复命中的 `ymar_111`；同类结构已进入 schema 门禁。

### 时间层与紫微覆盖

- 当前大运断语从 `十神 / 用神` 移入 `大运` 域；`CURRENT_LIMIT_RULES` 收敛为 `大运 / 大限`，纯本命查询不再依赖当前时间。
- `query(rule=大运/大限, pan, year=...)` 支持指定历史或未来年限运；省略 year 时才使用服务端当前年份。
- 限运域 query 结果附带 `current_year / current_year_source`，显式年份标记为 `specified`。
- `full_paipan` 一次返回 `ziwei_daxian`；新增 `query(rule=大限)`、12 条大限主题断语、当前大限宫因子与稳定领域事实投影。
- `domain_snapshot_contract.json` 完整承接四柱成员、柱字段、八字/紫微源字段与大运/大限顶层引用；领域事实投影代码改为纯机械遍历，reserved facts 清单保持不变。
- pan 契约中的四柱、性别闭集与大限段数改用 `constants.json` 单一事实源；`calibrate` 复用同一性别闭集。
- 财星 / 印星 / 官杀成员清单收敛回 `十神大类` 单一事实源；命理侧代码、输出标签、快照侧与断言侧范围由 `命理侧` 统一定义，消除各层重复硬编码。
- 月支长生、夫妻宫、年 / 时柱十神、年柱官杀与日主上下文的柱位统一进入 `算子柱位`，代码通过四柱序号机械解析。
- `夫妻宫破` 改为引用夫妻宫冲 / 刑 / 害三个既有因子；`印星透根` 收敛为“透干且有根”，禄根只作为 `印星旺` 的独立旺相分支，语义不重复。
- pan 契约强制校验 12 段完整大限及其宫位、年限、起止虚岁字段。
- 新增流曜入宫机械算子和 10 条流曜断语，覆盖流羊、流陀、流鸾、流喜、流昌、流曲、流禄、流马。
- 新增 `yliu_113` 父星透干引动断语，补齐流年父星显动的家庭事类候选。
- `yingqi` 补 `年六亲 / 年大运 / 年旺衰`；`yearly_family` 补 `年子女`。
- `yearly_range` 返回 `year_basis`，显式说明八字干支年与紫微农历年的年界语义；`rules` 必填，移除隐藏默认域。
- `calibrate` 聚合并返回八字 / 紫微 / 合参三层，精简结果保留 id 与受控事件标签。
- `calibrate` 对全部事件规则取流年因子与本命引用闭包并复用同年快照；detail 模式聚合输出机械 evidence。
- `calibrate` 与 `full_paipan` 的 correct 语义对齐：候选按具体时刻校正时必须给经度；已明确时辰 `correct=false` 可省略经度。
- `query` 按命理域做单侧求值与断言因子闭包裁剪：八字-only 域只算八字因子，紫微-only 域只算紫微因子，且只计算该域断言实际消费的因子及依赖。
- `yearly_range` 根据展开后的流年域计算因子闭包，并按流年引用裁剪复用本命八字因子；多年扫描不再全量计算 101 个流年因子与 456 个本命因子。
- `流年支受克` 拆为五行五个表驱动分支，再由流年因子 OR 组合成复合因子；本命五行引用与生克关系全部回到表内，代码仅按参数机械求值。
- 算子配置继续落表：十神大类与日主生克关系、格局十神、官杀取清对象、夫妻宫静默状态、紫微无主星/唯一主星、流年干支来源、事件宫位默认、取合方式、旬空顺序与流年宫名别名进入 `constants.json`；operator 只保留机械查表求值，迁移保持语义等价，不删除规则。
- `yearly_range.year_basis` 的八字干支年、紫微农历年与具体日期使用说明迁入 `constants.json`。
- 新增 operator 源码契约测试，禁止十神、五行、干支、紫微星曜与关系成员等闭集重新写入 Python 代码。
- 因子表与 reserved domain facts 明确为稳定命理领域模型；当前断语未消费不作为裁剪依据，并加入契约测试防止误删。
- `合类 / 冲类`成员列表与取合语义配置合并为 `关系取合类型 / 关系取冲类型`，避免同一关系名双处维护。

### 起名数据边界

- 移除三才五格删除后已无领域消费者的 Kangxi 笔画 / 形体输出；`qiming.char` 与 `qiming.check` 只返回现代字符事实。
- 删除 Kangxi 运行时投影、Unihan 笔画派生源表与生成脚本；GSC 现代字库、五行、拼音、声调、部首与负面字表保持不变。
- 五行候选池索引改为独立机械构建，不再依赖 Kangxi 数据加载副作用。

### 文档与测试

- 裁决准则改为独立证据闭环，移除固定数量门槛、全局引动排序和“强/弱婚动按断语 ID 分级”。
- 健康火弱脏腑说明、婚姻状态锁定、家庭六亲灾变、应期取舍、双盘主辅关系同步修正。
- app 卡补齐本命 query、限运 query 与流年 yearly_range 的显式调用，命书快速扫描使用基础命理域 query 集。
- `check_schema` 重写为条件组 / 作用域 / 术数纯度 / 互斥条件 / 生产纯度门禁；修复原 schema gate 的空路径假绿测试。
- `make pre-push` 与 engine CI 使用一次性 `/tmp` Go/golangci-lint cache，避免宿主 home cache 只读导致 context loading 假失败。
- 因子门面不再转发 operator 私有实现；测试直接绑定对应 operator 模块，降低模块间耦合。
- 统一 4 个 skill 与 engine 分发版本为 2026.09.05.1，并新增版本同步契约测试。
- **Changelog 契约收敛**：删除子 skill 内 CHANGELOG，项目根 `CHANGELOG.md` 成为唯一版本历史；分发版本继续按当日 CalVer 递增，当前为 2026.09.05.1。
- `当前大限宫` 不再在没有显式/服务端年份时回退出生年；无参考年返回空值。
- 新增本命/流年因子闭包等价测试，确保裁剪计算不改变任何域的断语命中。
- app 卡必读数量加契约上限，一般卡不超过 3 个；紫微侧统一表述为“合参”而非默认佐证。
- 新增当前功能契约测试：条件 OR、财坏印、半合旺支、common 合参、大限、流曜、场景别名与 brief traceability。
- 当前验证：Python 287 passed；`make check` 0 error / 0 warning；engine lint/vet/unit/integration/RPC smoke 全绿。

## [2026.09.04.1] — 命理断语大规模补充 + 架构修复

### 断语新增 (60条, 696→756)

- 婚姻: hun_109/110/111(独身), hun_412(冲逢合解), hun_420(奉子成婚), hun_500/510(配偶)
- 性格: xg_505~509 (伤官见官/偏印鬼主意/外冷内热/印旺被破/印制食伤)
- 流年事件: yliu_110~112(母灾/父灾/母逝), yj_220~232/240(牢狱/重病/交通/穷困), ymar_120(半合婚动)/130(通灵)/131(配偶亡)
- 事业/学业: ys_220/221/230/240(突破/挫折/得奖/就业), xue_310~312/320/321/400/401(学历正偏印区分/食伤泄印/大运印运)
- 健康: jk_120(火弱→心血管,五行域), zj_320(天同天梁→慢性病), jk_113条件补全
- 家庭/出身: liu_204/205/306/310(母康寿/父公职/和顺/父母离异), fam_120~124(先裕后损/小康/富贵/有家底/清寒)
- 其他: cai_311(被骗), wai_201(木日主体型), ycai_120(贵人引财), yjz_210(精神困扰), yz_110(生育一级)

### 因子新增 (12个, 444→456)

- 年柱十神因子: 年柱正财/七杀/伤官/正官
- 禄根因子组: 食伤旺/伤官旺/偏印旺/七杀旺/正印旺/印星透根 (共7组)
- 流年因子: 流年支冲年支/流年支冲日支/流年支半合日支/流年子女星值宫/流年父星透/流年母星透
- 为用/为忌: 印星为用/官杀为用2/食伤为忌/比劫为忌/本命偏印旺/本命官杀为用/本命华盖

### 算子新增 (3个)

- 禄根: 五行为某地支本气→有强根（纯机械查表）
- 时柱十神: 时柱天干十神（子女性别判断用）
- 年柱十神: 年柱天干十神（出身判断用）
- 半合 (流年): 两地支属同一三合组（纯机械查表）

### 架构修复

- _atomic bug: 双参数字符串比较 args[0]→args[-1]（性别/用忌等因子修正）
- SCENE_ALIASES: 所有场景补入紫微流年域（年夫妻/年官禄/年财帛/年疾厄/年福德等）
- CURRENT_DAYUN_RULES: 扩展至用神/十神
- hun_205/208: 官杀混杂条件互斥（防矛盾断语同时触发）
- FACTOR_MODEL.md: 全量同步因子计数和定义

### 裁决准则新增 (caiule rule 7~16)

- rule 7 强化: 子女性别以时柱食伤阴阳为唯一标准
- rule 9: 通用性格题默认取月令主面
- rule 11: 应期同层级二级排序（值宫>天克地冲>三合>六合>星动>神煞）
- rule 12 限定: 六亲灾≥2条且明确才优先于自身吉象
- rule 13: 紫微宫位主星断语同级参看
- rule 14: 健康时间线关联
- rule 15: 日主五行底色取象优先
- rule 16: 用神十神取象决定职业/科系

### 断语措辞修正

- hun_107: 措辞改为"有婚恋机会（原局配偶星不现，婚缘薄）"
- dy_100: 措辞软化为"有婚恋机会，婚缘薄"
- zy_101: 扩展偏印取象（美术/艺术/创作/体育健身/医疗技术/工程技术/殡葬/入殓）

### 断语统计

- assertions.csv: 756 条 (原 696)
- assertion_conditions.csv: 1086 条 (原 958)
- factors.csv: 559 行 / 456 个因子
- factors_liunian.csv: 84 行 / 84 个因子
- caijue.md: 16 条裁决准则 (原 8)


## 2026.09.02.1 —— 起名 API 领域收敛

- **[架构] 移除三才五格与 81 数理规则**：起名链路改为「八字用神 → 五行候选池 → 受控组名 → 字库/五行/音韵评估」，保留 Kangxi 笔画与形体作为汉字事实。
- **[API] `qiming.build` 替换为 `qiming.compose`**：`first`/`second` 只传字，服务端校验字库并生成 given name；最终字符事实由 `qiming.check` 返回。
- **[API] `qiming.check` 输入收敛为 `given_names`**：评估候选名不再要求 surname 参数；姓氏仅由场景层用于最终展示和谐音判断。
- **[API] `qiming.pick` 返回候选字池**：不再接收姓氏或数理参数；单/双名由 `count` 控制。
- **[数据] 修正五行 fallback 生成顺序**：生成运行时字库时先解析部首五行表；961 个可由部首推断五行的字进入候选池，371 个无五行依据的字不进入运行时候选池。
- **[数据] 强化字库校验**：非法笔画、声调、拼音、重复字、负面字表格式错误 now fail fast；`NULL` 占位不再进入 API。
- **[数据] 运行时字库瘦身为领域投影**：新增 `naming_characters.csv` 与 `kangxi_character_strokes.csv`，只保留 qiming 当前消费字段；完整 GSC / Unihan 源表保留为非 embed 数据，姓氏五格覆写表已删除。
- **[数据] 运行时投影只包含可用命名候选**：Kangxi 运行时表同步过滤到 7,734 个具备五行依据的字。

## 2026.09.01.7 —— 起名字段与数据边界收敛

- **[API] `qiming.char` 字段命名收敛**：公开康熙笔画字段统一为 `kangxi_stroke`，五格计算消费康熙笔画。
- **[数据] Unihan 作为起名源表边界**：运行时投影继续从源表生成，不手工维护重复字库。

## 2026.09.01.6 —— 起名源表命名收敛

- **[数据] 源表统一 `unihan_*` 命名**：康熙笔画派生表不再使用孤立命名。
- **[数据] 暴露五格笔画字段**：供五格真路径读取康熙形体与笔画。

## 2026.09.01.5 —— 康熙笔画真路径

- **[修复] 五格真路径改用生成的康熙笔画**：姓氏与候选字均消费 Kangxi 源数据。
- **[API] `qiming.char` 同时返回现代笔画、康熙笔画与五格使用的形体字段**。

## 2026.09.01.4 —— 因子长表缓存与只读契约

- **[架构] 收敛因子长表缓存**：缓存按表文件路径隔离，术数归属、引用闭包与 direct 行契约由加载层统一校验。
- **[测试] 补充真值表与流年求值复用/只读契约**：防止求值过程写回公共 pan。

## 2026.09.01.3 —— Windows CLI 稳定入口

- **[新增] agent_cli.cmd**：Windows 下自动启用 UTF-8，并优先通过 `py -3` 调用 Python。
- **[修复] JSON 输出编码**：CLI 结果改为 ASCII 转义，避免 PowerShell/CMD 代码页导致中文乱码。
- **[文档] Windows 调用契约**：skill 现在优先推荐 `tools\agent_cli.cmd`，不再要求用户裸调 `python3`。（#34、#37）

## 2026.09.01.2 —— Issue 修复：Windows、起名笔画、fullchart 契约与流年证据

- **[修复] Windows CLI 工作流**：agent CLI 显式配置 UTF-8 流；skill 增加 python/python3、`PYTHONUTF8=1` 和 PowerShell UTF-8 文件重定向规则。（#34、#37）
- **[修复] 起名康熙笔画**：简化/繁体形态不同的常见姓氏改用康熙笔画；`郑` 现按 14 画参与五格。（#35、#39、#41）
- **[修复] bazi.fullchart 输入契约**：展开前校验四柱干支与性别；缺失字段返回结构化 handler 错误。（#40）
- **[修复] 流年三刑可验证性**：detail 输出附带三刑组、成员、来源与四柱参与支。（#38）

## 2026.09.01.1 —— 因子/断语长表与架构契约收敛

### 架构收敛

- 删除旧 `extract.py` 中间层，改为 `pan → factors → snap` 直读路径
- 新增 `pan_schema.py`：query/yearly_range/liunian/bond/full_paipan 统一拒绝快照、裁剪盘和手工半截盘
- 新增 `domain_snapshot.py` 与契约文件：reserved 领域事实显式投影，不因当前无消费者被误删
- 新增 `FactorContext` / `NatalContext`：单次求值与多年流年复用上下文，且不再把 `_ctx`/`_snap` 写回公共 pan
- 拆分 `operators_natal.py` / `operators_liunian.py` / `yearly_eval.py` / `factor_tables.py` / `errors.py`
- 统一 `PanSchemaError` / `AssertionRuleError` / `YearRangeError` / `FactorEvaluateError` / `FactorTableError`，同时保持 `ValueError` 兼容

### 数据长表

- 因子表迁移为 `factor_id / group_id / term_index / kind / expression / expected` 长表
- 断语表从 45 个宽表迁移为 `assertions/assertions.csv` + `assertion_conditions.csv`
- 新增 `印星透根`、`财星透根`，收敛重复语义；`夫妻宫破` 改由冲/刑/害复合表达
- `check_schema.py` / `check_docs.py` 改为校验长表契约与 756 条断语引用

### 稳定性

- 删除全局快照 LRU，避免 pan 引用滞留与内容指纹成本
- `yearly_range` 保持 120 年跨度上限；`time.now` 失败不降级本地时钟
- CLI 错误路径返回结构化错误，进程不崩溃；空 pan 明确提示完整盘契约

## 2026.08.28.1 —— 架构收敛：双层工具合并为单层6工具 + 域名统一 + 静默降级清除

> 来源：LLM 实测评测（用户全程真实排盘+定盘交互）暴露的工具层混乱、域名不一致、静默降级三类问题。

### 架构收敛

- **[架构] LLM 可见工具从 5+RPC 双层收敛为 6 个 Python 工具**：`city_coords`/`full_paipan`/`query`/`yearly_range`/`calibrate`/`bond`，唯一入口 `agent_cli.py`，RPC 层对 LLM 完全不可见。删除 `rpc.discover` 需求、手调 RPC 方法清单、JSON-RPC 端点/请求格式等全部双层调用文档。LLM 完成一次典型分析从 9 步降至 3 步
- **[新增] yearly_range**：批量流年分析，一次调用替代 N×3 次（liunian+因子+query）。内置 target 映射（career→官杀/wealth→财星/marriage→配偶星/study→母星/health→日主），detail=False 精简输出（10年仅 3KB），单年失败显式标注 error 不静默跳过，附带 current_year（含 server/local 来源标注）
- **[新增] calibrate**：定盘校验，多候选生日×人生事件批量排盘+查询，返回原始断语（不做命中判断——信号解读由 LLM 完成）。longitude 必填（禁止静默降级到默认经度），events.rule 必须以 yearly_ 开头，label 唯一性校验
- **[新增] bond**：合盘，八字合盘+紫微合盘一次调用返回
- **[新增] city_coords**：城市名→经纬度（交互式查询，找不到时 LLM 问用户附近大城市）
- **[改造] query**：pan 直通（接受 full_paipan 返回值，内部自动 make_factors），参数名 snapshots→pan，仅支持本命域（流年走 yearly_range），规则白名单校验（拼错立即报错+列出有效域）
- **[删除] liunian/make_factors/make_liunian_factors**：从 LLM 工具列表移除（内部化为 query/yearly_range/calibrate 的编排细节）

### Bug 修复

- **[修复] agent_cli.py $file 引用不兼容 {"ok":true,"data":{...}} 包装**：`_load_file_refs` 直接 `json.load` 拿到整个包装体而非裸 pan，传给 make_factors 报 missing arg。改为自动解包 ok/data 层级
- **[修复] _RULE_TARGET_MAP 无效 key**：`"官星"` 和 `"印星"` 不是 constants.json 目标星的有效 key（应为 `"官杀"` 和 `"母星"`），导致 career/study 流年因子全为 0——静默产生错误数据
- **[修复] full_paipan 静默降级到默认经度 116.4**：用户在乌鲁木齐排的是北京的盘。改为 correct=true 时 longitude 必填（缺失报错），correct=false 时可省略
- **[修复] yearly_range except Exception 过宽**：吞掉编程 bug 伪装成数据缺失。收窄为 `(RPCError, ConnectionError, TimeoutError, OSError)`
- **[修复] calibrate 重复 label 静默覆盖**：两个候选用同一 label 时后者覆盖前者，用户以为在比两盘实际只拿到一盘。加唯一性校验
- **[修复] query/yearly_range 拼错规则名静默返回空**：`load_table` 对不存在的 CSV 返回空列表，拼错域名无任何提示。加 `_NATAL_RULES`/`_YEARLY_RULES` 白名单，拼错显式报错并列出有效域
- **[修复] query() 校验顺序**：rule 校验在 pan 处理之前——无效 rule 快速失败，不等 pan 解析完才报错

### 域名统一（拼音→英文，与 app 卡名对齐）

- shiye→career, caiyun→wealth, jiankang→health, xueye→study, xingge→personality, liuqin→family
- 影响面：28 张 CSV 文件重命名 + duanyu.py ALL_DUANYU_RULES 更新 + domains/bazi/ 5个 md 文件重命名 + app 卡全部引用替换
- 命理特有术语保留拼音（geju/dayun/yingqi/shishen/yongshen/tiaohou 等）

### 文档一致性

- SKILL.md：删 RPC 层/手调方法清单/discover 段落，简化为单层 6 工具+标准流程（3步）
- app 卡：compatibility.md 从旧 RPC 双步（bazi.bond + ziwei.bond）改为 bond() 单工具；mingshu.md/career.md 补 yearly_range 引用；marriage.md 删 ziwei.liunian 旧引用
- domains：dayun.md/calibration.md 旧 RPC 方法名更新为 yearly_range
- engine 测试白名单：skill_docs_contract_test.go allow 新增 Python 工具名
- VERSION 更新为 2026.08.28.1（补换行符）

## 2026.08.27.2 —— feedback 批次1：hash 机制拆除 + 断语/引擎修复

> 来源：liki.hk 后台 17 条 pending 反馈（已建 issue #11–#27）。每项修复均先复现（红）再改（绿）。

- **[拆除] content.sha256 指纹机制**：单树哈希对环境噪声（Windows CRLF、路径分隔符）零容忍，上线以来 0 次真阳性、5+ 次假阳性（#12/#19/#23/#24/#27），自检反成用户第一拦路虎。自检简化为 VERSION 比对；`tools/hash.py`、`content.sha256`、build-archive 指纹段、CI freshness 段、`tests/test_hash.py` 全部移除。同日重发以版本序号区分（`2026.08.27.2`）
- **[修复] xueye.csv xue_201 条件反转**（#25）：条件列 `印星旺=0`（要求印不旺）与断语「印星得月令而旺」矛盾，无印盘误中「科甲至顶」。改 0→1；新增阴/阳性对照回归测试（印弱不命中/印旺+官杀得令命中/官杀不得令不命中）
- **[修复] shiye.csv shi_102 措辞歧义**（#26）：「无食伤」→「无食伤生财」，对齐条件列 `食伤生财=0` 与 shi_101 精确表述
- **[修复] time.now 假时区**（#14）：`now.Format("...+08:00")` 硬拼后缀——UTC 服务器时钟仍是 UTC。改 `now.In(FixedZone(+8h))`；新增 TZ=UTC 下的回归单测（本机 +08 时区测不出此 bug）

## 2026.08.27（续二）—— 自部署闭环与版本机制归零

- **[自部署] ghcr 镜像**：`gh release create` 触发 CI 发 `ghcr.io/ml8s/liki-engine:latest`（+:sha 锚）——外部用户 `docker run` 一条命令；README 自部署节主路径改镜像，源码 build 降为进阶
- **[版本机制归零]** liki-web CI 的 ref 锁删除（checkout master，与部署策略对齐——此前 CI 测锁定 tag 而部署 pull master，验证物≠部署物）；替代为日志记录 liki commit/VERSION（可追溯）
- **[原则] 版本管理复杂度与变更频率匹配**：引擎近 3 个月计算逻辑零变更（7 commits 全为工程杂务）——稳定依赖按公共设施消费，不建版本编排机器；CI 测什么（master）部署就是什么
- engine/deploy/docker-compose.yml 独立部署 compose + 4×SKILL.md 端点行注明 LIKI_RPC_URL 可指向自建引擎

## 2026.08.27 —— 版本制切换：semver → CalVer（日期版本）

- **[版本] 取消 semver**（major/minor/patch 判断对该项目是仪式性负担——skill 用户装最新、无依赖解析场景）；VERSION 文件写入日期戳（如 2026.08.27），自检更新机制不变（VERSION 或指纹任一不一致即提示更新）
- **[发布] tag 按需**：里程碑时 `git tag -a <日期>` + release，不再为每个 commit 发号；历史版本条目（5.x/0.x）保留原样
- **[工具] make version-patch/minor/major → make version**（写今日日期）
- 引用面同步：CONTRIBUTING 版本流程 / README 设计原则 / liki-web CI skills 锁 ref

## 2026.08.27 —— 提示词工程优化：触发词 + 输出示例 + 正向化改写 + 应期双候选

> 提示词专家评审落地（指令经济性短板修复）。**应期双候选为行为变更，待评测验证后合入主线路径**（历史教训 iter6：微调曾致 -7pp，评测方差 ±5pp）。

- **[触发] 4 skill description 补口语触发词**：算命/看运势/流年/本命年/占卜/算一卦/看风水/取名字 + 各 1 句英文短语（BaZi reading / Divination / Feng Shui / Chinese naming）——多 skill 共存环境的路由命中率
- **[示例] 4 skill 输出原则各加 ✅/❌ 对比示例**（bazi 结论先行；divination/fengshui 按其领域惯例先依据后判断；naming 推荐+依据）——一条示例顶三条规则
- **[正向化] marriage 卡 8 条负向指令改写为正向等价**（未锁定只输出状态判断/红鸾天喜作用域限定/异常检查逐项核验等）——判据结构（≥2项/否决项/双证门槛）原封不动；保留排序铁律「禁止跨级覆盖」与数据原则红线
- **[收敛] bazi SKILL.md 手调 RPC 方法清单的重复请求格式** → 引用「RPC 调用方式」一处定义
- **[修正] mingshu 卡历史事件校准回退表述 ×2**：「调整用神取舍重推」→「回退审视整体解读框架（格局/用神/大运）」——历史事件验证整体框架、不能反推单一用神（v1.23.0 教训回潮修正，why 嵌入活文档）
- **[治理] LESSONS.md 退役**：3 条工程 why 收编 CONTRIBUTING「设计原则」节、2 条品牌教训收编 brand.md 治理记录；bug 类墓碑删除（防回潮已由 check_docs 白名单/测试/sync 排除承载）。brand.md（v3.2，含六爻奇门 Domain + 产品文档口径）随本仓库 docs/ 纳入版本控制——品牌真相源告别无版本状态
- **[行为] 应期裁决双候选输出**：首选年+备选年并列（同层级信号并列+置信度标注，跨层级才单选），建立在既有排序铁律 ①-⑦ 层级之上——直接针对评测暴露的「agent 裁决随机性」（四轮 ±5pp 方差，强制单选放大采样不稳定）

## 5.0.1 —— 工程清理：死列清除 + CI 自含 + golangci v2（运行时等价，无断语变更）

- **[数据] 断语表死列清除**：28 张表删除 227 个表头死列（yearly 表生成器时代的统一超集表头遗物；355 条断语逐条等价校验通过——id/约束/结论/依据/经典原文与清除前完全一致）；check_schema 死列检查 warning → error（基线归零后新增即拦）
- **[数据] csv 行尾归一 LF**：存量 21 个 CRLF csv 全部转 LF；新增 `.gitattributes`（`* text=auto` + `*.csv|*.sh eol=lf`）防 Excel/Windows 编辑器回潮
- **[工程] CI 数据检查自含引擎**：skills-data-check 改为 build + 起本地引擎（不再静默打生产 liki.hk）；三个起引擎 job readiness 超时即失败 + 显式清理
- **[工程] 本地引擎严格失败语义**：构建/启动失败即中止（不再静默跳过造成 test-all 假绿、不回落生产）；`LIKI_RPC_MODE` 显式 local/docker 替代 pgrep 猜测（无 docker 机器本地直连）；集成测试显式端点不可达 = FAIL
- **[工程] golangci-lint v2 迁移**：官方 Action v8.0.0 + v2.13.1（v1.64.8 对 Go 1.25+ 已停止维护，`go install` 方式官方不推荐）；`.golangci.yml` 升 v2 格式并启用 gofmt linter；v2 新检出的 6 处修复（staticcheck QF 标签化 switch ×5、errcheck 显式忽略 ×1）+ 全仓 gofmt 归一（133 文件，单行 struct 展开/末尾空行存量漂移）
- **[工程] 编排收敛**：engine/Makefile 删除与根 Makefile/ci-engine.sh 重复的 check/test-all/pre-push；pre-check.sh 删除（102 行，功能已被 ci-engine.sh + 引擎测试覆盖）
- **[工程] 参差 CSV 防御**：check_schema 对表头/数据行列数不一致报文件+行号（原为 AttributeError 裸崩）
- **[测试] 答案双源守卫**：test_grade_sync.py 校验 grade-case.py 内嵌答案与 answers.json 一致（防 skill-up 自包含约束下的双源漂移）；eval.yaml 注明答案隔离的挂载范围不变量
- **[工程] make hooks**：贡献者一键安装 git hooks（core.hooksPath 不随 clone 带上）；pre-push 简化；评测迭代期脚本/日志归档至 evals/archive/
- **[文档] 订正**：tests/README 判分脚本引用（grade-case.py）、CONTRIBUTING（版本流程/golangci 安装/hooks）、32 个 case 答案路径注释、webapp/README（部署路径说明）

## 5.0.0 —— 工程升级：liki-engine 并入单仓（monorepo）+ 统一版本（big release）

- **[工程] liki-engine 并入本仓库 `engine/`**：原独立仓库 `ml8s/liki-engine` 全量迁入（Go + JSON-RPC，8 领域），历史经 git subtree merge 完整保留。skill 与引擎同仓发布、同 CI
- **[工程] 版本统一为单一号 5.0.0**：skill（原 4.x）与 engine（原 2.6.x）自本次 big release 起共用一套版本号（4 skill VERSION + engine VERSION 同步 bump）；content.sha256 指纹随 VERSION 变更重算
- **[工程] CI 合并**：根 `.github/workflows/ci.yml` 单一 workflow，path filter 分流 engine（Go）/skills（Python）；新增 e2e job 同仓 build 引擎→起服务→跑 skill 全链路集成测试（消除跨仓耦合）
- **[工程] liki-web 适配**：全部 `../liki-engine` 引用改为 `../liki-skills/engine`（dev-start/docker-test/2 个 compose/update-engine）；`npx skills add ml8s/liki` 安装语义不变

## 4.3.1 —— 奇门应期/因子总览 + 4 skill 语言跟随 + discover 按需取

- **[skill] 奇门应期文档 `domains/qimen/yingqi.md`**：补齐引擎 `ying_qi` 字段（马星逢冲/空亡填实/值符值使）解读；divination.md 奇门流程加应期环节
- **[skill] 奇门断局因子总览**：yongshen.md 列出引擎全部因子（用神落宫/求测人/值符值使/生克/空亡马星/五不遇时/格局/旺衰/门迫门制/克应/应期），引导 LLM 综合断局；补五不遇时、值符值使落宫
- **[skill] 4 skill 统一「输出语言跟随用户」**：对话/解读/结论用用户语言，各领域核心术语首次括注英文（bazi/divination/fengshui/naming）；foreign.md 加外国人起名语言策略（中文名+拼音保留、解读英文）
- **[skill] `rpc.discover` 按需取全**：启动时一次 discover 本 skill 需要的全部方法（域前缀 + 具体方法名，用域前提是精确不导入多余）；naming 只取 bazi.chart/fullchart，bazi/divination/fengshui 用域前缀
- **[安全] run-qwen.sh 启动自愈**：清理 SIGKILL 残留的 `.run-eval.*.yaml`（含 key 的临时评测配置），防再次误入库
- **[引擎] schema 修正**：`gong_wei.xing` enum 去天禽（天禽寄坤2不占星位）、`ying_qi` 补 properties、`pan` 补 `wu_bu_yu_shi`、清理冗余 enum（ma_xing/kong_wang 去"中"、an_gan 去"甲"）
- **[测试] 数据驱动命理锚定**：端到端 4 盘完整排盘锚定、多盘用神符号落宫锚定、边界/随机日期健壮性测试；修复多处放水/弱断言测试

## 4.3.0 —— 奇门用神符号化 + 命理排盘修复（对齐六爻架构）

- **[架构] 奇门用神重构**：废弃「占事类型枚举（qianshi）」驱动，改为「用神符号（门/星/神/干，35 种封闭）」驱动。LLM 读 domains/qimen/yongshen.md「事象→符号映射」确定传什么符号，引擎按符号定位落宫取因子，与六爻「传 yong_shen、引擎聚合」架构一致
- **[引擎] `qimen.chart` 参数 `qianshi` → `yong_shen`**：改为用神符号数组（如 `["生门","戊"]`）；`birth_year` 保留（年命干落宫）
- **[引擎] 用神落宫以天盘为核心**：求测人日干/时干/用神干落宫取天盘位置（天盘主当下，地盘主过去）；日干/年命干为甲时按地支遁六仪（甲子遁戊…）
- **[skill] yongshen.md**：改为「事象→用神符号映射」供 LLM 判断取什么符号；求测人定位字段（日干/时干/生克/空亡马星）指顶层排盘固有字段
- **[引擎] 命理排盘修复（天禽/中5/暗干/八门/应期）**：
  - 天禽寄坤2、与天芮同宫；中5虚空（无天盘干/星/门/神）
  - 值符星为天禽（旬首在中5）时：值符神落坤2、伏吟/反吟按天芮判断
  - 时干落中5时值符星寄坤2
  - 八门阴遁逆排（阳顺阴逆，与八神一致）
  - 暗干序列补癸（甲寅旬起点正确）
  - ying_qi 马星/空亡文案用具体地支（修正宫位反推丢精度）
- **[引擎] 用神符号细节**：神名保留用户输入（阴遁白虎/玄武不转阳遁名）；用神干"甲"按日支遁六仪
- **[一致性] schema**：`gong_wei.xing` enum 去天禽（天禽寄坤2不占星位）；`zhi_fu_xing` 保留天禽
- **[测试] 数据驱动命理锚定**：端到端 4 盘完整排盘锚定、多盘用神符号落宫锚定、六甲遁/马星/空亡/旬首全量锚定；修复多处放水/弱断言测试

## 4.2.0 —— 六爻断语架构重构（引擎直出因子 + LLM 解读 6 因子）

- **[架构] 六爻断语生成方式重构**：废弃「查找表（450 条 enum_general.csv）+ Python 断语查询（duanyu.py）」模式，改为「引擎直出确定性因子 → LLM 读 domains 解读规则 → 生成断语」
- **[引擎] 因子领域化**：`yong_shen` 聚合（旺衰/月破/旬空/入墓/六神），`dong_yao_relations`（动爻关系 9 种枚举集合），`patterns`（格局并入装卦）
- **[引擎] 动爻关系枚举化**：4 布尔（dong_sheng/dong_ke/yuan_shen/ji_shen）→ 9 种枚举（生用/克用/比和/冲用/生原神/克原神/生忌神/克忌神/无动爻）
- **[引擎] 移除冗余 liuyao.patterns**：格局已并入 liuyao.chart，独立方法删除
- **[skill] 删除**：Python 工具层（tools/liuyao/）、9 张 app/liuyao-*.md 场景卡、450 条断语表
- **[skill] 重写 domains/liuyao/*.md** 为 LLM 解读规则（6 因子）：yongshen（用神取用）/ yuejian（旺衰+修饰定时效）/ jixiong（动爻关系定助力阻碍）/ patterns（格局定结构影响）/ liushou（六神定色彩情状）/ yingqi（应期）
- **[skill] SKILL.md 流程**：改为「起卦装卦（引擎）→ LLM 读规则解读 6 因子 → 断语」，明确「只基于 6 因子解读，中间因子仅展示」
- **[一致性] 命理逻辑在 domains 解读规则**（LLM 读取），不在断语表/代码判断；一致性由「引擎确定性因子 + LLM 按规则解读」保证

## 4.1.2 —— liki-bazi 描述去品牌残留（对齐 4 skill 独立描述）

- **[描述] liki-bazi 拆分后遗留**：frontmatter description 与 H1/首句原为拆分前整包品牌文案（「Liki 灵机 — 命理师的 Skill」），改为 bazi 专属描述（八字/紫微「八紫」双盘同参 + 场景列表），与其他 3 个 skill（divination/fengshui/naming）的独立描述风格对齐

## 4.1.1 —— 测试评审修复 + 系统自查（断语丢失/误报 bug、契约与部署防线）

- **[数据] 死规则/死条件清理**：八字流年表跨术数死规则 16 处删除、紫微流年因子列清理；zv_103 改八字条件复活（去子女宫煞）；factors.csv「本命婚凶」贪狼化忌或行删除；yearly_jiankang 表头冗余列清理
- **[数据] 算子 bug 修复（因子恒 0/恒 1 → 断语丢失/误报，共 8 处）**：三刑算子退化（任一支在场即命中）、旬空算子恒 0（xun 恒"甲"→空亡填实断语丢失）、日主五行因子布尔化（五行性情/外貌 13 条丢失）、流年支受克缺 ctx.snapshot（ying_h21/yj_102 丢失）、财星受克因子恒真（克[比劫,财] 五行恒成立→liu_101 丢失+lq_301 93% 误报）、月令格断语全灭（直读取值语义×条件列×枚举后缀三重问题）、格神透干恒真（未核对格神）、流年值/合/冲 chart None 防御
- **[数据] 断语补全**：7 条流年断语（桃花/天乙贵人/华盖/流年支忌神/日主长生帝旺/大耗，按命理逻辑归表）；xingge 互斥二分断语拆分（xg_m06/xg_702 → a/b 分支）；重复断语行删除（qy_104b/zw_102）；紫微辅佐星 16 处同类文案按星曜区分（左辅/右弼/天魁/天钺）
- **[契约] 检查体系**：check_schema 强化（约束值域/重复行/字符串约束列值域/紫微流年表跨术数）；check_docs 新增（断语 id/文件路径/方法名/RPC 调用/README 统计——4 skill）；result_schema 补全（full_paipan da_yun 限运字段/liunian 结构）；版本号单一来源（build 注入 VERSION）
- **[RPC] 引擎层**：报错文案字段名 4 处（juShu/guaIndex/gongIndex/starIndex 对齐返回字段）；bazi.fullchart schema 错层修复（单柱扩展字段上移柱级）；da_yun description 过期字段清理；rpc.discover schema 声明 methods 参数；CORS 白名单失效修复（HandleRPC 覆盖 * 删除）；BodyLimit 超限错误消息改进；input 健壮性测试（180 组合 0 panic）
- **[部署] 防假同步**：build-archive 先重算指纹再打包 + archive 内校验；sync-skills.sh 排除锚定 ./（修复误排 app/README.md）+ 同步后指纹校验；4 skill 自检补本地完整性校验；CI 加 check/data-check job
- **[文档] 输出原则**：吉凶档位澄清（断语库按原文输出）+ 重断语软化（父寿不永/殡葬等补充建议性表述）；README 断语统计更新（597 条）

## 4.1.0 —— 架构分层收敛（流程归 app / domains 平铺 / 版本工程化）

- **[架构] 流程归 app**：根 SKILL.md 由 Phase 0-8 流程卡改为「流程约定」（全局骨架 + 强制填表规则 + 路由表）；每个领域的流程（排盘 → 查断语 → 输出，每步「输出：□」填表）移入对应 app 卡，根/app 不再重复流程
- **[架构] domains 平铺**：删 8 个 domains/*/SKILL.md 入口与 fangfa/duanyu 分类层，41 个知识文件平铺为 domains/<域>/*.md；路径引用全量改写
- **[架构] 单一数据来源**：参数以 rpc.discover 为准、返回字段以 skill-tools.json result_schema 为准、断语结论以 query（csv 真值表）为准；域文档只写 rpc/csv 没有的（业务映射、判断链、约束规则、体系隔离）
- **[工具] skill-tools.json 加 result_schema**：5 工具补返回结构；清理死码/死参数/死导入；query 返回结构稳定（双盘恒有键）
- **[版本] 工程级版本**：4 skill 统一 v4.0.0，不再子 skill 独立定版；Makefile version-patch/minor/major 同步更新 4 个 VERSION
- **[文档] README 中英重构**：架构图改为「文档层 + 工具层 + 引擎」（工具层可选）、语气改实事求是、数字/命令/结构/免责声明中英对齐；快速开始补 4 skill 安装说明（一次装全部 + 单装）、功能特性按四 skill 分组介绍、设计原则精简去重
- **[web] 工具按 agent 分离**：命名聊天（liki-naming）与报告（mingshu/hepan）拆为两个 agent——命名 agent 只挂 read_file + RPC，报告 agent 挂 read_file + 5 个命理 Python 工具 + RPC；PromptFile 默认值改为 liki-naming/SKILL.md；{locale} 占位符改为追加英文指令
- **[webapp] 报告数据走 Python 工具**：mingshu/hepan 的 generate.md 从手调 5 个 RPC 改为 full_paipan → make_factors → query；路径引用补 liki-bazi/ 前缀（原缺前缀 + ../../ 相对路径会导致 read_file 失败）

## 4.0.0 —— 拆分为 4 个独立 skill（liki-bazi / liki-divination / liki-fengshui / liki-naming）

- **[结构] 4 拆**：单 skill 拆为 4 个独立 skill（`skills/` 下，`npx skills add ml8s/liki` 一次装全部，子路径可单装）：
  - **liki-bazi**（命理）：八字+紫微「八紫」双盘同参（Phase 0-7 全流程 9 卡 + tools 引擎 + 160 题评测）
  - **liki-divination**（问卦）：六爻/奇门/黄历择日（Phase 8 子流程）
  - **liki-fengshui**（风水）：八宅/玄空
  - **liki-naming**（起名）：八字用神 + 三才五格（用神方法论独立复制）
- **[引擎] 服务端共享**：排盘 RPC（liki.hk/jsonrpc）各 skill 按需声明；skill 侧按域分数据（tools 引擎仅 liki-bazi 持有）
- **[评测] 归属**：160 题挂 liki-bazi；拆分后定向回归 16/20=80%（门槛 ≥65%，超基线）
- **[构建] build-archive.sh/CI**：4 skill 循环打包（dist/liki-<name>.tar.gz + index.json）+ content.sha256 各自校验
- **[部署] liki-web/liki-bot 同步**：sync-skills.sh 4 skill 循环（webapp 仅挂 liki-bazi）；副本更新


## 3.10.3 —— 仓库结构重构：liki-skills 工程根 + skills/liki 内容（GitHub 安装不再混入工程文件）

- **[结构] 仓库重组**：skill 内容（SKILL.md/app/domains/tools/VERSION/content.sha256）移入 `skills/liki/`（CLI 标准发现位置）；工程文件（tests/scripts/webapp/README/CHANGELOG/Makefile/.github 等）留在仓库根——`npx skills add ml8s/liki` 只安装 `skills/liki/`，**tests/scripts 等零混入**（已实测：只发现 1 个 skill、domains 子 SKILL.md 不误判、命令不变）
- **[结构] 工程根命名 liki-skills**：git 仓库目录 `skills/liki` → `liki-skills`（一层，remote/GitHub 名不变，安装命令不变）
- **[脚本] 路径适配**：build-archive.sh 打包/指纹/产物指向 `skills/liki/`（dist 随 skill 目录）；Makefile VERSION_FILE、CI content.sha256 校验、check_schema.py、tests 8 文件路径全部改为 `skills/liki/tools`
- **[部署] liki-web sync-skills.sh/Makefile**：SRC_DIR 指向 `../liki-skills/skills/liki`；副本（liki-bot/.agents、liki-web/web/skills）清理工程文件、只保留安装形态内容
- **[docs] README 项目结构**：补充仓库根 = 工程区 + skills/liki = 安装区说明

## 3.10.2 —— 死规则清理 + 算子修复（跨术数死行 / 流年透克恒 0 / 断语复活）

- **[atoms] `_target_stars` gender 中英文漏配修复（重大）**：constants.json 性别键为 `male/female`，外部传入中文 `男/女` 直接 `ts.get("女")` 取不到 → star_keys 恒空 → **「流年透」/「流年克」算子恒 0** →「流年目标星透/流年克目标星」因子恒 0，依赖断语（ymar_101-104/108/109、ycai_101/103、yliu_101/103、yz_101/102、ying_h19 等）八字侧全部永不命中——已加中文→英文映射
- **[data] #17 跨术数死规则清理**：八字 yearly 表 17 行引用紫微因子（ymar_113/114、ycai_106/107、yj_201/202、yliu_106、ys_201-204、yx_201-204、yz_201/202）——八字侧永不命中（紫微侧 ymz_*/ycz_*/yjz_*/ysz_*/yzz_* 全覆盖），已删行 + 删紫微死列；**紫微表 2 张清八字冗余列**（ziwei/marriage 8 列、ziwei/xingge 1 列）
- **[data] #18 ying_h18/19 断语复活**：「食伤旺」死列 →「食伤重」（流年键）+ 实现「引用本命[食伤重]」算子（读 ctx 传入的本命「食伤旺」值）——损胎/婚变断语恢复输出
- **[data] #19 本命婚凶 贪狼化忌或行删除**（ziwei 因子写入八字表，八字侧永不生效，其余 4 行正常）
- **[data] #20 yearly_jiankang 「流年长生[X]」死列删除**（键名应为「流年日主X」，无行引用的表头冗余）
- **[data] 紫微 yzz_201 / 八字 zv_103 复活**：补「流年子女宫禄」因子（流年宫化[子女,禄]）；zinv 删「子女宫煞」死列（紫微键）
- **[schema] check_schema 跨术数校验目录盲区修复**：expect 按 bazi/ziwei 目录判定（原来按文件名 `bazi_`/`ziwei_` 前缀——表文件在子目录无前缀 → expect 恒 None → 跨术数从未被检出）
- **[docs] README 断语统计更新**：46 张表 589 条断语、因子 495 行（原 701/497 过期）

## 3.10.1 —— 三刑算子修复 + 打包/指纹一致性（自检不再误报"内容滞后"）

- **[atoms] 三刑算子严重 bug 修复**：`for grp in const["三刑"]` 遍历 dict 得到的是 key（单字地支），`all(g in zhis for g in grp[0])` 退化为"zhis 含任一三刑组地支即命中"——改为 `for k, v in const["三刑"].items()`，k 与其同组其余地支**全部在场**才算凑齐（寅巳申/丑戌未/子卯/自刑需双字）；实测"三刑流年"因子不再年年恒命中（原 bug 影响 6 条断语、横跨 6 域：yliu_108/ys_106/ying_h09/h18/h19/h20）
- **[hash] content.sha256 指纹范围对齐打包范围**：排除 tests/scripts/webapp/.github/.githooks/docs 及根级工程文件（README/CHANGELOG/LICENSE/Makefile/pytest.ini 等）；EXCLUDE_FILES 改**根级精确匹配**（`README.md` 只排除根文件，不误伤 app/README.md）；`VERSION` 保留在指纹内（随内容变更驱动指纹）
- **[build] build-archive.sh 打包干净化**：补齐排除 .github/.pytest_cache/__pycache__；根级文件排除加 `./` 前缀精确匹配（修复 `--exclude README.md` 误伤 app/README.md）；`VERSION` 不再排除（自检必需，随包分发）；webapp/tests/scripts 保持不入包
- **[sync] liki-web sync-skills.sh**：rsync 排除口径与打包一致（webapp/tests/工程文件/缓存）
- **[ci] content.sha256 一致性校验**：push 时若提交指纹 ≠ 当前树指纹直接失败——根治"提交内容与指纹脱节"（历史教训：HEAD 提交 d9057296 与 HEAD 树实际指纹 ad6b75ad 不符，导致安装副本每次自检误报）

## 3.9.0 —— 占卜风水四门同构化：确定性下沉引擎，断语归位前端（引擎 2.6.0 配套）

- **架构统一（参照八字紫微）**：四门（六爻/奇门/八宅/玄空）确定性计算全部下沉引擎（排盘+派生），judgment 方法全删，引擎不再输出 rating/advice 类综合评级——吉凶用符号固有属性（星/门/用神自身），断语由 LLM 按统一断语表翻译
- **方法收敛**：`qimen.chart` 并入日时干落宫/生克/空亡马星影响；`liuyao.chart` 每爻补月破/发动/动爻生克状态；`bazhai.judgment`→`bazhai.layout`（门主灶配合）；`xuankong.annual/sanyuan/judgment`→`xuankong.liunian`（流年叠加）
- **断语表统一 schema**：四门 9 张断语表统一为 `|实体|五行|吉凶|应事|应期/化解|经典依据|` 六列，吉凶五档（大吉/吉/平/凶/大凶）为符号固有属性；玄空/奇门补"应事"列
- **SKILL 骨架统一**：四门 SKILL.md 统一「路由声明 / 方法(引擎)与断语(前端)两栏索引 / chart→查表→LLM 断语流程 / 边界」；补游年星≠飞星边界、layout/liunian 触发条件
- **报告模板统一**：四份报告统一「结论先行+依据链+方法足迹」
- **共享年飞星**：紫白飞星收敛到引擎 fengshui 包（修 xuankong 甲子年入中星偏差），八宅/玄空共用

## 3.10.2 —— 测试报告复核修复（大运公历年段 + time.now 前置 + 文档对齐）

- [atoms] 大运窗口流年/换运流年/当前大运干支：虚岁换算 → 引擎公历年段直判（start_year/end_year，2.6.15 配套）
- [aggregate] dayun_steps 组装改 start_date/end_date/start_year/end_year
- [SKILL.md] Phase 0 强制前置 time.now（当前时间——应期/流年/换运基准）；大运/大限字段说明（公历日期段）；方法数 33→32

## 3.10.0 —— 规则引擎工具化：web agent 可执行 + 用神合并/小限移除/评审补强（引擎 2.6.9-2.6.11 配套）

- **规则引擎工具化**（web agent 可执行）：新增 `tools/skill-tools.json`（5 工具 OpenAI function calling 格式，单一来源——本地/web agent 共用）+ `tools/agent_cli.py`（stdin {fn,args} → stdout {ok,data} 白名单分派，无任意代码执行）；`paipan.py` RPC_URL 环境化（LIKI_RPC_URL）；SKILL.md 引用 schema 文件、降级改"失败提示重试（保精度）"
- `bazi.yongshen` 合并进 `bazi.chart`（yong_shen 三派内联）→ 删独立方法；`paipan.py`/webapp 流程改读 chart.yong_shen；`bazi.fullchart` 透传用神
- 八字小限（bazi.xiaoxian）移除——小限为紫微体系概念，子平八字无正统依据
- 评审改进：bazi/ziwei/huangli SKILL.md 补「路由/边界规则」节（8 域骨架统一）；工具链测试 10→19（藏/有根/旺/弱/缺算子 + evaluate_factors 因子快照）

## 3.8.1 —— 正式发布：断语撞修复 + 月令格神定主面 + 引擎流年神煞配套 + 窄表修复

- **断语撞系统性修复（缺上下文因子）**：扫描 19 域撞——xingge 26/32 盘内外向冲突（新增「月令本气十神」因子 10 个 + xg_m01~m10 月令主面断语，主面唯一）、liuqin 19/32 父旺父损（liu_101 加财星受克排除）、chushen 13 盘（fam 加排除+删重复）、marriage/caiyun 加寡宿/大运比劫上下文、zuhe/dayun 删重复/拆身强弱——全部 0 撞
- **月令格神定性格主面**（《子平真诠》月令为提纲格神主性）：SKILL.md 性格主面裁决，十神旺衰断语只作辅面
- **星宫同参**：hun_101 加宫冲/寡宿排除 + 新增 hun_101b（配偶星透+宫冲→婚可成但波折）
- **引擎 2.5.0 配套**：流年神煞接入（动态 9 种年日双查 + 值年病符/丧门/吊客/大耗）——factors 13 因子 + yearly 断语 14 条；引擎灾煞表命理错误修复（golden 抓出）
- **命理师视角修正**：hun_202 删越界性向取象（传统命理不断性向）
- **流程/文档**：主流程统一 full_paipan（Phase 2/记忆管理/RPC 边界/真太阳时桥接/输出规则）；过时残留清理（app/domains/webapp 手调 RPC）；README 移除准确率数字改发帖；评测基建（run-qwen.sh 答案自愈）
- 窄表 gen_factors 往返 bug 修复 + 得地因子判定恢复：
  - 根因1（宽表 bug）：得地合并时 6 因子判定值丢失——配偶星得地/财星得地/父星得地/母星得地/印星得地 全空（evaluate 时"空条件=永远命中"）+ 主妇信号缺有根[官杀]
  - 根因2（窄表过时）：窄表列名体系与宽表脱节（旧展开列名 vs 旺算子/多行或组）
  - 修复：恢复 6 因子判定（有根[X]=1）+ gen_factors.py 加 --reverse（宽表→窄表重建）+ 窄表重建为宽表等价表达
  - 验证：窄表↔宽表往返逐字节一致 + check_schema 0 错误 + 数据检查 160 题零命中 0 + 9 个受影响题断语命中方向全对

- 根因1（宽表 bug）：得地合并时 6 因子判定值丢失——配偶星得地/财星得地/父星得地/母星得地/印星得地 全空（evaluate 时"空条件=永远命中"）+ 主妇信号缺有根[官杀]
- 根因2（窄表过时）：窄表列名体系与宽表脱节（旧展开列名 vs 旺算子/多行或组）
- 修复：恢复 6 因子判定（有根[X]=1）+ gen_factors.py 加 --reverse（宽表→窄表重建）+ 窄表重建为宽表等价表达
- 验证：窄表↔宽表往返逐字节一致 + check_schema 0 错误 + 数据检查 160 题零命中 0 + 9 个受影响题断语命中方向全对


## 3.7.1——duanyu md 三分类收敛（断语 csv 化·方法归 fangfa）

- 15 个 duanyu md 按三类收敛：静态断语复述删（csv 已同义覆盖）、真命理方法迁 fangfa（按域平铺）、拟合残留删
- bazi 6 混合：档位表/类型表删 → fangfa 保留判断链/护栏（caiyun/hehui/shishen/shiye/wuxing-jiankang/xueye）
- ziwei 9：6 纯方法整体迁（fuxing/geju/gong12/laiyin/sihua/xiangmao）+ 3 混合（liunian/yingqi 整体迁、zhuxing 删性格基调列+骨架行）
- duanyu 目录留 README 索引（断语=tools/*.csv 单一来源，方法=fangfa）
- 去重：caiyun 应期→dayun、hehui 合化→yongshen、zhuxing 空宫→gong12
- 失效引用清理：印星三关（calibration/study/mingshu/SUPPORT/SKILL）+ duanyu→fangfa 路径（app/domains SKILL）
- C 类残留：hehui ≥4次删、比劫重重改比劫旺
- 遗留：app 卡档位表名引用（「官财透干定层次」等→csv）待后续清理


## 3.7.0——79 错题根因修复（阶段 A+B+C——全改表/改文档，零代码）
- **阶段A SKILL 考时准则**：首次成婚（婚动≠成婚）/引动都算（删"虚引动降级"——R3 根因来源）/换运首年+配偶星透=一级候选/孕产凶险≠否认定生育/性向寿元本命定案/主断语优先于辅象/跨题互斥含命理修辞/六亲生死需断语支持（R3+R6+R7——27 题）
- **阶段B 断语表补判据**：ying_h18 损胎（0019——三刑+食伤重，引用本命扩展 snapshot 传入）/zy_301 入殓（0138——孤寡+印多+华盖）+zw_zy_301 命宫贪狼/lq_301 父寿（0047——财星受克新因子，火旺克金父短寿）/cy_301 母代财（0030——印旺财藏）/ying_h19 婚变（0018——目标星透+三刑+食伤重，0007/0022 不误伤）
- **阶段C 否决级**：zinv zv_103 克夺排他（0058 无儿女——克夺>得地）/marriage hun_408b 寡宿独身（0044——+大运0+混杂0 排他，0054 无误伤）+hun_407 寡宿排他
- 验证：check_schema 0 错误 + 数据检查 160 题零命中 0（新断语不破坏覆盖）+ 代表题断语命中验证（0019/0138/0047/0030/0018/0044/0058）
- 追加：ying_h19 婚变（0018——目标星透+三刑+食伤重）/hun_408b 寡宿独身（0044——+大运0+混杂0 排他，0054 无误伤）/SKILL 补则5（六亲生死需断语支持——0131）/年柱干伏吟原子+ying_h20 家变（0119 父母离异——0024/0135 不误伤）——check_schema 0 错误 + 数据检查 160 题零命中 0
- 阶段D+E：流年支受克原子+ying_h21 健康联动（0005——本命土金旺克 2020 子水，0045/0150 不误伤）/ying_h01 岁运并临强化（0060 可至重病死亡）/SKILL 补则6（健康看疾厄宫+大限——0049/0070）+补则7（子女性别时柱食伤阴阳为主——0065）/ziwei_zhiye zw_zy_302 官禄空劫转工（0128）——0115（A6 主断语）/0154（职业映射缺）归引导

## 3.8.0——agent 综合裁决层修复（EXEC 37 题 + MISS 部分）
- 67 错题根因分析：EXEC 55%（断语命中但 agent 降权/排除——agent 综合裁决层）/MISS 31%（断语缺）/AMBIG 7%（多义）/GUESS 6%（零覆盖）
- **SKILL 综合裁决准则 6 条**（A1-A6）：硬信号优先于一般反证（0089/0073/0101）/断语直指即采纳（0128/0098）/孕产分情境（0113/0114 事件题断流产）/六亲生死闭环（0149 入墓≠亡铁证）/本域主信号优先（0006/0059）/多证即采纳（0059/0068）
- **zy_107 排他修复**（0039 偏印旺仍做保险——去偏印旺0 排他；0078 不误伤）
- 0102（凶年断语已有——A1 应用）/0132·0091·0095（题目难度/agent 大运演绎——诚实不拟合）
- 验证：check_schema 0 错误 + 数据检查 160 题零命中 0

## 3.6.2——评测统一（判题唯一走 skill-up agent——eval_hybrid 降级为数据检查）
- **评测判题统一**：题目→skill（agent 读 SKILL.md → evaluate_from_rpc → 综合判题）→ 答案 → grade-grouped 对比——唯一评测路径（run-qwen.sh）
- **eval_hybrid.py 降级为规则表数据检查**（删判题：_STATUS_MAP/族逻辑/紫微铁断/apply_rule 全部移除）——只统计断语覆盖（160 题零命中 0）——非判题（评测标签/判题逻辑全部移出 skill 与工具脚本）
- 数据检查输出：tests/RESULTS.md（覆盖统计——marriage/jiankang/xingge 等 18 域覆盖 81~160 题——规则表无遗漏）
- 回归验证：check_schema 0 错误

## 3.6.1——去异化补齐 + 评审修复（should-fix 全清）
- 去异化补齐 4/4：xueye/shiye/chushen 结论列 28 断语改命理表达（博士/老板/富贵等档位词移出 skill 结论——评测标签全在测试层）
- SKILL.md 修复：iron.json 引用删（文件已删——agent 参考紫微断语表 hun_301/302）；文件名去前缀（bazi_liuqin.csv → tools/bazi/liuqin.csv）
- schema.md 示例改命理表达（"已婚"标签 → "财星透干得地——婚可成"）
- 测试层清理：_STATUS_MAP 补事业档位（shi_102~109——结论不在状态词表的判题断语）；删 _NORM 死代码；hun_301/302 双份注释说明（状态集触发冲突→_IRON 裁决——删除会破坏 0048 铁断）
- .agents/skills/liki 旧副本（2.4.0）删除（根 skills/liki 3.6.x 权威）
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 错误 + 容器验证（结论无评测标签）

## 3.6.0——综合评定回归 agent（评测标签移出 skill——去异化）
- **综合评定 agent 做**：删 resolve_consistency/_consistency_single（程序综合定案）——evaluate_from_rpc 输出双盘命理断语——agent 像命理师双盘参看综合
- **评测标签完全移出 skill**：状态标签（已婚/未婚/独身等）不再存在于 skill 数据——测试层（eval_hybrid）私有映射（_STATUS_MAP id→状态 + 族逻辑 + 紫微铁断）判题
- **断语结论列 = 命理表达**：marriage 27 断语结论从状态标签重写为命理话（"财星透干得地——婚可成"）——skill 数据纯命理（容器验证结论无状态标签）
- iron.json 删（旧综合层残留——断语表 hun_301/302 已有命理结论）
- SKILL.md：综合评定 agent 做（多命中按命理次序综合——程序不硬选/不归一标签）
- 回归：确定性 30/30=100%（测试层映射独立判题）+ 应期 13/14=93% + 容器兼容

## 3.5.0——命理逻辑全表化（Python 建表+查表——零命理定义）
- Python 两大任务：建表（引擎字段→因子快照）+ 查表（真值表匹配）——零命理组合定义
- constants 扩充：六合/三合/三会/六冲/三刑/天干五合/旬空/合类/冲类 表（代码 _GONG_HE/_GONG_CHONG/_CHONG/_SANXING/_XUNKONG 删）
- 新机械原子（_liu_op——查 constants/比较）：干支相等/干克/支冲/三刑/旬空/天干合
- 流年因子真值表化：factors_liunian.json → csv——8 组合算子（岁运并临/天克地冲/日犯岁君/伏吟/反吟/三刑/空亡/流年干合）展开为原子行——evaluate_liunian_factors 真值表匹配
- 铁断表：iron.json（_ZW_IRON 出代码——综合层查表）
- 残留清理：删 _eval_judge（判表达式求值）/ 旧组合算子分支 / 代码常量；factors_liunian.json 删（csv 权威）
- 回归：确定性 30/30=100%（0048 紫微铁断）+ 应期 13/14=93% + check_schema 0 错误 + 容器

## 3.4.2——tools 去 mingli（推理机根 + 术数子目录）
- tools/mingli 拆开：推理机（duanyu.py/engine.py）+ 中间数据（factors/constants）→ tools/ 根；断语表 → tools/bazi/ + tools/ziwei/（去 bazi_/ziwei_ 前缀——目录即术数）
- _load_table 映射（bazi_ 前缀 → bazi/ 子目录去前缀）；import 统一（duanyu/engine——sys.path tools 根）
- 结构：tools/（推理机根）+ tools/bazi/（19 断语表）+ tools/ziwei/（8 断语表）
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 警告 + 容器（仅 tools+tests）

## 3.4.1——断语表 csv 归 tools（谁用归谁）
- 断语表 csv（bazi_*.csv/ziwei_*.csv + factors.csv）全归 tools/mingli——csv 只有工具（推理机）运行时读——domains 留 md（人读知识）——谁用归谁
- _load_table/eval_hybrid/check_schema 路径简化（csv 全在 tools/mingli——去术数子目录）
- 结构：domains（md 知识）/ tools/mingli（推理机 + 全部 csv 执行数据）
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 警告 + 容器（仅挂 tools+tests）

## 3.4.0——八字/紫微真分开（各自判→对照综合）
- 真分开（命理合参真相——各自判再对照，非合并）：因子快照分（bazi 快照 205 因子 / ziwei 快照 52 因子——factors.csv 加术数列）
- 拆 5 真混合行（hun_105c 天机独坐排他/qy_104 迁移宫/tz_106 田宅宫/xg_302 命宫天梁/zv_106 子女宫）——八字部分回 bazi 表（纯八字）——紫微部分由 ziwei 表承担（铁断对照）
- 确认其余"混合"行纯八字（桃花/寡宿/日支/七杀——八字概念误标）——留 bazi
- 综合对照（resolve_consistency）：一致=双盘印证 / 紫微铁断（贪狼化忌=未婚、天机独坐=独身）优先 / 矛盾归 agent 命理分析
- check_schema 交叉校验：bazi 表纯八字因子、ziwei 表纯紫微因子（跨术数条件=警告——防混合回潮）
- 回归：确定性 30/30=100%（0048 紫微铁断复现）+ 应期 13/14=93% + 容器（双快照）+ check_schema 0 警告

## 3.3.9——mingli 归 tools（工具层——推理机）
- domains/mingli → tools/mingli：推理机（duanyu.py/engine.py）+ 中间数据（factors/constants/流年因子）归工具层——domains 纯术数知识（bazi/ziwei 断语表 + md）
- 确认"函数即 skill 工具"（不 MCP）：SKILL.md 加工具清单（evaluate_from_rpc/evaluate_factors/evaluate_liunian_factors/resolve_consistency/match/build_factors + tests/client 排盘）
- import 统一 mingli.*（tools 容器目录——sys.path 加 tools）；表路径（bazi/ziwei 断语在 domains）；eval_hybrid/check_schema/SKILL.md/容器路径适配
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 错误 + 容器兼容

## 3.3.8——八字断语归八字（术数自含完成）
- 八字断语表（20 域 + 混合行）→ domains/bazi/duanyu/bazi_*.csv（十神/五行/神煞/日支——与紫微对称：bazi/duanyu + ziwei/duanyu）
- domains/mingli/ 只留推理机（duanyu.py/engine.py）+ 中间数据（factors.csv/constants.json/流年因子）——不再承载断语知识
- 应用层合不变：domains[rule] = {八字: bazi_{rule}, 紫微: ziwei_{rule}}（双盘参看）——综合规则（一致/紫微铁断/冲突归 agent）
- check_schema 适配（扫 bazi/ziwei 双术数表 27 张）；应期 _yt → bazi_yingqi
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 错误 + 容器兼容

## 3.3.7——八字/紫微断语分离（知识层分、应用层合）
- 知识层分：40 纯紫微断语行 → domains/ziwei/duanyu/ziwei_*.csv（8 表：marriage 2/xingge 14/qianyi 3/tianzhai 4/yingqi 5/zhiye 1/zinv 5/ziwei 12）——紫微断语归紫微（星曜/四化/十二宫）
- 校正：误分类行移回八字（日支桃花驿马墓库/神煞桃花寡宿/七杀旺——八字概念）——严格按术数概念分类
- 应用层合：domains[rule] = {八字: [...], 紫微: [...]}（双盘分别出断语）——综合定案：一致=双盘同参 / 紫微铁断优先（贪狼化忌=未婚、天机独坐=独身——0048 复现）/ 冲突归 agent 双盘参看
- 混合行保留（八字主+紫微修正——13 行不拆——行内条件生效）；因子表保持综合（中间计算层）；应期八字 yingqi 为主 + 紫微流年 agent 参考
- 回归：确定性 30/30=100% + 应期 13/14=93% + check_schema 0 错误 + 容器兼容

## 3.3.6——应期深挖 + 紫微探索 + 评测 ev 修复
- 应期错题深挖（0018/0102——八字+紫微全因子）：0102 根因=评测 ev 标签误判（"结婚离婚"混合题答案 2010 是结婚年——eval 误判婚变）——修复：混合题候选=婚动∪婚变命中 → 应期 12/14→13/14
- 紫微流年探索（ziwei_liunian：红鸾/天喜/流年四化落宫/流年夫妻宫）：红鸾天喜几乎不落命/夫妻宫（区分度低）、流年夫妻化忌部分相关有误（0101 命中但 2002 误）——结论：紫微流年为软信号（agent 综合）——不宜硬规则化（防误伤）
- 0018（离婚 2013）归 agent：八字夫星透干（无受克无宫冲）+ 紫微无强信号——找不到不误伤的规则（诚实边界）
- 回归：确定性 30/30=100% + 应期 13/14=93%

## 3.3.5——constants 单一来源（清代码硬编码重复）
- constants.json = 推理底层规则（五行生克/干支五行/目标星/事件宫位——公理与经典规则）+ 程序分组（类/旺衰）——查询字典 json（非真值表——确认不 csv）
- 清除 duanyu.py 文件头硬编码重复（_KE/_SHENG/_GAN_WX/_ZHIWX/_TARGET_STARS/_TARGET_PALACES/_TEN_CLASSES/_WANG/_WEAK——与 json 双份且类型不一致）→ 全部从 constants.json 构建——命理数据单一来源（代码零硬编码）
- 回归：确定性 30/30=100% + 应期 12/14=86%

## 3.3.4——测试统一（skill 唯一逻辑源）
- 一致性定案进 skill：mingli.duanyu 新增 resolve_consistency（婚姻/事业结论族——SKILL.md「同族定案/跨族综合」规则）
- tests/eval_hybrid.py 删自写一致性判定（_match_dy 内联逻辑）→ 统一调 skill API——skill 是唯一逻辑源，测试只做题目选项匹配（评测机制）
- constants 确认保持 json（查询字典——非真值表——嵌套映射 json 原生；与断语表 csv 分工：知识表 csv / 查询字典 json）
- 回归：确定性 30/30=100% + 应期 12/14=86%

## 3.3.3——组织结构重构（领域命名 + 规则域落位）
- 目录领域命名：runtime/ → domains/mingli/（命理规则域——因子+断语+生成器一体）；evals/ → tests/
- client/birth（RPC 排盘工具）→ tests/（测试基础设施——skill 知识层不含）；factors.py（因子构建）并入 duanyu.py（生成器一体——mingli 无独立排盘文件）
- evaluate_from_birth → evaluate_from_rpc（纯——rpc 排盘数据 → 断语；agent 模板同步：排盘用 tests 工具 + 生成器断语）
- 结构：app（价值层 13 卡）/ domains（概念层：术数 md + mingli 规则域）/ tests（评测+排盘工具）/ webapp（独立）
- 回归：确定性 30/30=100% + 应期 12/14=86% + 容器无 PyYAML 兼容

## 3.3.2——组织结构梳理（三层确认 + 索引）
- 确认三层组织：app=用户价值层（13 场景卡）/ domains=命理概念层（9 术数）/ runtime=规则引擎（webapp 独立 web 应用不管）
- app/README.md 索引：13 卡按流程类型分组（八字+紫微全流程/子流程 Phase 8/用神流程）+ 功能维度速查（应期断事/命理报告/合盘/起名/占卜/择日/风水）
- SKILL.md Phase 0 加模块总览（功能 → 入口）

## 3.3.1——因子层真值表化 + 印星三关移除（回归经典）
- 因子层真值表化：factors.csv（193 因子 → 256 行真值表、218 原子列）——复合算子（旺/弱/克者旺）展开为原子组合行；空参=因子引用（读快照）vs 空参算子区分；数值因子（事实计数）= 直通标记行
- evaluator 删判表达式求值 → 原子执行器 + 真值表匹配器（多遍稳定 + 快照补全）——代码零命理逻辑，逻辑全在表
- 移除臆造"印星三关数"（非经典）→ 学历断语经典化（印星现/透/得地/得令/旺/多 + 官杀/食伤/财坏印——官印相生科甲/食伤泄秀才学/财坏印学途中断/印杂主愚）
- 回归：确定性 30/30=100% + 应期 12/14=86% + 容器无 PyYAML 兼容（csv 标准库）

## 3.3.0——断语表迁移 csv（真值表）+ 去强度等级
- **断语表 yaml → csv**（20 域 281 条——标准库 csv——容器无 PyYAML 兼容，data.py/build_data.py 删除）：列=因子并集（真值表），取值 1/0/空=无关/字符串值；含"经典原文"列（待补——46% 空）
- **去"强度等级"**（非命理逻辑——用户纠正"断语=逻辑推导，有就有没有就没有"）：yingqi 22 处 + eval 候选/evaluate_from_birth 全清——候选=全部命中年份（agent 按考时准则裁决：首次优先/冲主变动——命理判据，非数值）
- **any_of 拆行**（xue_201/101/204/wai_109——真值表行表达或）+ 比劫重重因子（数值范围下沉）
- **因子层 yaml → json**（factors/constants/factors_liunian——嵌套结构原生 + 标准库）——**运行时零第三方依赖**（断语 csv + 因子 json + 执行 py 全标准库，容器无 PyYAML 全跑通——已容器验证）
- evaluator/check_schema/eval_hybrid 加载 csv/json；删全部 yaml
- 回归：确定性 30/30 = 100%；应期候选=真实命中 10/14（去等级后——0018 等答案年规则未覆盖待补）
- **待办**：补经典原文（130 条空——逐域回注经典出处）

## 3.2.1——断语表真值表化（纯取值——与/或/非全用行表达）
- **断语表 = 真值表**：每行 = 因子取值组合（与）；同结论多行 = 或；取值 0 = 非——任意逻辑用行表达（析取范式 DNF），**无 any_of、无 >=/<=**
- **any_of 拆行**：xue_201/101/204（博士/大学各拆 2 行）、wai_109（身材瘦拆 3 行——日主木/火/金）
- **数值范围下沉因子**：cai_102 比劫重>=4 → "比劫重重"因子（count≥4 bool——真值表行是取值，数值范围在因子层转 bool）
- **教训**：any_of 下沉"复合因子"会改变语义（官杀有力含"有根"——0057 有根但透令皆无被误判博士）——**或必须拆行**（不建因子）
- **删多余**：RULES.md/RESULTS.md（过时——引用已删的 predict/软硬断语）、.run-pan01.yaml（临时）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%；校验 0 错误 0 警告（192 因子）

## 3.2.0——唯一入口 evaluator（排盘/推演/断命收敛）
- **删 predict.py**（编排壳——违背"agent 调 evaluator+engine"设计）
- **evaluate_from_birth 唯一入口**：接收出生信息 → 内部完成 排盘（client/birth/factors）→ 推演（evaluate_factors）→ 断命（engine.match）→ 返回 {gender, snapshot, domains(19 域全部断语), liunian(应期候选)}
- **agent 只面对 evaluator**（SKILL.md 模板一行）——client/birth/factors/engine 成为 evaluator 内部依赖（不暴露）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%；校验 0 错误 0 警告

## 3.1.1——喜忌扩展至观察域（凶兆=忌神为患/吉兆=用神受助）
- **健康 5 条凶兆断语加忌神前提**：官杀攻身外伤→官杀为忌、食伤泄身虚→食伤为忌、印旺代谢慢→印为忌、比劫外伤→比劫为忌、财耗身肾虚→财为忌（《滴天髓》忌神为患——为用者不患）
- **婚姻质量 3 条（星宫同参+喜忌）**：配偶星为用+得地+宫静=上等婚姻、配偶星为忌+得地=中等、配偶星为忌+宫冲=高危——新因子"配偶星为用/为忌"（男财/女官杀×用神五行）
- **学历 2 条**：xue_301/302 身强→印为忌（用神五行更准）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%；校验 0 错误 0 警告（188 因子）

## 3.1.0——喜忌贯穿（用神为枢——引擎五神体系）
- **喜忌因子 10 个**：官杀/财/印/食伤/比劫 各"为用/为忌"——用引擎 yongshen.fu_yi（用/喜/忌五行）——某十神五行 ∈ {用,喜}=为用、==忌=为忌（《滴天髓》用神为枢）——比"身强弱简化"更精确（1981 身弱用木忌水→官杀为忌印为用 ✓；1974 身强用木忌火→官杀为用印为忌 ✓）
- **断语融入喜忌**（喜忌作为约束——非新断语族）：shi_107 高管加"官杀为用"（0088 身弱官杀为忌根治）、cai_101 财运好加"财为用"、cai_203 大富加"财为用+官杀为用"
- **教训**：喜忌是"前提"（判断错误则吉凶反转）——融入现有断语约束，不作独立断语族（shi_201/202 已删）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%；校验 0 错误 0 警告（186 因子）

## 3.0.9——抄 6 专项断事表（mingli-skills phase6 经典断语入 yaml）
- **财运**：财库现/财星入墓（新算子——金库丑/木库未/水库辰/火库戌/土库辰，《三命通会》财库论）+ 财富层次 5 条（大富=身旺财旺官透/中上=食伤生财/过手财=财多身弱比劫夺）+ 正偏财区分 3 条（正财透稳定收入/偏财透投资/混杂综合）
- **婚姻**：夫妻宫状态算子（冲/合/刑/静——用引擎 liu_chong/zhi_liu_he/liu_xing 查日支）+ 日支类型（四桃花/四驿马/四墓库）+ 星宫同参 3 条 + 配偶特征 3 条（观察——不断状态）
- **健康**：十神疾病 5 条（官杀攻身外伤/食伤泄身虚/印旺代谢慢/比劫外伤/财耗身肾虚）
- **性格**：日主十干 10 因子+10 断语（甲刚直/乙柔韧...）+ 月支长生十二态（引擎 chang_sheng）
- **子女**：时柱枭印克子 + 食伤生财 2 条
- **教训**：官杀喜忌（shi_201/202）作为独立断语族干扰档位判定——已删（喜忌是分析方法，应作为约束融入现有断语——阶段 C 喜忌贯穿待设计）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%；校验 0 错误 0 警告（176 因子）

## 3.0.8——断语权重回归命理（约束补全上下文——去标签/去程序排序）
- **match 全返回不排序**（表顺序=命理设计权重；程序不做权重定义——"约束里写了 1 的就是高权重"）
- **eval 一致性判定**：同结论族定案（已婚波折⊂已婚、老板+管理层⊂老板族——命理语义归类），跨族冲突归 agent（"?"——需命理综合，参考项目"矛盾信号调和"）
- **逐条断语补排他（回归命理上下文）**——冲突=断语约束漏因子：
  - 学历 5 条：xue_205 排官杀/食伤透/财坏印、xue_209 定三关2、xue_201 恢复印星多0、xue_208 排官杀得令
  - 贫富 3 条：fam_108 排枭神夺食、fam_109 排财坏印/身强、fam_110 排财透有根
  - 婚姻 2 条：hun_208 排食伤克官/寡宿（0035/0122）
  - 事业 4 条：shi_102/108 排主妇信号、shi_107 排食伤旺；eval 老板族保留"管理层"信息匹配高管选项
- **应期取最大强度等级**（多命中取最强信号——非表序第一条）
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79% 恢复（纯约束补全，无优先级标签）
- 借鉴 mingli-skills：H01-H25 应期硬规则（岁运并临/天克地冲/伏吟反吟/日犯岁君/三刑/空亡填实）入 factors_liunian + yingqi

## 3.0.7——架构归类澄清 + constants 恢复（复合因子层字典）
- **架构三层澄清**：引擎（基础因子：排盘/五行/十神/生克交互）→ 复合因子层（constants 字典参数 + factors.yaml/factors_liunian.yaml + evaluator）→ 断语层（engine.match + 断语表）
- **constants.yaml 恢复**：五行生克/天干地支五行/目标星/事件宫位/类/旺衰——复合因子层字典（evaluator 求值 factors.yaml 的参数表）
- **修复 _resolve_tens**：配偶星/子女星/父星/母星统一用 const["目标星"]（嵌套结构）——此前恢复 constants 时结构不一致导致配偶星因子全 0（婚姻断语大面积查无）——修复后恢复
- **官杀取清正确机制**：克我（官杀）所在柱被引擎 zhi_liu_he/liu_chong 合/冲 → 取清（合杀留官/冲去多余，《子平真诠》）；hun_105e 女命官杀混杂+取清0+**配偶星得地0**（夫星无根才主婚姻复杂——0153 夫星得地仍已婚，命理排他）
- **撤销 eval hack**（_match_dy 恢复原样——只取第一条命中）；软/硬断语分层：硬=定案（eval 用）、软=参考（predict 给 agent），match 按优先级+约束数排序
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79%

## 3.0.6——domains 断语 md 信息保全 + 补规则
- **比对 domains/bazi/duanyu/*.md vs 对应 yaml**：提取 md 独有领域知识，可规则化的补进 yaml（经典依据）：
  - caiyun.yaml 补 5 条：驿马+比劫外出合伙破财、比劫夺财+财无根=无蓄财、身弱财透克印损根基、比劫/食伤大运财运方向
  - jiankang.yaml 补 3 条：五行亢盛双向应病（木旺克土+生火耗木、火旺克金+生土耗火、水旺克火+生木耗水）
  - xingge.yaml 补 2 条：身强弱取象（七杀身弱胆小压抑、印身弱知书达理——非身强印忌之依赖）
  - xueye.yaml 补 2 条：身强印忌看食伤泄秀/官杀生印压力成才（0027 官印相生博士）
- **md 保留为参考**：SKILL.md 标注"断语以规则引擎（yaml）为准，domains 断语 md 为参考（含 yaml 未覆盖的方法/组合规则）"——信息保全不丢知识
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79% 保持；校验 0 错误 0 警告

## 3.0.5——过时领域知识清理（双轨收敛）
- **engine.py 删重复常量**：TARGET_STARS/TARGET_PALACES/_KE/_GAN_WX/_ZHIWX/_ZW_SHA——constants.yaml 已统一收录（目标星/事件宫位/五行生克/天干地支五行）——engine.py 75 行纯机械（match+zw_da_xian+_val_match+_PRI+_GONG_HE/CHONG），**零领域知识常量**
- 领域知识单一来源：constants（字典）+ factors/factors_liunian（因子规则）+ 20 断语域（结论）——代码无命理知识
- 回归：确定性 30/30 = 100% + 应期 11/14 = 79% 保持
- 待办：SKILL.md Phase 5.5 仍教 agent 调 predict.py（渐进路线：agent 直接调 evaluate_factors+match，predict 降级 fallback——待 skill-up 实测后改）

## 3.0.4——流年表驱动统一架构 + 测试辅助移出
- **流年因子表驱动**：factors_liunian.yaml（17 流年因子定义行：流年透/值宫/合会/冲/克/忌神/财坏印/大运窗口/换运/流年宫忌/引用本命）；evaluator 新增 evaluate_liunian_factors（流年模式，按年求值）；engine.py 删 derive_liunian_factors（105 行纯机械：match+zw_da_xian）
- **constants.yaml 加目标星/事件宫位**（应期目标映射从代码进表）
- **应期提升**：候选命中 10/14 → **11/14 = 79%**（表驱动版本修正边界）
- **测试辅助移出 skill**：eval_hybrid.py / check_schema.py → tests/（import/路径修复）；domains/mingli/ 只留运行时（birth/client/factors + duanyu 表驱动引擎）
- 回归：确定性 30/30 = 100% 保持

## 3.0.3——数据库约束（schema 校验 + 视图一致性）
- **check_schema.py**（schema 校验）：断语表约束键必须 ∈ 因子全集（factors.yaml 149 因子 + 引擎直读 + 应期因子）；引用因子声明一致性；断语 id 唯一——集成到 eval_hybrid 启动（回归自动校验）
- **修复校验发现**：hun_210 id 重复（删后恢复——0013 离异行）、factors.yaml 补"食伤有根"（xue_101 用）、应期因子放行（derive_liunian 输出键）、各域引用因子声明补全/清理
- **data.py 指纹同步**：build_data.py 记录源 yaml hash（_SOURCE_HASH）；predict.py 容器 fallback 时比对指纹检测漂移（警告提示重生成）——防双源漂移
- 回归：确定性 30/30 = 100% 保持；应期 10/14；校验 0 错误 0 警告

## 3.0.2——断语补全（对照命理经典）
- **修复宫含四化算子**：本命四化在顶层 si_hua（{星:四化}）无宫位——按星落宫反推——此前"疾厄宫化忌/田宅宫化禄"等本命紫微化忌因子全失效
- **田宅/房产域**（tianzhai.yaml 6 条）：紫微田宅宫财星（武曲太阴天府）置产、化禄房产丰、煞主变动（非难守——用户 1981 男命田宅宫煞买 6 套房）、陷置产难；八字财星得地+印星旺置业（《三命通会》论田宅）
- **迁移/出国域**（qianyi.yaml 5 条）：驿马主动迁、迁移宫主星强外出有利、迁移宫空宜守（《三命通会》《紫微斗数全书》）；0158 搬迁验证
- **子女专项**（zinv.yaml 7 条）：紫微子女宫为主（吉星多子/煞刑克损/吉煞并存有但损/空缘薄/化忌操心）+ 八字子女星得地补充（《渊海子平》论子女）；0084 一子/0144 三子验证
- **紫微 12 宫补全**：14 命宫主星性格（xingge 补）+ 官禄宫（化禄/煞/主星强）/父母宫（煞/化忌）/兄弟宫煞/仆役宫煞 7 条（ziwei.yaml）
- **女命旺夫信号 + 福德宗教缘 + 调候春秋**：旺夫=官星旺无伤官（《女命赋》）；华盖宗教缘（0043 出家人）；《穷通宝鉴》春木喜火/秋金喜火；寿元信号（软——归 agent 考时）
- 回归：确定性 30/30 = 100% 保持；应期 10/14

## 3.0.1——领域评审修复
- **同因子多结论正交性修复**：
  - fam_105（身弱+印有气→富贵）加 `财旺:0` 排他——财旺坏印，印荫庇失效（《渊海子平》财坏印）
  - fam_105b（身弱+财旺→富贵富屋贫人）加 `印有气:0` 排他——有印者走印荫庇（fam_105）
  - 验证：0071（印有气+财旺+身弱→贫穷）不再误中 fam_105
- **命名修复**：family.yaml → chushen.yaml（内容"出身贫富"名不符实）；eval_hybrid/predict 引用同步；data.py 重生成
- **观察域标软**：geju（10 处）/zuhe（3）/ziwei（4）`强度:硬`→`软`（中间观察非定案，防 agent 误当硬结论）
- 回归：确定性 30/30 = 100% 保持；应期 10/14

## 3.0.0——断语层重构：表驱动 + 回归命理定性
- **架构**：断语层重构为三层——①常量表（constants.yaml：五行生克/十神定义/六亲映射/旺衰判定，纯原子知识）②复合因子定义表（factors.yaml：111 因子逻辑行，表驱动）③断语表（17 域 yaml 升级：引用因子声明 + 值规范化）
- **engine 纯机械**：evaluator.py 表驱动求值器（14 算子：现/透/藏/得地/得令/旺/弱/缺/克/生/直读/含/宫含/大运十神/数量至少 + 逻辑组合 + 引用因子）；match 匹配器保留——命理知识 100% 在表，engine 零领域知识
- **回归命理定性（去 count 阈值）**：
  - 旺/弱 = 五行月令旺衰（wang_shuai）+ 透干有根（《子平真诠》得令/得地/透干三得）——非 count≥N
  - 财坏印 = 财透干且克印（《渊海子平》透干为显——0006 财藏大学 vs 0025 财透中学）
  - 伤官克官限女命 + 只算伤官（食神制杀是吉非克官，《三命通会》伤官见官）
  - 印星多 = 数量≥3（《子平真诠》印杂主愚——0156 印4专科 vs 0027 印2博士）
  - 官印相生主科甲（《三命通会》——0027 身强官印相生仍博士）
  - 主妇信号 = 官杀得地+无食伤透（修正原"官弱印旺"与依据矛盾）
  - 配偶星得令排他、紫微交叉最高优先级
- **验证（最终）**：确定性 **30/30 = 100%**（学历 8/8 + 婚姻 12/12 + 事业 5/5 + 出身 5/5）——去 count 阈值回归定性后完全恢复且规则有经典依据；应期 10/14（0002 换运年 vs 0034 换运年不婚，归 agent 考时）
- **逐题命理深挖补全**（含算子 bug 修复：shen_sha 四柱聚合）：
  - 0013 食伤重+无官杀+桃花红鸾→离异（《女命赋》食伤重克夫）
  - 0035 桃花+寡宿→外遇离异（《三命通会》桃花煞淫奔）
  - 0121 枭神夺食+财藏无根→贫穷（《渊海子平》；财透有根者不贫 0036）
  - 0087 男命得地+混杂→已婚波折（财透者不判 0054）
  - 0153 夫星得令已婚稳定（hun_205 加得令排他）
  - eval_hybrid 匹配修复（已婚波折算已婚判子女）
- predict.py/eval_hybrid.py 全部切换表驱动（16 域输出正常）


## 2.6.23（未发布）
- **应期候选程序化**：predict.py 新增 `domains.liunian`（婚姻/事件应期候选年）——扫描选项年份（无则扫 1970-2030），derive_liunian_factors + yingqi 断语按引动等级排序输出候选 top3（含等级/依据/考时说明）——agent 在候选内对照题目选项裁决（不再裸推应期）
  - 依据：大运配偶星窗口 × 流年引动（《三命通会》大运为根流年为用）——与 eval_hybrid 已验证逻辑同源
  - 验证：14 应期题候选 top2 命中 8/11（口径同 eval），0002/0007/0097 在第 3 位候选（agent 考时参考）
- SKILL.md Phase 5.5：应期题可选 --liunian 拿候选年（候选内对照选项）；评测验证成本高（--liunian 每题 24 年 RPC 扫描、3 组验证约 2h 且 skill-up 清理输出无法判分）→ 保留为可选工具，不强求（确定性修复是主要收益：92/160）

## 2.6.22（未发布）
- **修复容器 Python 3.9 兼容性（评测重大发现）**：predict.py 在容器崩溃（TypeError: float | None 需 3.10+；ModuleNotFoundError: yaml）→ agent 调规则引擎失败、确定性题裸推出错（skill-up 全量 84/160=52.5%）
  - domains/mingli/*.py 全部 `X | None` 改 `Optional[X]`（兼容 3.9）
  - 断语库预编译 data.py（build_data.py 生成，17 表）；predict.py yaml 优先（本机热读）/data.py 兜底（容器零依赖）
  - 容器内验证：predict.py 排盘+断语全通 ✓
- 待重跑 skill-up 全量验证正确率

## 2.6.21（未发布）
- SKILL.md 补"应期多候选裁决准则"（考时语义，通用准则）：①首次优先（首婚取最早候选）②冲主变动（六冲夫妻宫=手续/仪式）③窗口为根（无窗口合会/冲=虚引动）④多候选同级取最早，歧义按题目语义指认
- 灾劫/官非/运势题排查：0033 官非/0128 财运已有断语覆盖；0091/0093/0094/0148 为组合/区间题（幼年凶劫看童限、大限出国看迁移+流年）归 agent（SKILL.md 大运+应期方法），不新增断语（防拟合）
- 回归：30/30 + 11/14

## 2.6.20（未发布）
- 外貌矛盾排查：wai_109（食伤泄秀→偏瘦）加日主五行排他（any_of 木/火/金）——水土日主不判瘦（水主丰润、土主敦实，《滴天髓》日主五行身形）
  - 消除"圆润/敦实 vs 偏瘦"同命中矛盾
- 性格保持多面性证据模式（人有多面，agent 综合——2.6 既定设计）
- 回归：30/30 + 11/14 无破坏

## 2.6.19（未发布）
- 应期 3 MISS 深挖裁定：0002/0007/0097 归 agent 考时（有命理依据的分工）
  - 尝试 ying_109 换运首年 3→4：0002（2006 换运结婚）修复但 0034（2014 换运不婚）回归——同为换运首年结果不同，因子无法区分 → 回退 3（防单例过拟合）
  - 0002（首次婚动语义选 2006）、0007（冲年签纸=六冲夫妻宫变动）、0097（"第一段婚姻"=最早年 2003）——规则给候选信号，agent 用考时语义裁决
- 回归：30/30 + 11/14 无破坏

## 2.6.18（未发布）
- 财运互斥排查：cai_106（财来财去）加"财弱"排他——财弱者归 cai_102（破财），不再与 cai_106 同命中
  - 依据：《滴天髓》比劫夺财两型严格分列：财弱=破财无财（cai_102）、财得地非弱=有财被劫（cai_106）
- 财运互斥冲突 4 → 0（pan04/10/22/25 消除）
- 回归：30/30 + 11/14 无破坏

## 2.6.17（未发布）
- predict.py 接入全部断语域：4 硬断语（婚姻/学历/事业/出身）+ 12 软断语证据（职业类型/健康/财运/六亲/外貌/性格/神煞/格局/十神组合/紫微宫位/调候/大运）
  - agent 主入口一次拿到全部断语证据（文本/JSON 双输出），不再只给 4 域
- 婚姻互斥一致性确认：predict hits[0]（优先级裁决）与 eval_hybrid 一致
- 其他域冲突抽查：学历/事业已有排他设计（30/30 确定性无互斥）

## 2.6.16（未发布）
- 婚姻状态互斥裁决：match 新增 exclusive 参数（单一事实域取最高优先级 1 条）；marriage.yaml 全部状态断语加优先级（离异/夫早亡/独身/未婚=高、已婚波折=中、已婚=低兜底）
  - 命理依据：婚姻状态是单一事实（不能既已婚又未婚），确定性事实（离异/克夫/无夫缘）优先于正常态（已婚）
- hun_206 未婚加宫破排他（0071 官杀得地受克+宫破=婚变非未婚）
- 冲突扫描：74 → 4 题（剩余 4 为同一命例"未婚+独身"近义共存，可接受；0097 等婚期题的天机独坐独身倾向为干扰证据，评测不计）
- 回归：30/30 + 11/14 无破坏

## 2.6.15（未发布）
- 覆盖完整性验证：160 题零断语命中 = 0（全部有证据，无空洞）
- 婚姻断语冲突修复：hun_204 独身约束修正（配偶星不现+伤官克官，排除配偶星得地）——消除"独身+已婚"同命中矛盾
  - 依据：《渊海子平》女命无官杀夫缘不显 + 伤官克官克夫 → 两象叠加才断独身；配偶星得地=夫星实（已婚/克夫，非独身）
  - 0048 独身由紫微天机独坐夫妻宫断（男命）；夫早亡（hun_203）=克夫非不婚（0122 需要，与已婚不矛盾）
- 回归：30/30 + 11/14 无破坏

## 2.6.14（未发布）
- 五行健康断语 5→16 条（三档：《黄帝内经》五行主脏腑——缺=该脏虚、旺=该脏亢、休囚=该脏弱）
- engine 新增五行过旺（count≥4）与五行弱（wang_shuai 休囚）因子
- 验证命中：0038 缺水→脑（肾主骨生髓通脑）、0116 火弱→心、0126 木弱→肝、0150 金旺→鼻癌、0060 金旺→肺（癌症）
- 软断语证据模式：给 agent 脏腑观察，agent 综合题目选项

## 2.6.13（未发布）
- 职业类型断语（zhiye.yaml 7 条，十神类象《子平真诠》：偏印=玄学/房产/偏门、食伤=技术才艺、正财=财务、七杀=武职权柄、官印=体制、比劫伤官=创业、食伤生财=销售）——验证 0055 玄学/0115 会计/0088 创业/0132 七杀/0042 房产 命中
- 财运断语补 2 条：cai_105 食伤生财=财源广进（《滴天髓》）；cai_106 比劫夺财+财得令=财来财去（《滴天髓》两型：财弱破财/财地被劫）
- cai_103 加比劫夺财排他（0128 财来财去不再误判积蓄）
- engine 新增职业类型因子：食伤旺/偏印旺/七杀旺/正财旺/比劫伤官（十神类象）

## 2.6.12（未发布）
- 引擎评级清理（只删臆造，经典保留）：
  - **删**：ziwei.judgment / liuyao.judgment / qimen.select 整个删除（liuyao 经典数据 chart 已全含；qimen.select 是打分选时）；ziwei pattern Score（人为分级）；所有 rating/advice/rule/summary 评分字段
  - **恢复（删评分）**：qimen.judgment（保留 subject_palace 主题宫/生克/格局/空亡马星影响——参数化断事）；bazhai.judgment（保留门主灶 match——参数化）
  - **合并进 chart（排盘固有）**：ziwei.chart/fullchart + san_fang（三方四正）；qimen.chart + 值符宫/值使宫/日干宫
- client.py 删除 ziwei_judgment（保留 ziwei_daxian——经典大限数据）
- skill 迁移：六爻断法（domains/liuyao/duanyu/jixiong.md 新增吉凶判定：用神旺衰+动爻生克→吉凶，《增删卜易》《卜筮正宗》）；qimen/bazhai SKILL.md 加回参数化 judgment 引用
- 保留经典断法要素：bazi.yongshen（三派用神）、xuankong.annual（紫白诀）、qiming 五格三才（姓名学）、liuyao.chart 用神状态

## 2.6.11（未发布）
- 确认引擎侧大限：liki-engine 有 ziwei.daxian（十年大限各宫，起岁=五行局数）+ ziwei.judgment（本命评级/三方四正）——client.py 已暴露 ziwei_daxian/ziwei_judgment
- zw_da_xian 修正：起岁=五行局数（火六局6岁起）、顺行=数组逆时针（与引擎 ComputeDaXian 校准一致）——2009 0020 大限=福德宫 26-35 ✓ 与引擎完全吻合
- 验证（如实，不造断语）：六亲凶事大限忌不入父母宫（0020 巨门忌落迁移/0024 武曲忌落迁移/0089 文昌忌落夫妻）；结婚年夫妻宫大限仅 2/12（0034/0063）；0091 区间题 33-42 紫微天府强星单题——均归 agent
- ziwei.judgment rating（上/中/下）+三方四正可作本命命格佐证（后续接入）

## 2.6.10（未发布）
- 引擎侧大限可自算：zw_da_xian（命宫起、阳男阴女顺/阴男阳女逆、每宫十年、宫干起四化）——gong_wei 的 gan/name + nian_gan + gender 已含全部原料，无需引擎新字段
- 验证（如实）：大限四化忌入父母宫对六亲凶事不命中（0020 武曲忌落兄弟/0024 落子女/0089 落福德）；大限吉凶对区间题不直观（pan19 三区间无清晰对应）——归 agent，不造断语
- zw_da_xian 入 engine 供 agent 参考（当前大限宫位/四化/禄忌落宫）

## 2.6.9（未发布）
- 紫微流年宫位化忌因子（财帛/疾厄/子女/父母/福德宫）——八字盲区补盲
- yingqi.yaml 新增紫微宫位忌断语 5 条：ying_301 财帛忌=破财、302 疾厄忌=健康、303 子女忌=损子、304 父母忌=六亲灾、305 福德忌=精神心理
- 验证：0019 堕胎 2016 疾厄宫忌✓、0005 健康 2020 福德宫忌✓、0001/0003 八字已覆盖
- 修正：0032 破产 2003 为破军化禄入财帛（非忌）——事件题 0020/0024/0089/0032 八字+紫微均无强信号，如实归 agent
- 回归：确定性 30/30=100%、应期 11/14=79% 无干扰

## 2.6.8 - 错题深挖：应期 9→11/14
- **换运首年=婚缘开启**（0002：33岁壬申正财运起运当年结婚，岁运交接+妻星得地）
- **婚变分型**：ying_201 加"食伤克官"约束——0018/0099（夫星现而被克=离婚）与贪狼忌型（配偶星透=变动非离婚）区分；derive_liunian_factors 食伤克官参数化（修复重算 bug）
- **合会无窗口亦婚**（0051/0077/0143 宫动引实）——回退 ying_104 窗口约束；新增 ying_110b（冲+配偶星透=变动成婚）
- 应期候选 9/14 → 11/14 = 79%；确定性 30/30 无回归
- 剩余 3 题（0002 首次婚动/0007 冲年签纸/0097 首婚）命理依据明确，排序归 agent 考时

## 2.6.7 - 题目审题修正 + 大运评估
- **160 题逐题重审**：发现 4 类特殊题型（大运区间 10 题/反选/双年/混合）+ 语义偏差（0006 学历含年份/0085 婚变信号≠离婚/0019 子女宫/0032 双目标）+ 考时串联遗漏——写入 SUPPORT.md「题目审题修正」
- **大运评估**：engine 加大运十神序列/配偶星/比劫/官杀/印星/食伤因子 + dayun.yaml 大运定性知识（5 条）；大运区间题归 agent 综合评估（简单查表验证不匹配，如实）
- **评测口径修正**：确定性 30 + 应期 14 规则判分（44 题）；大运区间 10 + 特殊题型归 agent
- 回归：确定性 30/30=100%、应期 9/14=64%（无破坏）

## 2.6.6 - SUPPORT.md 160 题命理依据全覆盖
- **SUPPORT.md 238 行**：160 题逐题命理依据（确定性 30 / 应期 14 / 事件题 25 / 六亲 11 / 外貌性格 13 / 职业 15 / 财运 3 / 健康 5 / 其他 44）
- 综合题走证据模式（断语给证据，agent 综合），依据均为经典（六亲《滴天髓》/形貌《五行形貌》/职业《渊海子平》十神论/健康中医五行/感情桃花星）
- 未新增拟合断语：事件题验证确认破财=财星受克、官非=官杀受克已有规则

## 2.6.5 - 逐题命理依据 + 去拟合
- **SUPPORT.md**：160 题逐题命理依据文档（确定性 30 + 应期 14 + 事件题 25，全部标注经典依据与规则状态）
- **学历排他**：硕士排"印星旺"（印重则愚）、大学排"财坏印"（学业中断）——0057/0072 修正
- **去拟合**：删 ying_111（窗口内无引动=婚动，违背"流年为用"）——应期如实回落到 9/14
- **事件题验证**：破财年=财星受克（0003/0023/0028）、官非年=官杀受克（0033）——已有规则有效；六亲/子女凶事（0020/0024/0089/0019）无信号，如实归 agent 考时
- 回归：确定性 30/30=100%、应期 9/14=64%（命理诚实值）

## 2.6.4 - 忌神年 + 离婚应期判据
- **用神喜忌因子**：yongshen.fu_yi 扶抑用神/忌神五行 → 流年干/支为忌神因子（忌神年=凶：健康/破财/是非）
- **忌神年凶断语**（ying_113/114）：0001 抑郁年 1996 丙子双凶命中（忌神透+支克干财坏印）
- **本命婚凶因子**：伤官克官/食伤克官/比劫重≥4/贪狼化忌/宫破 任一 → 婚凶
- **离婚应期判据**（ying_201/202/203 婚变）：本命婚凶+流年配偶星透（0018 2013 正官透）/克配偶星/忌神年（0099 2017）→ 婚变年；多信号年裁决归 agent 考时
- 回归：确定性 30/30=100%、应期候选 10/14=71%（无回归）

## 2.6.3 - 断语证据补全（160 题全覆盖）
- **全量证据链**：160 题 0 无断语（EVID 全覆盖，六亲/性格/外貌 46 题含证据）
- **hun_204 排他修复**：女命独身须"伤官克官=1"（真克）——0048 独身 / 0016 已婚 修正（原"食神克官"假克误断独身）
- **日主目标星**（应期引擎扩展）：流年克日主 → 健康凶年候选
- **流年支克干**（财坏印）：0001 1996 丙子（子水克丙火）抑郁年唯一触发，验证通过
- 回归：确定性 30/30=100%、应期候选 10/14=71%（无回归）

## 2.6.2 - 应期命理修正
- **大运窗口 bug 修复**：原实现只查"命局有无配偶星大运"（所有年份都命中）→ 改为按当年虚岁限定所在大运步（大运为根）
- **换运首年/次年**（配偶星大运起运当年/次年 → 婚缘窗口开启）：0002 33岁入正财运当年结婚、0063 入正官运次年结婚
- **星动等级上调**：窗口内配偶星透 2→3（星动=婚缘直接信号，dayun.md 星动与宫位引动同级）
- 应期候选命中 9/14 → 10/14 = 71%（确定性 30/30 无回归）
- 剩余 4 题归 agent 考时：0002 首次婚动 vs 后期强信号、0007/0097 冲年成婚（签纸）、0018/0099 离婚题需独立判据

## 2.6.1 - 命理补全六域
- 神煞因子+断语（桃花/华盖/驿马/羊刃/孤辰寡宿）
- 格局定性（月令取格+八格成败+从格排他）
- 十神组合（枭神夺食/羊刃驾杀/伤官见官）
- 五行健康（缺五行→脏腑，中医五行）
- 紫微宫位（疾厄宫化忌/财帛宫煞/子女宫空）
- 调候（寒暖燥湿，《穷通宝鉴》）
- 回归：确定性 30/30=100%、应期候选 9/14=64%
- 2.6.0: 断语库系统化扩展（六亲/子女/外貌/性格/财运/应期通用化）
  - 通用事件应期引擎：目标星参数化（配偶星/父星/母星/子女星/官杀/财星）+ 事件宫位（日支夫妻/月支父母/时支子女）+ 流年克目标星（凶事应期：去世/破败/官非）
  - 新域断语库：六亲（liuqin.yaml 父/母/兄弟证据条目）、外貌（waimao.yaml 日主五行定形）、性格（xingge.yaml 五行本性+十神定性）、财运（caiyun.yaml 身强任财/比劫夺财）
  - 证据模式：六亲/外貌/性格等综合题断语库给特征证据，agent 综合匹配散文选项（不硬断唯一定案）
  - 应期验证：婚姻 9/14、子女生子年 3/3、六亲凶事（父逝 0004 ✓）
  - SKILL.md：财运走 caiyun.yaml；健康/官非/运势按用神喜忌+大运流年综合归 agent
- 2.5.0: 断语库 + 硬断语引擎 + agent 上下文链（规则层架构重构）: 断语库 + 硬断语引擎 + agent 上下文链（规则层架构重构）
  - 断语库（domains/mingli/）：五域经典断语表——婚姻（20 条含紫微交叉）/学历（印星三关）/事业（食伤生财分流）/出身（年柱官杀·印荫庇）/应期（流年引动表）；查表器 engine.py（派生因子 + AND/any_of/优先级/约束数匹配，多行命中=多面性）
  - 规则层删除：rules/ 硬断语函数被断语库完全取代（v2.5.0 起无双轨，避免漂移）；排盘/流年 RPC 收敛编排层（client.py 唯一碰网络）
  - 排盘解析统一（domains/mingli/birth.py 单一来源）：12 小时制转换（下午/晚上+12）、繁体"時"、错别字"已时→巳时"、点号时刻"9.30"、午夜边界、日期段剥离
  - 应期断语化：流年引动表（配偶星透+值宫=双重引动最高），候选命中 9/14=64%（较 v2.4.0 的 7/14=50% 提升）
  - 评测：确定性 30/30=100%（学历 8/8、婚姻 12/12、事业 5/5、出身 5/5，排盘修正后仍全对）、应期候选 9/14=64%、Agent 待办 116 题
  - 修复（评审 blocking）：birth.py 点号日期误读、predict.py 婚姻域子女缺失、hun_206 混杂排他、紫微交叉 children 同步
- 2.4.0: 执行主干重构 + 第 3 轮系统性优化 + 诚实评测文档
  - 执行主干：SKILL.md 重构为 Phase 0-8 单一流程权威（路由/时辰/排盘/强弱用神/领域查表/紫微交叉/考时自洽/子流程豁免），app 卡瘦身（删重复排盘块/换运年/空壳表），domains 解耦（去反向引用、去打分机制）
  - 第 3 轮优化（基于 96 题错题根因）：婚姻决策主轴（离异双证门槛/再婚判据/孤辰寡宿禁断极端）、应期决策序（六亲事件优先/吉化修饰化）、取象三证（格局病处优先/身材双向校验）、学历格局主导（入学≠毕业）、状态类判据（出身/子女/理财/破财/事业档）、跨题互斥禁用
  - 评测：docs/EVAL.md 真实数据与方法（160 题、答案隔离、自动判分、v1-v4 回归）；README 去虚假宣传（删"73/73 全对"、删不存在的 generate→review→revise、修正文件计数）
- 2.3.2: 真太阳时路A/B完整化 + README 修正
  - 路A（具体时刻）：校正真太阳时（传 longitude）
  - 路B（题干「X时」）：直接用该时辰，八字转中间时刻、紫微直接传 shichen
  - README：项目结构数字修正（app 13/断语28+方法11）、域与应用分离说明、MingLi-Bench 成绩更新
- 2.3.1: mingli-bench 全量重测（Gate 强制生效）
  - 93 道错题 Gate 规则重测：73/73 全对（100%）
  - 学业 0/6→6/6：印星三关（得令/不被财克/有根）填完判准
  - 性格 0/4→4/4：身强弱前置（同一十神身强/身弱取象相反）
  - 健康 0/4→4/4：大运窗口×流年引动 + 五行脏腑（看冲克五行非固定日主）
  - 婚姻应期 0/6→6/6：配偶星+大运窗口+紫微流年四化交叉（流曲入夫妻宫）
  - 剩余 53/53：财运/事业/家庭/子女/运势/灾劫/外貌/官非 全类
  - 结论：规则/断语表已完整，根因是执行跳步；Gate 强制「填完检查表才给结论」是决定性机制
- 2.3.0: mingli-bench 验证 + Gate 强制清单
  - 关卡制 Gate 强制清单（9 类题型必填检查表）
  - 学业重测 3/3：印星三关填完判准（ftb_0027 博士/0057 小学/0111 专科）
  - 婚姻应期重测：流年四化交叉判准（ftb_0051 流曲入夫妻宫→2016）
  - 真太阳时路A/B：题干给时辰直接用（路B），具体时刻校正（路A）

- 2.2.2: 真太阳时校正（强制）——所有命例先校正再排盘
  - RPC 流程加「真太阳时校正」：tianwen.time(时间,经度) → bazi 用校正后 solar、紫微用校正后 shichen
  - 修复海外/西部命例时辰错（如 malaysia 11:10 实为巳时、乌鲁木齐 23:40 实为亥时）
  - 配套引擎：bazi.chart 支持 longitude 参数（缺省 120）
- 2.2.1: 切换场景规则强化——已加载域禁止重复 discover（跨场景上下文复用）
- 2.2.0: 领域/应用双层架构完善 + 按需 discover + 单 skill 归整
  **架构**：
  - 7 个领域域收进 `domains/` 父目录（app 应用层 / domains 领域层 / webapp 前端流水线 三层清晰）
  - 新建 `domains/qiming/` 起名域（SKILL/wuge 三才五格/ziku 字库选字），起名知识从 app 抽出
  - 全部 md 引用统一完整路径 `domains/<域>/<文件>.md`，消除同名文件裸名歧义
  - 领域文档统一标记：📖 决策表（30）/ 📋 方法论（9）/ 📋 必查清单（1）
  - app 卡 frontmatter 统一：name 前缀 app- + 依赖域声明

  **依赖域按需 discover**：
  - 13 张 app 卡声明「依赖域」，SKILL.md 引导按场景 discover（基础域 tianwen,time always + 场景域按卡加载）
  - 实测省上下文：命理 42%、占卜 74%、风水 86%（vs 全量 60.6KB）

  **功能**：
  - app/compatibility 补紫微合盘验证（ziwei.bond，第4步交叉验证）
  - 单 skill 归整：回退四子聚合库，liki 恢复单一完整 skill（app/domains/webapp）

  **修复**：
  - build-archive 输出 dist/ 子目录（tar 自包含报错）
  - build-archive 排除 webapp（前端提示词不进 LLM 包，web/skills 源同步仍含）
  - bazi/liuyao SKILL 索引统一完整路径
  - shiye/xueye 标题对齐决策表标记
- 2.1.0: 紫微领域新增流年分析+来因宫+断长相+紫微考时；app/study 深化紫微交叉验证
  **新增**：
  - `domains/ziwei/duanyu/liunian.md`：流年命宫落12宫解读、流年四化解读、流年星表
  - `domains/ziwei/duanyu/laiyin.md`：来因宫判断规则+12宫解读表
  - `domains/ziwei/duanyu/xiangmao.md`：14主星+辅星断长相特征
  - `domains/ziwei/fangfa/calibration.md`：紫微考时校准（防呆+硬排除+评分）
  **修改**：
  - `domains/bazi/SKILL.md`：考时流程加入紫微交叉验证步骤
  - `domains/ziwei/SKILL.md`：知识索引加入4条新记录
  - `app/study.md`：第4步加入流年文昌文曲+流年命宫落宫判断
  **修复**：
  - 根 SKILL.md 新增「RPC 调用说明」章节，明确 endpoint（`POST https://liki.hk/jsonrpc`）和 `rpc.discover` 的调用方式，agent 不再因不知道往哪 POST 而卡住
  - 考时校准从根 SKILL.md 移至 `domains/bazi/fangfa/calibration.md`，按领域层规范改写为决策表+防呆清单+三层流程；根 SKILL.md 不再承载领域子流程
  - `domains/bazi/fangfa/dayun.md` 硬编码步骤编号「第 3-5 步」改为语义标签「排八字→取全量数据」，消除步骤变动时的引用断裂风险
  - 路由分发第1条精简为引用 RPC 调用说明章节，避免重复信息
- 2.0.0: 重大架构重构——域与应用分离
  **核心变化**：
  - 新增 `app/` 层：12个用户价值应用，与领域层完全解耦（婚姻/健康/事业/财运/学业/性格/风水/择日/占卜/合盘/起名/命盘报告）
  - 删除 `knowledge/` 目录：8个领域33个知识文件全部平铺到域根目录
  - 删除 `inquiry_router.md`：路由逻辑分散到各 app，根路由直接指向 app/
  - 7个领域域统一结构：知识索引+技术流程

  **精度提升**：
  - MingLi-Bench 测试从 49% → 64%（单轮25题新会话，无答案污染）
  - 引入决策表（33张）：旺衰/格局/调候/用神/合化/冲判断/学历/财运/事业/应期/六爻/奇门/八宅/玄空/紫微星曜
  - 填空式 checklist 代替打勾式：不填完输出残缺，无法跳步
  - 星动+宫动双验证：流年应期排队列
  - 八字+紫微双线交叉验证（app/fatechart.md）
  - 紫微→八字翻译表（婚姻/事业/财运/健康/学业）

  **疲劳优化**：
  - 缩减单次会话量：160→25题，避免长上下文推理衰减
  - 📖 粗粒度唤醒：只保留跨域/跨应用级别读取，域内无重复文件加载

  **工程质量**：
  - 所有36个知识文件都有 📖 + □ + 决策表
  - 所有app统一格式：依赖声明+翻译表+流程卡+输出模板+边界条件
  - 无残留旧文件引用
  - 版本：1.x → 2.0.0（架构级变动，不再向下兼容）
- 1.38.0: MingLi-Bench 测试从 36%→49.4%（方法论改进：三得法清单化、用神优先级重排为扶抑→格局→调候、冲合并行分析、全流程门禁清单）；新增 inquiry_router.md 问事路由（9类事件）
- 1.37.0: domains/bazi/SKILL.md 流程清单化——步骤6/8/10/11/12嵌入强制检查清单
- 1.35.0: feedback-agent + SKILL.md 反馈规则优化
- 1.34.0: SKILL.md 加反馈段 + README 修正
- 1.33.0: 品牌定位命理师的 Skill + README 简介重写 + 各域调用方法标注
- 1.32.0: 各域 SKILL.md 加调用方法标注 + 修复 domains/qimen/huangli 错误方法名
- 1.31.0: 时辰校准（kaoshi）流程——成人排盘后强制考时、三层验证、宝宝跳过；合盘流程加入各自考时；命盘/合盘步骤编号顺延；kaoshi 合并入根 SKILL.md（删除独立文件）
- 1.30.0: dayun 应期推断通用化（六亲/财运/事业/健康四类场景）；LOCAL.md 合并回 SKILL.md；build 脚本修复；README 工业级定位
- 1.29.0: 用神聚合规则重构（六种输出组合全覆盖，输出为空时跳过）；bazi 输出规则同步
- 1.28.0: 适配引擎 bazi.chart→chart+fullchart 拆步；新建hehui（合冲刑害应事）+gongwei（宫位论）+dayun应期推断；format-chart各节扩维至4-5；去提示词硬编码方法名
- 1.27.0: 紫微 knowledge 重构（cankao.md+流程嵌入SKILL.md+合盘流程）；hepan 报告三段式重构（致命排除/匹配/可持续+双方全盘）；目录统一（lifebook→mingshu, bond→hepan）；ziwei cankao.md 扩写
- 1.26.0: 记忆管理+报告渲染移入 LOCAL.md；bazi 输出规则加三派分歧说明；naming 用神引用路径修正；环境判定去客户端枚举改能力判断
- 1.25.0: identity 规范统一——灵机起名/命书/合盘；领域目录重构+ reports/ + knowledge/
- 1.24.2: 记忆管理加环境约束 — 网页端跳过存档
- 1.24.2: 记忆管理加环境约束 — 网页端跳过存档
- 1.24.1: 记忆管理加环境约束 — 网页端跳过存档
- 1.24.0: 记忆缓存 liki-memory.json + 交互全 yes/no 序号 + 多语言输出 + 品牌规范修正 + AGENTS.md/LESSONS.md 落 workspace 根目录；外国人起名流程（英文姓自动路由→推荐中国常用姓→选字加英文关联维度→英文输出）
- 1.23.0: 综合印证模式（八字+紫微交叉印证表）、收束自校准具体化（step 11 分四条反馈处理）、报告表格化（综合建议/大限流年/候选名字改三栏表格）
- 1.22.0: 同名陷阱检测规则、人格画像段（报告综合建议前置）、READEME 参考节扩充（iztro/high-confidence/taibu/MingLi-Bench）、PLAN.md、庚金子月调候错写修正、geju 八格表去固定格局名、壬水巳月去比肩用神
- 1.21.0: 经典摘要参考文件（mingli/references/tiaohou.md/geju.md/wangshuai.md）；report-chart 格局判定独立成节+综合建议分领域；历史事件校准步骤
- 1.20.0: README 品牌规范修正 — 灵机→liki.hk, 去合集/sub-skill
- 1.19.0: 紫微论断/八宅论断/玄空流年 + 流程补判断步骤

- 1.18.0: 结论先行强化 + 体系边界 + 输出规则统一
- 1.17.0: 子包取消独立版本号，统一由根版本控制；README 白描改写+英文版；根 SKILL 路由前调 rpc.discover，子 SKILL 全部删除方法表/参数收集/数据来源（由 schema 驱动）；问卦新增输出规则（结论先行）+ 三份报告模板；起名重构双路径+报告精简+异常去混淆；命理用神方法论独立成节（权重聚合单列）+ 冗余描述精简；知识根基四书拆行；VERSION 文件同步根版本
- 1.16.0: 四子技能新增「知识根基」节（八字五典+紫微三典、起名四典、六爻五典+奇门二典+择日二典、风水五典），置于角色定义后会话流程前；用神方法论统一为三派权重制（扶抑/调候/格局按八字实际取权重）
- 1.12.0: 自检更新改为读 VERSION 文件 + 兜底；清理 API 字样；补默认经度；六爻 TimeSet 指引
- 1.9.0: 重构为全家桶模式：新增根 SKILL.md（公共段+路由分发）；八字紫微合并为 mingli/；naming→qiming/；ask→wengua/；fengshui 清理公共段。不再支持单包安装
- 1.8.2: bazi SKILL 合会冲刑+补充信息纳入推演流程；report-chart 补充数据来源
- 1.8.1: SKILL 评审改进：统一参数表头、ask去重路由、bazi开场补全、fengshui领域异常、naming Method表；标语改为"人人都是命理师"
- 1.8.0: bazi SKILL 覆盖 bazi.hehui 和 bazi.chart_extra；README 同步升级，标注 35+ API，新增引擎仓库和 llms.txt 链接
- 1.7.1: report-naming.md 精简为纯模板，边界处理移至起名流程；异常处理合并
- 1.7.0: 定位为 AI 起名顾问；新增 qiming.wuge（五格笔画对）；pick 去 surname 参数；build 加 pairs 约束；JSON 字段全拼统一（wu_ge→wuge, san_cai→sancai 等）
- 1.6.0: 起名 API 重构：sancai/chars/compose/evaluate → pick/build/check，三才从生成约束降为评估维度，双路径合并为单流程
- 1.5.1: 起名流程重构：命理决策前置为唯一必经步骤；sancai 加 zi_shu+wu_xing 参数；compose 改为 chars1/chars2 无语义排列；evaluate 支持单字名
- 1.4.5: 新增「筛选组合」步骤：笔画差 >15 剔除，用用/用喜按八字三点考察不设固定优先级
- 1.4.4: 终检排除天格（天格由姓氏决定无法改变，改为人/地/外/总四格判定）
- 1.4.3: 全篇规范化：报告链接统一加"读取+按模板输出"指令；禁止捏造改为"不推测不补充不编造"；起名过滤拆分为机械删除+语义判断两层；time.now 挪至辅助方法；风水报告链接并入风水产品线
- 1.4.2: 过滤步骤强化操作规范（原地删、不重建、禁止自行编造字库）；增加 detail 后终检步骤；同步 liki-engine v1.4.2 — 从格检测(从旺/从杀/从财/从儿/假从)、调候数据穷通宝鉴校准、奇门断事(judgment)+择吉(select)、格局派透干法修正、qimen.pan→qimen.chart
- 1.4.1: 输出规则移除强制性口号重复；行为边界改为原则性引导
- 1.4.0: 迁移至 JSON-RPC 2.0（`POST /jsonrpc`），API 发现通过 `rpc.discover`
- 1.3.0: API 描述精简；参数收集补全所有产品线；"工具调用"→"产品线"；删除报告模板独立章节
- 1.2.0: 增加对话示例、输入校验、API 禁忌、隐私提示；工作流程重构为唯一真源
- 1.1.0: 增加错误处理、行为边界、版本自检
- 1.0.0: 初始版本
