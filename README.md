# liki-engine — 灵机命理引擎

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-AGPL--3.0-green)](./LICENSE)
[![API](https://img.shields.io/badge/API-2.0.0-blue)](https://github.com/ml8s/liki-engine)
[![CI](https://github.com/ml8s/liki-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/ml8s/liki-engine/actions/workflows/ci.yml)

liki-engine 是[灵机命理服务](https://github.com/ml8s/liki-skills)的底层计算引擎，提供 44 个 JSON-RPC 命理 API：八字、紫微斗数、奇门遁甲、六爻、起名、黄历、风水等。

灵机命理服务由以下部分组成：
- **[liki.hk](https://liki.hk)**：在线服务入口
- **liki-engine**（本仓库）：核心计算引擎，提供精确的命理数值计算
- **[liki-skills](https://github.com/ml8s/liki-skills)**：面向 AI agent 的技能库，包含服务流程、调用方法、领域知识

> 对 AI agent：请访问 [liki-skills](https://github.com/ml8s/liki-skills) 获取技能定义和工作流指引。在线服务入口：[liki.hk](https://liki.hk)。

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

## API 总览

全部通过 `POST /jsonrpc`，标准 JSON-RPC 2.0，仅支持命名参数。每个方法详参可通过 `rpc.discover` 自省。

| 领域 | 方法数 | 方法 |
|------|--------|------|
| 八字 | 11 | chart fullchart yongshen hehui chart_extra bond liunian liuyue liuri liushi xiaoyun xiaoxian |
| 紫微斗数 | 7 | chart daxian liunian liuyue liuri judgment bond |
| 起名 | 5 | char pick build wuge check |
| 黄历 | 4 | date month bond.date bond.month |
| 八宅风水 | 3 | minggua chart judgment |
| 玄空飞星 | 3 | sanyuan annual chart |
| 六爻纳甲 | 3 | qigua chart judgment |
| 奇门遁甲 | 3 | chart judgment select |
| 工具 | 3 | time.now tianwen.time city |

### 自省

```bash
curl -s http://localhost:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{}}'
```

返回完整的 OpenRPC 1.4.1 文档，包含所有方法的 `params`/`result` schema。

## 端点

| 路径 | 说明 |
|------|------|
| `POST /jsonrpc` | JSON-RPC 2.0 端点 |
| `GET /health` | 健康检查 |
| `GET /version` | 构建版本 |

## 领域方法详情

### 八字（11）

| 方法 | 功能 |
|------|------|
| `bazi.chart` | 排盘：四柱+纳音+大运+性别（最小集）。如需完整十神/藏干/神煞/长生/空亡，传入 `bazi.fullchart` |
| `bazi.fullchart` | 扩展命盘：传入 `bazi.chart` 的结果，补全十神、藏干、神煞、长生、空亡、自合、魁罡 |
| `bazi.yongshen` | 用神分析：扶抑（旺衰）、调候（穷通宝鉴）、格局（子平）三派综合 |
| `bazi.hehui` | 合会冲刑：天干五合、地支六合/三合局/三会方、六冲、六害、相刑 |
| `bazi.chart_extra` | 补充信息：三元（胎元/命宫/身宫）、拱夹、纳音生克、长生十二宫 |
| `bazi.bond` | 双人合盘：日主、天干关系、地支关系、纳音、五行互补 |
| `bazi.liunian` | 流年运势 |
| `bazi.liuyue` | 流月运势 |
| `bazi.liuri` | 流日运势 |
| `bazi.liushi` | 流时运势 |
| `bazi.xiaoyun` | 小运 |
| `bazi.xiaoxian` | 小限 |

### 紫微斗数（7）

| 方法 | 功能 |
|------|------|
| `ziwei.chart` | 排盘：十二宫星曜分布、亮度、四化 |
| `ziwei.daxian` | 大限：十年大限各宫吉凶 |
| `ziwei.liunian` | 流年命盘及各宫变化 |
| `ziwei.liuyue` | 流月命盘及各宫变化 |
| `ziwei.liuri` | 流日命盘及各宫变化 |
| `ziwei.judgment` | 综合盘论断：格局+四化+三方四正+综合评级(上/中/下) |
| `ziwei.bond` | 合盘 |

### 起名（5）

| 方法 | 功能 |
|------|------|
| `qiming.char` | 查字：查询单个汉字的五行、笔画、部首、拼音 |
| `qiming.pick` | 按五行取字（分笔画组），用神/喜神需分两次调用 |
| `qiming.build` | 组名：传入字库和笔画约束对，全量排列生成候选名 |
| `qiming.wuge` | 查五格可行笔画对：返回人/地/外/总四格全吉的笔画组合 |
| `qiming.check` | 批量评估名字：五格、三才、五行、音韵全量判定 |

### 黄历（4）

| 方法 | 功能 |
|------|------|
| `huangli.date` | 按日查宜忌，含时辰吉凶 |
| `huangli.month` | 按月查宜忌 |
| `huangli.bond.date` | 八字合参择日 |
| `huangli.bond.month` | 八字合参择月 |

### 八宅风水（3）

| 方法 | 功能 |
|------|------|
| `bazhai.minggua` | 命卦查询：东四命/西四命、命卦、四吉四凶方 |
| `bazhai.chart` | 综合命卦与飞星分析 |
| `bazhai.judgment` | 门主灶论断：分析门/主/灶卦位与命卦配合吉凶 |

### 玄空飞星（3）

| 方法 | 功能 |
|------|------|
| `xuankong.sanyuan` | 三元九运查询 |
| `xuankong.annual` | 流年飞星：入中星、九宫飞布、吉凶评级 |
| `xuankong.chart` | 山向飞星盘 |

### 六爻纳甲（3）

| 方法 | 功能 |
|------|------|
| `liuyao.qigua` | 起卦：三枚铜钱随机摇六次 |
| `liuyao.chart` | 装卦分析：纳甲、六亲、六兽、用神、旺衰、应期 |
| `liuyao.judgment` | 断卦：传入 chart + 事件类型，返回用神状态和评级 |

### 奇门遁甲（3）

| 方法 | 功能 |
|------|------|
| `qimen.chart` | 排盘：天盘、人盘、神盘、九星八门格局。`kind` 默认 `shi`（时家），可选 `ri`/`yue`/`nian` |
| `qimen.judgment` | 断事：分析用神状态、吉凶格局、门宫生克，输出评级和断语 |
| `qimen.select` | 择吉选时：在日期范围内遍历时辰，按事件类型评分排序，返回最佳时段 |

### 工具（3）

| 方法 | 功能 |
|------|------|
| `time.now` | 服务端当前时间（UTC、本地、北京时间） |
| `tianwen.time` | 根据公历时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间 |
| `city` | 根据城市名查经纬度（基于 OpenStreetMap Nominatim） |

## 性能

| 操作 | 耗时 | 分配 |
|------|------|------|
| `bazi.chart` | ~43µs | 3.8KB, 52 allocs |
| `bazi.yongshen`（三派） | ~5µs | 0.9KB, 13 allocs |
| `ziwei.chart` | ~107µs | 5.4KB, 62 allocs |
| `liuyao.chart` | ~10µs | 0.1KB, 2 allocs |
| `qimen.chart` | ~26µs | 1.5KB, 36 allocs |
| `qimen.judgment` | ~2µs | 0.2KB, 5 allocs |
| `qimen.select`（1天） | ~494µs | 38KB, 729 allocs |

`go test -bench=. -benchmem` 测量。

## 项目结构

```
├── cmd/liki/           # 主入口
├── internal/
│   ├── engine/         # 领域引擎（bazi/ziwei/qimen/liuyao/huangli/xuankong/bazhai/qiming/ganzhi/tianwen/fengshui）
│   ├── agent/          # JSON-RPC 注册和路由
│   └── http/           # HTTP 服务和中间件
├── scripts/test-rpc.sh # 冒烟测试（94 个断言）
├── .githooks/          # Git hooks
└── .github/workflows/  # CI
```

## License

AGPL-3.0
