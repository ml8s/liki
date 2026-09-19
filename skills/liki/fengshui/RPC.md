# Fengshui RPC 契约

本域无 Python 工具层。所有调用都使用根 `SKILL.md` 声明的 JSON-RPC endpoint：

```http
POST /jsonrpc HTTP/1.1
Content-Type: application/json
```

所有请求都使用 HTTP `POST`、`Content-Type: application/json`。`id` 可使用各报文中给出的固定值。

成功业务响应读取 `result.data`；discover 响应读取 `result.info` 和 `result.methods`。错误响应读取 `error` 并停止。

```json
{
  "jsonrpc": "2.0",
  "id": "example",
  "result": {"_product": "liki", "data": {}}
}
```

```json
{
  "jsonrpc": "2.0",
  "id": "example",
  "error": {"code": -32000, "message": "deterministic engine error"}
}
```

## 1. 固定 discover 报文

```json
{
  "jsonrpc": "2.0",
  "id": "discover-fengshui",
  "method": "rpc.discover",
  "params": {
    "methods": "bazhai,xuankong,time"
  }
}
```

`bazhai` 和 `xuankong` 会返回各自全部 schema。校验：

1. `result.info.version` 按点号整数逐段比较，不低于本地 `VERSION.txt`。
2. `result.methods[].name` 至少包含下方业务契约中的全部方法。
3. 任一缺失即 fail closed。

## 2. time.now

```json
{
  "jsonrpc": "2.0",
  "id": "fengshui-time-now",
  "method": "time.now",
  "params": {}
}
```

## 3. bazhai.chart

```json
{
  "jsonrpc": "2.0",
  "id": "fengshui-bazhai-chart",
  "method": "bazhai.chart",
  "params": {
    "birth_year": 1984,
    "gender": "male"
  }
}
```

## 4. bazhai.layout

`ming_gua` 传 `bazhai.chart.result.data.ming_gua.gua.name`。门 / 主 / 灶卦必须来自用户确认的方位换算，不得由模型猜测。

```json
{
  "jsonrpc": "2.0",
  "id": "fengshui-bazhai-layout",
  "method": "bazhai.layout",
  "params": {
    "ming_gua": "兑",
    "door_gua": "乾",
    "master_gua": "坤",
    "stove_gua": "艮"
  }
}
```

| 参数 | 必填 | 闭集 |
| --- | --- | --- |
| `ming_gua` | 是 | 坎、坤、震、巽、乾、兑、艮、离 |
| `door_gua` | 是 | 同上 |
| `master_gua` | 是 | 同上 |
| `stove_gua` | 是 | 同上 |

## 5. xuankong.chart

`period_date` 是宅运起盘日期：建成、入住或大装修改宅日期。坐山 / 向山必须相对，相差 12 山。

```json
{
  "jsonrpc": "2.0",
  "id": "fengshui-xuankong-chart",
  "method": "xuankong.chart",
  "params": {
    "period_date": "2010-06-01",
    "zuo_shan": 0,
    "xiang_shan": 12
  }
}
```

## 6. xuankong.liunian

流年调用必须原样传回 `xuankong.chart.result.data` 完整宅盘；不得裁剪、重建或改写 `chart_digest`。

```json
{
  "jsonrpc": "2.0",
  "id": "fengshui-xuankong-liunian",
  "method": "xuankong.liunian",
  "params": {
    "year": 2026,
    "chart": {"...": "xuankong.chart.result.data 原样对象"}
  }
}
```
