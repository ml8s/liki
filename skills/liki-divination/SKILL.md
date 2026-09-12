---
name: liki-divination
description: "问卦占卜/算一卦测事 — 六爻起卦、奇门决策、黄历择日。占卜吉凶成败、应期方向、择吉日。Divination: Liuyao / Qimen / date selection. 命理结论为传统文化视角，仅供参考，不构成专业建议。"
---

# Liki 问卦 — 六爻 / 奇门 / 黄历择日

覆盖六爻结果与应期、奇门策略与方向、黄历择日。LLM 负责语义路由和解释结构化因子；Python 负责排盘、snapshot、digest 校验和 answer 契约。

## 启动与工具

1. 外部安装副本先读本地 `VERSION` 与远程 `VERSION`；不一致时提示更新命令并等待确认，远程 10 秒不可达时标注后继续。托管环境跳过检查。
2. 读 `tools/skill-tools.json` 取工具 schema。
3. Python 依赖见 `tools/requirements.txt`；缺失时用 `python3 -m pip install -r tools/requirements.txt` 安装，不得跳过契约校验。
4. 只使用 `python3 tools/agent_cli.py`：stdin 传 `{"fn":"...","args":{...}}`，stdout 读 JSON；Windows 使用 `tools/agent_cli.cmd` 和 UTF-8 文件。底层 RPC 端点由 `LIKI_RPC_URL` 控制，是工具层内部依赖，不是 LLM 的直接调用接口。
5. 启动 CLI 会通过 `rpc.discover` 检查 engine 版本；低于 `2026.09.12.2` 时直接失败并提示升级 engine，不得降级调用旧 RPC。
6. 完成上述检查后进入路由。

## 路由

LLM 先读取 `app/question.md` 并判断用户目标；不要调用独立 route 工具。

| 用户目标 | 工具链 |
|---|---|
| 事件结果 / 能不能成 / 何时有结果 | `liuyao_snapshot` → `liuyao_ask` |
| 策略 / 进退 / 方向 / 行动时机 | `qimen_snapshot` → `qimen_ask` |
| 哪天适合做事 | `huangli_days` |
| 结果和策略混合 | 先澄清，只保留一个主目标 |
| 长期命局 | 不临时起卦，改用八字命理技能 |
| 高风险现实事项 | 不排盘，按安全边界转专业帮助 |

用户明确指定六爻、奇门或黄历时，安全放行后优先用户指定。

## 硬边界

- 起卦、排盘、择日、因子投影和应期候选全部来自工具；LLM 只解释返回字段。
- 六爻原始硬币必须原样传给 `liuyao_snapshot`；LLM/Python 不得自行换算爻值或判断动爻。
- `liuyao_ask` 只接受 `method=liuyao` 的 snapshot；`qimen_ask` 只接受 `method=qimen` 的 snapshot。
- snapshot 是 immutable 上下文；追问复用同一 snapshot。只有新事件才创建新 snapshot。
- 不直接向 LLM 暴露 raw chart 编排工具；LLM 不读取、修改或伪造 `snapshot_digest`。
- 奇门事象路由在 Python 表内完成；engine 只接收显式 `yong_shen`。
- 高风险问题由 `liuyao_snapshot` / `liuyao_ask` / `qimen_snapshot` / `qimen_ask` / `huangli_days` 的共享安全边界拦截。
- 输出只基于 answer 与 snapshot 因子；不得把条件性倾向写成确定结果。

## 输出契约

- 先给一句话判断，再列用神 / 盘面 / 动爻或方法与关键因子。
- 专断因子优先；冲突因子并列解释，三个以上同向因子才形成综合判断。
- 应期只解释 answer 引用且与所问对象相关的候选。
- 涉及健康、生育、年龄窗口、重大财务或时间敏感决策时，附现实专业确认提示；不得把条件性倾向写成必然结果。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。

工具参数错误按 `tools/skill-tools.json` schema 修正后重试；网络超时告知用户可重试；HTTP 403 更换 HTTP 客户端或请求头。反馈提交到 `https://liki.hk/api/feedback`，请求体使用 UTF-8。

## 交互与安全

- 流程表中标记 ⛔ 的步骤为阻塞确认：LLM 必须展示当前结果和编号选项，等待用户回复后才能继续。禁止跳过 ⛔ 节点直接起卦、排盘或展开分析。
- 流程表中标记 💬 的步骤为参数收集：LLM 一次列出所有待收集项和默认值，用户可一次回复或说“都用默认”。
- 每个交互步骤只输出该步骤的内容，禁止提前输出后续步骤的结果。
- 仅服务问卦 / 择日话题；明显焦虑时引导专业帮助，避免宿命化表述。
