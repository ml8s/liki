# TODO

## 子平格局相神系统（待实施）

**状态：已评审，未实施。** 当前工作区已有大量未提交改动，本文件只记录后续方案，不改变现有代码行为。

### 结论

采用“子平格局相神系统”，但按分层方式接入：

1. 定格神；
2. 定相神候选与角色；
3. 判成格、破格、救应；
4. 再与从格、调候、扶抑合参，输出唯一 `selected_yongshen`。

不直接把现有扶抑、调候表整体替换，也不把现有 `ge_ju` 候选直接当成格局派用神。

### 现状

#### engine

`engine/internal/engine/bazi/bazi_yongshen.go` 的 `ComputeYongShenSchools` 并列返回三派：

| 派别 | 当前输出 | 是否给出五行用神 |
| --- | --- | --- |
| 扶抑 | `fu_yi.yong / xi / ji` | 是 |
| 调候 | `tiao_hou.primary_wuxing / secondary_wuxing` | 调候候选 |
| 格局 | `ge_ju / yong_fa / pattern_god / structure` | 否 |

`computeGeJu` 当前流程：

- 先判建禄、月刃：临官为建禄，阳干帝旺为月刃，均输出逆用。
- 否则按月令藏干本气、中气、余气依次找透干，第一个透干者定格。
- 都不透时取月令本气，标记 `main_qi_not_transparent`，作为虚格。
- 通过 `geju_rules.json` 取得格名和顺逆用；比肩、劫财等记为杂格。
- `buildGeJuStructure` 只输出格神、克格神、生格神的透干、藏根、季节旺衰与关系事实。

当前没有：

- 相神对象与相神角色；
- 成格、破格、救应判断；
- 格局派五行用神、喜神、忌神；
- 格局裁决的 `rule_id`、证据链和置信度。

代码定位：

- `engine/internal/engine/bazi/bazi_yongshen.go`（第 31 行）：`GeJuResult` 明确不输出 `yong/xi/ji`。
- `engine/internal/engine/bazi/bazi_geju.go`（第 5 行）：当前只做格局候选，不做全盘复核。
- `engine/internal/mcpcontract/bazi_wire.go`（第 527 行）：schema 明确标注“格局派：月令格神 + 顺逆用 + 结构（非用神）”。

#### counsel

最终用神由 `counsel/app/natal/tools/yongshen/policy.v1.json` 裁决，但当前没有格局成格分支。

当前优先级：

- 从格：10–12；
- 调候冲突：20；
- 调候：21；
- 扶抑：32；
- 中和/未决：39–50。

问题：

- 格局不参与最终用神选择。
- `ge_ju_clear` 只判断 `pattern_god_source == main_qi`，漏掉中气、余气透干、本气不透、建禄、月刃。
- `ge_ju_clear` 仅用于 `YS-FUYI` 的置信度，不改变用神结论。
- `assertions.csv` 中的 `ge_*` 只是粗粒度断语，没有相神对象、成破条件链和救应链，不能直接充当格局裁决引擎。

### 目标架构

全部命理判断表驱动，Python/Go 只计算事实和查表，不在代码里散落格局规则。

1. **定格层**
   - 月令本气、中气、余气；
   - 透干来源；
   - 建禄、月刃；
   - 特殊格、杂格。

2. **相神层**
   - 格神十神；
   - 相神候选；
   - 相神角色：生格、护格、制破格、通关；
   - 成立条件、优先级、证据。

3. **成败救应层**
   - `成格 / 破格 / 救应 / 待复核`；
   - `rule_id`；
   - 证据；
   - 置信度。

4. **合参层**
   - 从格；
   - 格局成格或救应；
   - 调候极端月与冲突守卫；
   - 扶抑回退；
   - 中和、未决。

推荐初始优先级，最终以经典命例验证后固定：

1. 从格/假从；
2. 格局成格或救应，且相神有效、无冲突；
3. 调候极端月或冲突守卫；
4. 扶抑；
5. 中和/未决。

### 接口建议

优先扩展现有 `bazi_yongshen`，不新增重复用神方法。

建议在 `ge_ju_context` 中增加格局裁决对象：

```json
{
  "ge_ju": "正官格",
  "ge_shen": "正官",
  "pattern_status": "成格",
  "xiang_shen": [
    {
      "ten_god": "正印",
      "role": "护格",
      "rule_id": "GJ-ZG-001",
      "evidence": []
    }
  ],
  "confidence": 0.82
}
```

并在 `selected_yongshen` 中保留来源和规则：

```json
{
  "source": "ge_ju",
  "rule_id": "YS-GJ-001"
}
```

约束：

- `bazi_yongshen` 是唯一用神权威。
- `compute_factors.selected_yongshen` 只镜像同一份裁决，不重新计算。
- `naming` 继续消费同一份用神，不自行判格局。
- 暂不新增 `bazi_geju` MCP 方法，除非出现独立消费方。
- 相神是十神角色，不等于五行用神；最终五行用神必须经过十神到五行映射和全盘校验。

### 实施步骤

