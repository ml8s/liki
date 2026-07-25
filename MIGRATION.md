# Liki 项目拆分迁移指南

## 概览

将原来的 monorepo 拆分为两个独立项目：

| 项目 | 路径 | 职责 |
|------|------|------|
| **liki** (引擎) | `liki_engine/` | 纯命理计算引擎，提供 JSON-RPC 2.0 API |
| **liki_web** (网站) | `liki_web/` | 网站后端 + 前端，通过 JSON-RPC 调用引擎 |

## 拆分后目录结构

### liki (引擎) — 已清理完成

```
liki_engine/
├── cmd/liki/main.go           # 精简入口：仅 HTTP + JSON-RPC
├── internal/
│   ├── engine/                # 全部命理引擎模块
│   │   ├── bazi/ ziwei/ qimen/ qiming/ liuyao/
│   │   ├── huangli/ bazhai/ xuankong/
│   │   ├── tianwen/ ganzhi/ fengshui/
│   ├── agent/                 # RPC 方法注册 + handlers
│   │   ├── rpc_registry.go    # 29 个 RPC 方法定义
│   │   ├── tools_bazi.go      # bazi RPC handlers
│   │   ├── tools_ziwei.go     # ziwei RPC handlers
│   │   ├── tools_other.go     # qimen/bazhai/xuankong/liuyao/huangli
│   │   ├── tools_qiming.go    # 起名 RPC handlers
│   │   ├── types.go           # TimePoint, Person, resolvePerson 等
│   │   ├── compute_time.go    # time.now handler
│   │   └── city/              # 城市经纬度查询
│   └── http/
│       ├── rpc.go             # JSON-RPC HTTP handler (导出 HandleRPC)
│       ├── middleware.go       # CORS, Security, Recover, BodyLimit
│       └── ratelimit.go       # 频率限制
├── go.mod                     # 仅依赖 golang.org/x/time
└── Makefile
```

### liki_web (网站) — 需要新建

```
liki_web/                      # 网站后端 + 前端
├── web/                       # 前端静态文件 (已存在)
├── wiki/                      # wiki (已存在)
├── cmd/liki/main.go           # 从原项目复制
├── internal/
│   ├── agent/                 # LLM Chat Agent
│   │   ├── agent.go           # ChatAgent, LLMClient 接口
│   │   ├── chat_agent.go      # NamingChat 流程
│   │   ├── tools.go           # [重构] ChatToolRegistry → HTTP JSON-RPC
│   │   ├── mock.go            # MockLLM, MockToolRegistry
│   │   ├── data.go + data/    # tools.json, naming.txt
│   │   └── city/              # 城市查询 (或转而通过引擎调用)
│   ├── llm/                   # DeepSeek LLM 客户端
│   ├── payment/               # 支付系统 (SQLite, Dodo, 讯虎)
│   ├── product/               # 产品定义
│   ├── email/                 # 邮件发送 (Resend)
│   ├── dodo/                  # Dodo Payments
│   ├── xunhu/                 # 讯虎短信
│   ├── i18n/                  # 国际化
│   └── http/                  # Web HTTP handlers
│       ├── agent.go           # 命名聊天 SSE handler
│       ├── auth.go            # JWT 登录
│       ├── payment.go         # 订单/支付
│       ├── analytics.go       # 统计
│       ├── location.go        # 位置
│       ├── sse.go             # SSE 事件
│       └── server.go          # 路由注册
├── go.mod                     # module liki_web
└── deploy/                    # Docker/Caddy 配置
```

## Step-by-Step 迁移步骤

### Step 1: 在 liki_web 创建 Go 后端

