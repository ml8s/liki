# liki-engine — 开源命理计算引擎

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-AGPL--3.0-green)](./LICENSE)
[![CI](https://github.com/ml8s/liki-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/ml8s/liki-engine/actions/workflows/ci.yml)

**liki-engine** 是一个独立的开源命理计算引擎：以 JSON-RPC 提供 **32 个 API**（31 个 RPC + `rpc.discover`），覆盖**八字、紫微斗数、六爻、奇门遁甲、起名、黄历、八宅、玄空** 8 个领域的天文历算与命理数值计算。**不依赖任何上层 Skill**——你可以直接把它集成到自己的 Web 应用、移动端、桌面工具或任意 AI 框架中。

**它能帮你解决什么**：不用自己实现真太阳时校正、节气计算（VSOP87 秒级精度）、时区夏令时、农历闰月、紫微排盘等极易出错的底层算法——一个 `POST /jsonrpc` 就能拿到精确命盘数据。

## 为什么可以用它

**① 天文精度，非近似** — 节气时刻用寿星历 VSOP87D（移植 lunar ShouXingUtil）：章动/光行差/ΔT 全考虑，精度 ±15min → **秒级**（夏至 2026 与 lunar 差 2 秒）；真太阳时校正、夏令时、经纬度时区、闰月全部按天文算法。

**② 数据驱动测试保障** — 每个核心算法都有 golden 数据锚定：
- **115 个八字 golden 命例**：四柱/大运/节气边界/时辰边界/时区/起运精度
- **紫微 150+100 条**：complete_test + flow_golden（命宫/四化/大限流年）
- **流年 golden 7 条**（手算锚点，非自证）+ 流年/流月/流日神煞单元测试（年支+日支双查、值年神煞全分支）
- 14 个测试包全绿，`make test` 可复现

**③ 独立部署，零依赖** — 单二进制（Go 编译），或 Docker 多阶段构建；启动即用，无外部服务依赖。

## 快速开始

```bash
# 构建
go build -o liki ./cmd/liki/

# 启动（默认 :8080）
./liki -addr :8080

# 调用
curl -s http://localhost:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"bazi.chart","params":{"solar_time":"2000-06-15T12:00:00+08:00","gender":"male"}}'
```

所有响应同一格式：`{"jsonrpc":"2.0","result":{"_product":"<method>","data":{...}}}`。

## API 总览（34 个）

全部通过 `POST /jsonrpc`，标准 JSON-RPC 2.0，仅支持命名参数。每个方法详参可通过 `rpc.discover` 自省。

| 领域 | 方法数 | 方法 |
|------|--------|------|
| 八字 | 8 | chart fullchart bond liunian liuyue liuri liushi xiaoyun |
| 紫微斗数 | 8 | chart daxian fullchart liunian liuyue liuri liushi bond |
| 起名 | 4 | char pick build check |
| 八宅风水 | 2 | chart layout |
| 玄空飞星 | 2 | chart liunian |
| 六爻纳甲 | 2 | qigua chart |
| 奇门遁甲 | 1 | chart |
| 黄历 | 1 | days |
| 工具 | 4 | city time.now tianwen.time rpc.discover |

### 自省（rpc.discover）

`rpc.discover` 返回 OpenRPC 1.4.1 文档（含所有方法的 `params`/`result` schema），支持**子域按需加载**：

```bash
# 全量（所有方法 schema）
curl -s http://localhost:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{}}'

# 子域加载（methods 逗号分隔；支持精确方法名 + 领域前缀）
curl -s http://localhost:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{"methods":"tianwen.time,bazi.chart,ziwei.liunian"}}'

curl -s http://localhost:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{"methods":"bazi,ziwei"}}'
```

- **无 `methods`** → 返回全部方法 schema（体积较大 ~30KB）
- **`methods` 逗号分隔** → 按需过滤，支持两种写法：
  - 精确方法名：`bazi.chart`、`ziwei.liunian`
  - **领域前缀（子域）**：`bazi` → 该域全部方法；`ziwei` → 紫微域全部
- **用途**：AI agent / 客户端按场景只加载所需方法 schema，减少上下文与请求体积（如排盘场景只需 `tianwen.time,bazi,ziwei`）

## 端点

| 端点 | 说明 |
|------|------|
| `POST /jsonrpc` | JSON-RPC 2.0 主入口（唯一） |
| `GET /healthz` | 健康检查 |

## 领域方法详情

### 八字（10）

| 方法 | 功能 |
|------|------|
| `bazi.chart` | 排盘：四柱+纳音+大运+性别（最小集）。如需完整十神/藏干/神煞/长生/空亡，传入 `bazi.fullchart` |
| `bazi.fullchart` | 扩展命盘：补全十神、藏干、神煞、长生、空亡、自合、魁罡 |

| `bazi.bond` | 双人合盘：日主、天干关系、地支关系、纳音、五行互补 |
| `bazi.liunian` | 流年运势：干支/十神/神煞/伏吟反吟。流年神煞=动态 9 种（年支+日支双查）+ 值年 4 种（病符/丧门/吊客/大耗，`shensha[].name` 在 schema enum 声明） |
| `bazi.liuyue` | 流月运势 |
| `bazi.liuri` | 流日运势 |
| `bazi.liushi` | 流时运势 |
| `bazi.xiaoyun` | 小运 |


### 紫微斗数（8）

| 方法 | 功能 |
|------|------|
| `ziwei.chart` | 排盘：十二宫星曜分布、亮度、四化 |
| `ziwei.daxian` | 大限：十年大限各宫吉凶 |
| `ziwei.fullchart` | 全盘：长生、博士、小限、将前、岁前、杂曜 |
| `ziwei.liunian` | 流年：命宫+四化+辅星 |
| `ziwei.liuyue` | 流月：命宫+四化+月星 |
| `ziwei.liuri` | 流日：命宫+四化+日星 |
| `ziwei.liushi` | 流时：命宫+四化+时星 |
| `ziwei.bond` | 合盘：命宫互入+夫妻宫+吉煞星+四化+五行生克 |

### 起名（4）

| 方法 | 功能 |
|------|------|
| `qiming.char` | 单字分析：五行/笔画/吉凶 |
| `qiming.pick` | 选字：按姓氏/五行/笔画 |
| `qiming.build` | 组名：候选字组合+三才五格 |
| `qiming.check` | 校验：给定名字的三才五格吉凶 |

### 黄历（1）

| 方法 | 功能 |
|------|------|
| `huangli.days` | 每日宜忌/吉神凶煞 |

### 八宅风水（2）

| 方法 | 功能 |
|------|------|
| `bazhai.chart` | 八宅盘：命卦 + 四吉四凶方 + 流年紫白飞星 |
| `bazhai.layout` | 门主灶配合：chart + 门/主/灶方位 → 各 match（东四西四同组=吉） |

### 玄空飞星（2）

| 方法 | 功能 |
|------|------|
| `xuankong.chart` | 坐向盘：元运/山向星/旺山旺向/收山出煞 |
| `xuankong.liunian` | 流年飞星：chart（可选）+ year → 流年飞星盘 + 宅盘叠加凶星提示 |

### 六爻纳甲（2）

| 方法 | 功能 |
|------|------|
| `liuyao.qigua` | 起卦（随机，可不依赖时间） |
| `liuyao.chart` | 装卦：六亲/六神/世应/每爻旺衰·月破·动爻生克 |

### 奇门遁甲（1）

| 方法 | 功能 |
|------|------|
| `qimen.chart` | 排盘：九宫/八门/九星/八神 + 日时干落宫/生克/空亡马星影响 |

### 工具（4）

| 方法 | 功能 |
|------|------|
| `tianwen.time` | 真太阳时校正、农历转换 |
| `city` | 城市经纬度查询 |
| `time.now` | 当前时间 |
| `rpc.discover` | OpenRPC 自省 |

## 性能

| 操作 | 耗时 | 分配 |
|------|------|------|
| `bazi.chart` | ~43µs | 3.8KB, 52 allocs |
| `bazi.chart`（含用神三派） | ~10µs | 1.7KB |
| `ziwei.chart` | ~107µs | 5.4KB, 62 allocs |
| `liuyao.chart` | ~10µs | 0.1KB, 2 allocs |
| `qimen.chart` | ~26µs | 1.5KB, 36 allocs |

`go test -bench=. -benchmem` 测量。

## 测试与精度

```bash
make test        # 14 个测试包全绿
go test ./... -count=1
```

- **115 个八字 golden 命例**：四柱/大运/节气边界/时辰边界/时区/起运精度（testdata/bazi_golden*.json）
- **紫微 250 条**：complete_test（150）+ flow_golden（100）+ 文昌/文曲亮度（iztro 生成）
- **流年 golden 7 条**：手算锚点（非引擎自证），覆盖年支/日支动态神煞 + 值年神煞全组合
- **经典参考测试**：《滴天髓》《子平真诠》《穷通宝鉴》案例
- **schema 一致性**：Go 输出字段 vs OpenRPC schema 全量比对 + 神煞 enum 保障

## 部署

```bash
# Docker
docker build -t liki-engine .
docker run -p 8080:8080 liki-engine

# 或直接跑二进制（构建产物零外部依赖）
./liki -addr :8080
```

## 与 Liki Skill 的关系

[liki](https://github.com/ml8s/liki) 是一个基于本引擎的上层 AI Skill（流程定义 + 方法论文档 + 评测体系）。**本引擎独立可用**——Skill 只是众多可能的上层应用之一；你也可以用它构建自己的应用。

## 项目结构

```
├── cmd/liki/       ← 入口（HTTP JSON-RPC 服务）
├── internal/engine/ ← 各领域计算核心
│   ├── bazi/       ← 八字（排盘/用神/格局/流年/神煞）
│   ├── ziwei/      ← 紫微（排盘/大限/流年/合盘）
│   ├── qiming/     ← 起名
│   ├── huangli/    ← 黄历
│   ├── liuyao/     ← 六爻
│   ├── qimen/      ← 奇门
│   ├── bazhai/     ← 八宅
│   ├── xuankong/   ← 玄空
│   ├── tianwen/    ← 天文（真太阳时/节气/农历）
│   └── ganzhi/     ← 干支基础
├── internal/agent/ ← RPC 注册 + schema（OpenRPC）
└── internal/http/  ← HTTP 层
```

## License

AGPL-3.0。命理结论为传统文化视角，仅供参考，不构成专业建议。