1. 修复 `ge_ju_clear`，覆盖中气、余气透干、本气不透、建禄、月刃。
2. 新增格局相神表与成败救应表，表内包含流派/规则版本、条件、角色、优先级和证据类型。
3. 在 counsel 输出 `ge_ju_decision`，先不改变最终用神。
4. 影子模式运行，比较新格局裁决与现有 `selected_yongshen` 的差异。
5. 增加经典命例 golden cases，覆盖正官、七杀、财、印、食神、伤官、建禄、月刃、从格。
6. 对齐后打开格局优先分支。
7. 更新 `skills/liki/references/natal-bazi-geju.md`、`natal-bazi-yongshen.md` 和四个专家身份。
8. 重跑完整测试并重新生成 expert-packs、WorkBuddy 包。

### 验收标准

- 每个格局结论都能追到 `rule_id` 和结构证据。
- `ge_ju` 候选、`ge_*` 断语、格局裁决、最终用神四者不互相矛盾。
- 格局成格但用神完全无关的样本必须能逐条解释。
- 现有样本的结果变化必须可审计，不允许静默漂移。
- 相神、成破、救应不由 Python 代码硬编码。
- `needs_review` 必须给出明确原因，不强行选定。

---

## 现有用神计算方法补 golden（待实施）

**状态：已识别，未实施。**

### 问题

当前测试对用神计算的覆盖不足，无法证明现有方法在命理上正确：

- `tests/golden/bazi/README.md` 明确说明：四柱 golden 只做历法/四柱共识，旺衰、用神、解释不做多源投票，走经典文献与领域 oracle。
- `tests/fixtures/domain_oracle/bazi_yongshen_structure.json` 只覆盖少量结构事实：4 个格局 case、2 个扶抑 case、1 个调候 case，且主要验证结构字段，不验证最终 `selected_yongshen`。
- `engine/internal/engine/bazi/bazi_yongshen_reference_test.go` 多数经典案例只 `t.Logf` 或检查字段非空，没有断言具体用神、喜神、忌神。
- `counsel/tests/test_yongshen_policy.py` 主要验证 policy 求值机制，不等于验证命理结论。
- 缺少从盘面到最终 `selected / xi / ji / source / status / rule_id / conflicts` 的端到端基准。
- 缺少对扶抑、调候、格局、从格、冲突、`needs_review` 的系统性覆盖。

因此，现有实现存在“测试通过但用神结论错误”的风险，不能拿当前实现自己生成期望值来自证。

### 目标

为现有用神计算方法补一套独立的 golden/oracle 基准，覆盖三派及其合参结果。

每个 case 至少包含：

- 出生信息或四柱输入；
- 期望的 `selected.yong`；
- 期望的 `selected.xi`；
- 期望的 `selected.ji`；
- 期望的 `source`；
- 期望的 `rule_id`；
- 期望的 `status`；
- 期望的 `confidence`；
- 期望的 `conflicts`；
- 命理依据、出处或人工裁决说明。

### 覆盖范围

1. **扶抑**
   - 身强；
   - 身弱；
   - 中和；
   - 通根、透干、得令、得地、得势的边界；
   - 三会、三合、半合对强弱的影响。

2. **调候**
   - 极端寒暖月令；
   - 调候主神透干、藏支、缺失；
   - 调候辅神可用、不可用；
   - 调候与扶抑冲突；
   - `needs_review` 场景。

3. **格局**
   - 正官、七杀、正财、偏财、正印、偏印、食神、伤官；
   - 建禄、月刃；
   - 杂格；
   - 本气透干、中气透干、余气透干、本气不透；
   - 格神受冲克、生扶、护卫；
   - 后续相神系统的成格、破格、救应 case。

4. **从格**
   - 真从；
   - 假从；
   - 从格未证；
   - 有印比藏根破格；
   - 从格与调候、扶抑的优先级。

5. **合参与状态**
   - 三派一致；
   - 两派一致、一派冲突；
   - 冲突后 `needs_review`；
   - 无可用规则时的 `unresolved`；
   - 每个结论的 `rule_id` 可追溯。

### 实施方式

- 基准放在 `tests/fixtures/domain_oracle/`，作为独立 oracle，不放在 `tests/golden/bazi/` 的四柱漂移检测里。
- 增加消费侧测试，建议放在 `counsel/tests/`（`test_yongshen_golden.py`），直接调用 `resolve_yongshen` / `published_yongshen`。
- 期望值来自经典文献、命理规则表和人工裁决，不能由当前实现自动生成。
- 区分两类测试：
  - **oracle**：独立真值，验证结论是否正确；
  - **golden**：当前已确认结论的漂移检测，防止无意变更。
- 若使用经典命例，必须记录出处、原文依据和必要的流派说明。
- 同一 case 同时校验 engine 三派事实与 counsel 最终裁决，避免两层各自正确但合参错误。
- 增加一个汇总报告，列出每个 case 的命理依据和当前结果，便于人工复核。

### 验收标准

- 三派都有独立正例、反例和边界 case。
- 每个 case 的 `selected_yong / xi / ji` 都有明确期望值，不只是字段非空。
- `source`、`rule_id`、`status`、`confidence`、`conflicts` 全部可断言。
- 冲突和 `needs_review` 不被“为了跑通”强行改成 resolved。
- 新增 case 能在不依赖当前实现输出的情况下人工复核。
- 现有样本若结果变化，必须由 golden diff 明确暴露并给出解释。
