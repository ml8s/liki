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
6. README / 用户指南改动需遵循 [README_STYLE.md](./docs/README_STYLE.md)，并运行 `make lint-readme`
7. 提交 PR，描述清楚改了什么、为什么

## 代码规范

- 根 `skills/liki/SKILL.md` 只做产品路由；每个领域用 `ENTRY.md` 进入。SKILL.md 以中文为主，术语保持原文。bazi / divination 的 LLM 只调用对应 `tools/skill-tools.json` 中的 Python 工具；naming / fengshui 当前无 Python 工具层，只能使用 ENTRY 固定 discover 闭集内声明的 RPC
- 引擎 lint 用 golangci-lint v2（配置 `engine/.golangci.yml`）。本地安装用官方二进制脚本，**不要 `go install`**（golangci-lint 与 Go 版本强耦合，官方明确不推荐该方式）：

  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

- 每次升版本必须同步更新：`VERSION.txt`（make 统一 bump）+ `CHANGELOG.md`；需要生成分发包时运行 `make build-archive`

## 设计原则（为什么这样设计）

- **RPC 调用边界分层**：bazi / divination 的领域 `ENTRY.md` / app 卡只引用本领域 Python 工具；RPC 排盘、因子求值和断语匹配由 `agent_cli.py` 白名单编排。naming / fengshui 当前无工具层，只允许固定 discover 闭集内声明的方法。原因：双层工具会让 LLM 混用入口并漏做 Python 层契约校验。防回潮门禁：`tests/check_docs.py` 的方法白名单与根文档契约测试。
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

**添加新 RPC 方法时**（同步更新测试）：

```bash
# 检查方法计数
grep -c "Name:" engine/internal/agent/tools_*.go                        # 实际方法数
grep "expected.*methods" engine/internal/agent/rpc_registry_test.go     # 测试期望值
grep "want.*methods" engine/internal/http/rpc_test.go                   # 测试期望值
# 添加到方法白名单
grep -A 5 "METHOD_WHITELIST" tests/check_docs.py
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
