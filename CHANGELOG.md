# Changelog

## 2026.09.02.1 —— 起名 API 领域收敛

- **[架构] 移除三才五格与 81 数理规则**：起名链路改为「八字用神 → 五行候选池 → 受控组名 → 字库/五行/音韵评估」，保留 Kangxi 笔画与形体作为汉字事实。
- **[API] `qiming.build` 替换为 `qiming.compose`**：`first`/`second` 只传字，服务端校验字库并生成 given name；最终字符事实由 `qiming.check` 返回。
- **[API] `qiming.check` 输入收敛为 `given_names`**：评估候选名不再要求 surname 参数；姓氏仅由场景层用于最终展示和谐音判断。
- **[API] `qiming.pick` 返回候选字池**：不再接收姓氏或数理参数；单/双名由 `count` 控制。
- **[数据] 修正五行 fallback 生成顺序**：生成运行时字库时先解析部首五行表；961 个可由部首推断五行的字进入候选池，371 个无五行依据的字不进入运行时候选池。
- **[数据] 强化字库校验**：非法笔画、声调、拼音、重复字、负面字表格式错误 now fail fast；`NULL` 占位不再进入 API。
- **[数据] 运行时字库瘦身为领域投影**：新增 `naming_characters.csv` 与 `kangxi_character_strokes.csv`，只保留 qiming 当前消费字段；完整 GSC / Unihan 源表保留为非 embed 数据，姓氏五格覆写表已删除。
- **[数据] 运行时投影只包含可用命名候选**：Kangxi 运行时表同步过滤到 7,734 个具备五行依据的字。

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
- `check_schema.py` / `check_docs.py` 改为校验长表契约与 751 条断语引用

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

# Changelog

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
- `check_schema.py` / `check_docs.py` 改为校验长表契约与 751 条断语引用

### 稳定性

- 删除全局快照 LRU，避免 pan 引用滞留与内容指纹成本
- `yearly_range` 保持 120 年跨度上限；`time.now` 失败不降级本地时钟
- CLI 错误路径返回结构化错误，进程不崩溃；空 pan 明确提示完整盘契约

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
- 比对 domains/bazi/duanyu/*.md vs 对应 yaml：提取 md 独有领域知识，12 条补进 yaml（caiyun 5：破财细分/理财判据/财印/大运方向；jiankang 3：五行双向；xingge 2：身强弱取象；xueye 2：身强印忌）——经典依据
- **断语规则统一**（无"参考"弱化）：yaml（表驱动执行）+ md（规则文字细则/依据）——同一套规则；方法层（fangfa：定用神/考时）保留
- 继续补断语细则：xingge 十神组合取象 4 条（伤官配印/七杀有制/枭神夺食/劫财夺财）+ zhiye 官禄宫昌曲文职 1 条——累计 17 条 md 细则入 yaml（经典依据）
- hehui 合绊/合化吉凶（应期细则：半合需大运/流年引动）——复杂应期层，标注 agent 按 md 综合（SKILL.md 已声明 md 为规则细则）
- 回归：30/30=100% + 11/14=79% 保持；校验 0 错误

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
