# app — 场景卡索引

三张卡共享“输入 → 排盘 / 五行策略 → 取字 → 组名 → 校验 → 输出”的领域链；差异在卡内流程表中说明。全局工具契约与硬边界见根 `SKILL.md`。

| 卡 | 功能 | 依赖域 |
|---|---|---|
| `naming.md` | 通用起名 / 改名 | bazi,qiming |
| `foreign.md` | 罗马字姓转受控中文姓候选并起中文名 | bazi,qiming |
| `selfcheck.md` | 自选名字评估 | qiming |
