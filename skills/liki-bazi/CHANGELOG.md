# Changelog

## 2026.09.01.1

- 移除全局因子快照 LRU；求值复用改为显式 `FactorContext` / `NatalContext`。
- 新增 `pan_schema.py`，统一校验 `full_paipan`、流年、合盘入口的完整 pan。
- 新增 `domain_snapshot.py` 与契约文件，保留并保护 reserved 领域事实。
- 因子表迁移为 `factor_id / group / term` 长表，并引入 `印星透根`、`财星透根`。
- 断语表迁移为 `assertions.csv` + `assertion_conditions.csv` 长表。
- 拆分本命/流年算子与年度求值模块，新增统一错误契约。
- 修复考时辅助函数的场景别名引用，并将无效事件规则改为 `AssertionRuleError`。

## 2026.08.31.0

- 移除旧 `extract.py` 中间层，改为 `pan → factors → snap` 直读路径。
- `yearly_range` 增加完整 pan 校验与 120 年跨度上限。
