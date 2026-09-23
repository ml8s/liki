# Natal 领域入口（编排层）

Liki 本命域：**八字、紫微判断由对应专家执行**，本域负责编排（跨领域：考时/合参/合盘、综合命书）与路由。

## 路由

| 用户场景 | 交给 | 说明 |
| --- | --- | --- |
| 八字看命/流年/合盘/考时 | **liki-bazi**（八字专家）| 排盘 + 判断 + 合盘 + 考时（engine-pro-bazi）|
| 紫微看命/流年/合盘/考时 | **liki-ziwei**（紫微专家）| 排盘 + 判断 + 合盘 + 考时（engine-pro-ziwei）|
| 考时（时辰存疑）| 八字+紫微专家 | 参考两侧（调 liki-bazi + liki-ziwei）|
| 合参/双盘综合 | 八字+紫微专家 | 两侧断语同向综合（不臆造）|
| 综合命书 | app/mingshu.md | 综合编排（引用专家方法论）|

## 跨领域编排

- **考时（编码）**：调 liki-bazi 的 `calibrate_birth_time` + liki-ziwei 的 `calibrate_birth_time`，两侧吻合度对比；有真实事件时用 3-5 段已发生时段验证。
- **合参**：八字专家出八字侧断语、紫微专家出紫微侧断语；同向时综合（互证，非臆造），冲突时并列呈现两侧证据。
- **合盘（双人）**：双方分别调八字专家/紫微专家 `compare_birth_charts`。

## App 卡（综合编排，保留）

| App 卡 | 场景 |
| --- | --- |
| `app/mingshu.md` / `mingshu-full.md` | 综合命书（跨八字紫微，引用专家方法论）|
| `app/marriage.md` / `wealth.md` / `career.md` / `study.md` / `health.md` / `family.md` / `personality.md` | 各领域综合（引用 `liki-bazi/skills/bazi/*` 与 `liki-ziwei/skills/ziwei/*` 方法论）|
| `app/compatibility.md` | 双人合参 |

## 边界

- 八字/紫微的**实际判断工具**由专家执行（engine-pro-bazi / engine-pro-ziwei），本域不直接调用。
- 断语以专家工具输出为准（规则真值表），方法论卡用于理解依据；不自行编造。
- 命理是传统文化视角的条件性解读，不构成医疗、法律或投资建议。