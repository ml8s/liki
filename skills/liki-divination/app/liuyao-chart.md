---
name: app-liuyao
description: 六爻问卦 — 问吉凶、成败、阻碍与应期
依赖域: liuyao
---

# 六爻问卦

## 依赖的领域知识

[必读] - liuyao: `domains/liuyao/yongshen.md`（用神）+ `yuejian.md`（旺衰 / 修饰）+ `jixiong.md`（动爻关系）+ `patterns.md`（格局）+ `liushou.md`（六神）+ `yingqi.md`（应期）

## 📖 流程

| 步骤 | 条件 / 目标 | 动作 | 产物 |
|---|---|---|---|
| 1 | 确认问题 | 识别核心事项、目标、期限和视角；不足时只追问一个问题 | 可排盘的问题 |
| 2 | 场景复核 | 已由 `divination_route` 选中 `liuyao` 时才进入本卡 | 六爻适用性确认 |
| 3 | 起卦 | `liuyao_qigua`：auto / coins / yaos；原始硬币不换算 | casting 起卦收据 |
| 4 | 一次解读包 | 推荐用 `liuyao_read`：传 casting + matter/yong_shen；内部排盘并生成 snapshot、topic、timing、conditions、report 骨架 | 卦象、用神、证据、专题与报告上下文 |
| 5 | 解读六因子 | 旺衰、十二长生、月破、旬空、五行墓库、动爻关系、格局、六神 | 倾向、助力 / 阻碍、结构、色彩 |
| 6 | 复核新事实 | `day_clash_facts`、`moving_transformations`、`hidden_lines`、`branch_relation_facts`、`san_he_candidates`、`force_chain`、`yong_shen_candidates` | 日冲、动变、伏神、支间关系、三合候选与作用链 |
| 7 | 应期 | 读取 `domains/liuyao/yingqi.md`，只解释 `timing_candidates` 中与用神 / 事项相关的候选 | 时间窗口 |
| 8 | 条件还原 | 引用传统断语时调用 `liuyao_conditions` | 条件化候选与禁忌 |
| 9 | 报告 | 用 `liuyao_report(action=template, snapshot=...)` 填写结构化报告，再 `action=validate` 校验 | report v1 + 引用核对 |
| 10 | 事实审计 | 调用 `liuyao_audit`；可传文本报告或结构化 report | 盘面硬事实核对结果 |
| 11 | 会话固化 | `liuyao_session(action=create, ...)` 固化原卦、首次结论和报告 | 可追问 session |

工具返回的 `casting` 是起卦事实，snapshot 是解释上下文；卦盘 `lines`、干支与宫位是展示事实。吉凶解读只使用上述六因子，冲突必须并列。

## 工具调用

`liuyao_qigua` 支持 `mode=auto/coins/yaos`；coins 传六组三枚“正/反”，顺序为初爻到上爻。排盘优先回传上一步完整 `casting`，并传 `question`。

普通问事传 `matter`；高级用户可改传 `yong_shen`，两者互斥。感情必须先确认 `perspective=male/female/unspecified`；`unspecified` 不能推断用神，应改问双方角色或使用显式 `yong_shen`。不要把性别作为家庭代占的替代关系。

## 边界条件

| 异常场景 | 处理方式 |
|---|---|
| 用户说“随便算算” | 先确认最想判断的一件具体事项；确认后才起卦 |
| 问题包含多个独立目标 | 请用户选择本次主问题 |
| 追问同一卦 | 沿用原 `casting`、原 `solar_time` 和首次判断；不用当前时间重排 |
| 现实发生实质变化 | 说明这是新占问，征得确认后再起新卦 |
| 医疗 / 法律 / 重大财务 / 人身安全 | 不排盘、不给吉凶应期，引导专业或紧急帮助 |
| 时间跨度很大 | 说明六爻偏中长期，重大节点可分段起卦 |

## 📖 输出模板

| 输出 | 内容 |
|---|---|
| 结论 | 一句话倾向 / 成败条件 |
| 用神 | 用神六亲、爻位、世应、旺衰修饰 |
| 动爻 | 本卦 → 变卦、生克冲合、助力 / 阻碍 |
| 结构 | 格局与六神色彩 |
| 应期 | 相关机制、触发支、条件与时间窗口；不写确定日期 |
| 条件还原 | 引用传统断语时说明成立条件；不用绝对断语 |
| 建议 | 一个可执行动作或核查条件 |
| 边界 | 传统文化视角；高风险问题引导专业帮助 |
