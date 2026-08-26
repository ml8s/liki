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
4. 升版本用根 Makefile 统一 bump（skill 4 份 VERSION + engine VERSION 同步；小修 patch / 流程变更 minor / 破坏性 major）：
   ```bash
   make version-patch   # 或 version-minor / version-major
   ```
5. 同步更新 `CHANGELOG.md`（README 统计数字有变时一并更新）
6. 提交 PR，描述清楚改了什么、为什么

## 代码规范

- SKILL.md 以中文为主，术语保持原文；不写方法名和参数（由引擎 schema 驱动）
- 引擎 lint 用 golangci-lint v2（配置 `engine/.golangci.yml`）。本地安装用官方二进制脚本，**不要 `go install`**（golangci-lint 与 Go 版本强耦合，官方明确不推荐该方式）：
  ```bash
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```
- 每次升版本必须同步更新：`VERSION`（make 统一 bump）+ `CHANGELOG.md`；改任何分发内容后重算指纹（`make build-archive`）

## 推送前检查清单

**改方法名/函数名时**（全量搜索所有引用点）：
```bash
# 搜代码（含脚本）
grep -rn "旧方法名" --include="*.go" --include="*.sh" --include="*.py"
# 搜 skill 文档
grep -rn "旧方法名" skills/*/app/*.md skills/*/SKILL.md
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

**改 skill 文档后**（重算指纹）：
```bash
make build-archive
```

**推送前本地 CI**：
```bash
# 全量（engine + skills 单测 + 全链路集成，本地自动起引擎）+ 指纹/文档一致性
make test-all && make pre-push
```
