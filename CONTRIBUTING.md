# 贡献指南

感谢你考虑为liki.hk贡献！

## 提交 Issue

- 命理计算结果不准确 → 请附上出生时间、地点、期望结果和实际结果
- SKILL 行为异常 → 请说明触发了什么流程、期望什么、实际发生了什么
- 建议新技能 → 请描述用户场景，越具体越好

## 提交 PR

1. Fork 本仓库
2. 创建一个功能分支：`git checkout -b feat/my-change`
3. 安装 git hooks（一次）：`make hooks`
4. 升版本用根 Makefile 统一写入当日日期和序号（`skills/liki/VERSION.txt` + engine VERSION 同步；各领域 `skill-tools.json` 会同步 `info.version`）：

   ```bash
   make version
   ```

5. 同步更新 `CHANGELOG.md`（README 统计数字有变时一并更新）
6. README / 用户指南改动需保持现有标题结构和双语一致性，并运行 `make check`
7. 提交 PR，描述清楚改了什么、为什么

## 代码规范

### 本地工具链

- Go **1.26.6**（低于 1.26.6 会命中标准库安全漏洞）
- Python **3.12**
- Node **22**
- 可选：[uv](https://docs.astral.sh/uv/)，用于复现 counsel Python 依赖锁

### 语言和服务边界

- 根 `skills/liki/SKILL.md` 只做产品路由；每个领域用 `ENTRY.md` 进入。SKILL.md 以中文为主，术语保持原文。
- `engine-mcp` 只做确定性排盘/历法/风水计算；`counsel-mcp` 只做因子、断语、起名和问卦判断。
- 公开 `/engine` 与 `/counsel` 前缀由网关负责；两个服务内部只实现 `/mcp` 与 `/mcp/{domain}`。
- 不恢复旧 Python CLI 或 JSON-RPC 发现/调用路径；历史评测资产已归档，迁移前不得进入 gate。

### Lint 与安全扫描

- 引擎 lint 用 golangci-lint v2（配置 `engine/.golangci.yml`）。本地安装用官方二进制脚本，**不要 `go install`**（golangci-lint 与 Go 版本强耦合，官方明确不推荐该方式）：

  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

- Python lint 使用 Ruff 0.16.0：`make lint-python`。
- Go 漏洞扫描使用 `govulncheck`；counsel 哈希锁使用 `pip-audit`。两者在 CI 分别执行。
- 每次升版本必须同步更新：`VERSION.txt`（make 统一 bump）+ `CHANGELOG.md`；需要生成分发包时运行 `make build-archive`

## 设计原则（为什么这样设计）

- **MCP 调用边界分层**：`engine-mcp` 只做确定性排盘/历法/风水计算，`counsel-mcp` 只做因子、断语、起名和问卦判断。网关负责 `/engine` 与 `/counsel` 公开前缀；服务内只允许 `/mcp` 与 `/mcp/{domain}`。领域文档描述能力与变量绑定，不手写传输层细节。防回潮门禁：MCP smoke、schema 对照与根文档契约测试。
- **历史事件只验证整体框架**：校准/结论验证回退的对象是「格局+用神+大运」的综合解读框架，不是单一用神选择——事件是框架的综合结果，无法反推单一变量（v1.23.0 教训）。落地处：`app/mingshu.md` 历史事件校准节、`domains/bazi/calibration.md`。
- **三派用神必须聚合出唯一结论**：扶抑/调候/格局三派按决策表聚合（`domains/bazi/yongshen.md`），不并列列出让用户选——并列等于把专业判断推给用户（v1.16.0 教训，已落地 yongshen.md 聚合决策表）。

## 发布版本模型

- 日常开发 / 兼容性版本使用 CalVer，由 `make version` 统一 bump。
- 正式产品发行版使用 SemVer Git tag（如 `v5.0.0`），与 CalVer 并行。
- 不要把 `VERSION` 改成 SemVer；CalVer 是运行时和兼容性契约。
- 完整规则见 [docs/RELEASE_MODEL.md](./docs/RELEASE_MODEL.md)。

## 推送前检查清单

**改方法名/函数名时**（全量搜索所有引用点）：

```bash
# 搜代码（含脚本）
grep -rn "旧方法名" --include="*.go" --include="*.sh" --include="*.py"
# 搜 skill 文档
grep -rn "旧方法名" skills/liki/*/app/*.md skills/liki/*/ENTRY.md skills/liki/SKILL.md
```

**添加新 MCP 工具时**（同步更新测试与 manifest）：

```bash
# engine：补 handler、InputSchema/OutputSchema、领域路由与 tool schema 测试
grep -R "bazi_chart" engine/cmd/engine-mcp engine/internal/agent

# counsel：补 manifest schema、业务实现、MCP runtime schema 对照与错误映射测试
grep -R "compute_factors" counsel/app/natal/tools counsel/tests
```

**改 skill 文档后**（需要出包时）：

```bash
make build-archive
```

**推送前本地 CI**：

```bash
# 全量（engine + skills 单测 + 全链路集成，本地自动起引擎）+ schema/文档一致性
make gate
```
