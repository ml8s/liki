# MCP 清理重构 — 遗留变更队列（新会话接续文档）

> 基线：commit ef08c8d（2026.10.22.1），三套件全绿，工作树干净。
> 新会话接续步骤：① `cd engine && go test ./... && go vet ./...` 确认基线绿；
> ② `scripts/test-counsel.sh`；③ 根目录 `python3 -m pytest tests/`；
> ④ 读本文档"队列"节，按设计执行；⑤ 每层跑回归，绝不带病推进。

## 队列 1（下一个执行）：ElementName 空串 → nullable 迁移

**动机**：调候辅五行无候选时输出 `""`，消费方无法区分"领域事实无候选"与
"链路错误空值穿透"（歧义哨兵）。正确建模：缺失 = null。

### L1 域层（engine/internal/engine/bazi/）

- `TiaoHou.SecondaryWuxing string` → `*string`（nil = 无辅调候）
- `FuYi.Yong/Xi/Ji string` → `*string`（nil = 无候选）
- `Pillar.SelfHeName string` → `*string`（nil = 无自合）
- 构造点：`bazi_tiaohou.go:96` 无辅调候时 nil；`bazi_pillar.go:87 selfHeName()` 返回 nil
- 域内消费者：nil 判断替代 `""` 判断

### L2 wire（engine/internal/mcpcontract/bazi_wire.go）

- 对应字段 → `*ElementName` / `*string`（仓库既有指针惯例）
- JSON roundtrip 自动贯通：域 nil marshal → null → wire nil → null
- 验证点：跑 output schema 检查，确认字段渲染 nullable

### L3 counsel

- `contracts/yongshen.py`：`YongWuxing = Literal["木","火","土","金","水"] | None`

  （None = 无候选；**空串从枚举中删除**）

- `yongshen/evaluator.py:67` 的空值判断改 None 语义
- `factors.py` 消费点 `== ""` → `is None`

### L4 fixtures/baselines

- `engine/internal/mcpcontract/testdata/actual/bazi_chart_*_fullchart.json`：

  `fu_yi.yong/xi/ji`、`*.self_he_name`、tiaohou secondary 的 `""` → `null`

- counsel orthogonal baseline 同步（fu_yi 证据内 `""` → null）

### 回归

`go test ./...`、`scripts/test-counsel.sh`、根 `pytest tests/`、`make check`。

## 队列 2（队列 1 之后）：防御纵深类型化

- liuyao `board.lines`、qimen snapshot `conflicts` 从弱类型移植强类型

  （形状源自 engine liuyao.go Line 结构与 qimen_projection_contract）

- 现状有 `divination_contracts.validate_document` JSON 契约兜底，属加固非纠错

## 已关闭项（勿重复处理）

- qimen bench：已论证 bench 测的 `ComputeChart` 包装 = `ResolveMethod` +

  `ComputeChartWithMethod`，即 wire 计算核心；注释已固化于 bench_test.go

- deadcode 接口方法假阳性：豁免红线已写入本文档上级

  （`MCP_CONTRACT_DESIGN.md`）与清理提示词

- `ParseWangShuai` 保留于 prod：跨包（liuyao oracle 测试）使用的导出解析 API
