# MCP 清理重构 — 遗留变更队列（新会话接续文档）

> 基线：commit ef08c8d（2026.10.22.1），三套件全绿，工作树干净。
> 新会话接续步骤：① `cd engine && go test ./... && go vet ./...` 确认基线绿；
> ② `scripts/test-counsel.sh`；③ 根目录 `python3 -m pytest tests/`；
> ④ 读本文档"队列"节，按设计执行；⑤ 每层跑回归，绝不带病推进。

## 队列 2（已关闭，按内聚收口）

原计划的 liuyao `board.lines` / qimen `conflicts` 全量强类型迁移不做。保证链已经
内聚：engine Go registry 对输出 DTO 做 schema 校验（`LiuyaoLineOutput` 本身强类型），
counsel 侧 `validate_document` 是唯一 Python 校验出口。`board.lines` 按"engine
强类型 + counsel opaque 事实"处理，与 `factors` opaque + digest 的既有约定同类。

实际加固（已完成）：

- 新增 `DivinationConflict`（`id`/`reason`/`facts`，engine liuyao Conflict 同形），
  统一约束 liuyao snapshot `evidence.conflicts`、`liuyao_query` 与 `qimen_query` 输出。
- `liuyao_factors` 对 engine chart `conflicts` 从静默 `isinstance` 过滤改为按契约
  fail-closed 校验，坏形状显式报错，不再静默丢弃。
- 澄清：qimen snapshot 本身没有 `conflicts` 字段，原队列描述有误；弱类型实际位于
  两个 query 输出，已随本项收口。

## 已关闭项（勿重复处理）

- 残留门禁修复：`check_mcp_schema` 支持顶层 `anyOf` closed-object 输入分支

  （`naming_check` 显式五行 / selected_yongshen 两分支是既有发布契约）；

  natal schema 预算 24 KiB → 64 KiB（2026.10.22.1 契约强化使该组从 8 KiB 增至

  ~51 KiB，旧预算在基线上已失效）；删除 city 死代码 `countryZh`。

- 队列 1：Bazi 候选空串已迁移为 nullable。engine 域层用 nil；wire 用

  `*ElementNameNullable` / `*string`；counsel `YongWuxing` 为五行闭集 `| None`。

  覆盖 `fu_yi.yong/xi/ji`、`tiao_hou.secondary_wuxing`、柱位 `self_he_name`，
  并同步 selected yong 缺省语义与 fixtures。

- qimen bench：已论证 bench 测的 `ComputeChart` 包装 = `ResolveMethod` +

  `ComputeChartWithMethod`，即 wire 计算核心；注释已固化于 bench_test.go

- deadcode 接口方法假阳性：豁免红线已写入本文档上级

  （`MCP_CONTRACT_DESIGN.md`）与清理提示词

- `ParseWangShuai` 保留于 prod：跨包（liuyao oracle 测试）使用的导出解析 API