```bash
cd ../liki_web

# 初始化 module
go mod init liki_web

# 复制 web API 代码（从原始 liki 项目）
# 使用以下命令复制每个包，并替换 import 路径

# 复制 llm 包
mkdir -p internal/llm
cp -r /path/to/original/liki/internal/llm/* internal/llm/

# 复制 product 包
mkdir -p internal/product
cp -r /path/to/original/liki/internal/product/* internal/product/

# 复制 email 包
mkdir -p internal/email
cp -r /path/to/original/liki/internal/email/* internal/email/

# 复制 xunhu 包
mkdir -p internal/xunhu
cp -r /path/to/original/liki/internal/xunhu/* internal/xunhu/

# 复制 i18n 包
mkdir -p internal/i18n
cp -r /path/to/original/liki/internal/i18n/* internal/i18n/

# 复制 dodo 包 (替换 import)
mkdir -p internal/dodo
cp -r /path/to/original/liki/internal/dodo/* internal/dodo/
find internal/dodo -name '*.go' -exec sed -i 's|"liki/internal/|"liki_web/internal/|g' {} +

# 复制 payment 包
mkdir -p internal/payment
cp -r /path/to/original/liki/internal/payment/* internal/payment/
find internal/payment -name '*.go' -exec sed -i 's|"liki/internal/|"liki_web/internal/|g' {} +

# 复制 agent 包 (LLM 部分)
mkdir -p internal/agent/data
for f in agent.go chat_agent.go mock.go data.go doc.go naming_chat_test.go tools_test.go; do
  cp /path/to/original/liki/internal/agent/$f internal/agent/
done
cp /path/to/original/liki/internal/agent/data/tools.json internal/agent/data/
cp /path/to/original/liki/internal/agent/data/naming.txt internal/agent/data/
find internal/agent -name '*.go' -exec sed -i 's|"liki/internal/|"liki_web/internal/|g' {} +

# 复制 http 包 (除引擎保留的文件)
mkdir -p internal/http/testdata
for f in /path/to/original/liki/internal/http/*.go; do
  base=$(basename $f)
  case $base in
    rpc.go|rpc_test.go|middleware.go|middleware_test.go|ratelimit.go|ratelimit_test.go)
      continue ;;
  esac
  cp $f internal/http/
done
cp -r /path/to/original/liki/internal/http/testdata/* internal/http/testdata/ 2>/dev/null || true
find internal/http -name '*.go' -exec sed -i 's|"liki/internal/|"liki_web/internal/|g' {} +

# 复制 cmd 入口
mkdir -p cmd/liki
cp /path/to/original/liki/cmd/liki/main.go cmd/liki/
sed -i 's|"liki/internal/|"liki_web/internal/|g' cmd/liki/main.go

# 添加依赖
go get github.com/dodopayments/dodopayments-go@v1.98.0
go get github.com/go-ozzo/ozzo-validation/v4@v4.3.0
go get github.com/google/uuid@v1.6.0
go get github.com/resend/resend-go/v3@v3.7.0
go get github.com/standard-webhooks/standard-webhooks/libraries@v0.0.1
go get golang.org/x/time@v0.15.0
go get modernc.org/sqlite@v1.52.0
go get github.com/golang-jwt/jwt/v5@v5.3.1

go mod tidy
```

### Step 2: 重构 ChatToolRegistry

将 `internal/agent/tools.go` 中的 `ChatToolRegistry` 改为通过 HTTP JSON-RPC 调用引擎。

**关键变化：**
1. `NewNamingToolRegistry(engineURL string)` — 接受引擎 URL 参数
2. 移除 Go 函数 handler 注册，改为 RPC 方法映射
3. `Execute()` 方法发送 HTTP POST 到引擎的 `/jsonrpc`
4. `Schemas()` 保持不变

参考代码见 `scripts/liki_web_tools.go.ref`。

**映射关系：**
| LLM 工具名 | RPC 方法 |
|-----------|---------|
| `query_city` | `city` |
| `compute_time` | `time.now` |
| `compute_chart` | `bazi.chart` |
| `compute_ziwei` | `ziwei.chart` |
| `compute_naming_wuge` | `qiming.wuge` |
| `compute_naming_compose` | `qiming.compose` |
| `compute_naming_detail` | `qiming.detail` |
| `compute_naming_evaluate` | `qiming.evaluate` |

**ChatAgent 调用处需修改：**
```go
// 旧：直接传引擎 handler
namingTools := agent.NewNamingToolRegistry()

// 新：传引擎 HTTP 地址
namingTools := agent.NewNamingToolRegistry("http://localhost:8080/jsonrpc")
chatAgent := agent.NewChatAgent(llmClient, namingTools, agent.NamingPrompt)
```

### Step 3: 配置引擎地址

在 `.env` 中添加引擎地址：
```
LIKI_ENGINE_URL=http://localhost:8080/jsonrpc
```

### Step 4: 部署

引擎和网站可以独立部署：
- **引擎**：纯计算服务，无状态，可水平扩展
- **网站**：有状态（SQLite），处理用户请求/支付/LLM 对话

生产环境建议使用 Docker Compose 同时启动两个服务。

## JSON-RPC 协议

引擎 API 遵循 JSON-RPC 2.0 协议。所有方法返回 `{"_product":"...","data":...}` 格式。

### 请求格式
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "bazi.chart",
  "params": {
    "birth": {"time": "1984-02-04T18:30:00+08:00", "longitude": 116.4},
    "gender": "male"
  }
}
```

### 响应格式
```json
{
  "jsonrpc": "2.0",
  "result": {
    "_product": "chart",
    "data": { ... }
  },
  "id": 1
}
```

### 所有方法

通过 `rpc.discover` 获取完整 OpenRPC 文档：
```bash
curl -X POST http://engine:8080/jsonrpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{}}'
```
