# Changelog

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
- 1.4.2: 过滤步骤强化操作规范（原地删、不重建、禁止自行编造字库）；增加 detail 后终检步骤
- 1.4.1: 输出规则移除强制性口号重复；行为边界改为原则性引导
- 1.4.0: 迁移至 JSON-RPC 2.0（`POST /jsonrpc`），API 发现通过 `rpc.discover`
- 1.3.0: API 描述精简；参数收集补全所有产品线；"工具调用"→"产品线"；删除报告模板独立章节
- 1.2.0: 增加对话示例、输入校验、API 禁忌、隐私提示；工作流程重构为唯一真源
- 1.1.0: 增加错误处理、行为边界、版本自检
- 1.0.0: 初始版本

- 1.4.2: 同步 liki-engine v1.4.2 — 从格检测(从旺/从杀/从财/从儿/假从)、调候数据穷通宝鉴校准、奇门断事(judgment)+择吉(select)、格局派透干法修正、qimen.pan→qimen.chart
