---
name: liki-divination
description: "问卦占卜/算一卦测事 — 六爻起卦、奇门决策、黄历择日。占卜吉凶成败、应期方向、择吉日。Divination: Liuyao / Qimen / date selection. 命理结论为传统文化视角，仅供参考，不构成专业建议。"
---

# Liki 问卦 — 六爻 / 奇门 / 黄历择日

覆盖六爻吉凶与应期、奇门方向与时机决策、黄历择日。引擎与工具表负责确定性计算，LLM 只解释结构化因子。

## 启动与工具

1. 外部安装副本先读本地 `VERSION` 与远程 `VERSION`；不一致时提示更新命令并等待确认，远程 10 秒不可达时标注后继续。托管环境跳过检查。
2. 读 `tools/skill-tools.json` 取工具 schema。
3. 奇门用 `python3 tools/agent_cli.py`：stdin 传 `{"fn":"...","args":{...}}`，stdout 读 JSON；Windows 使用 `tools/agent_cli.cmd` 和 UTF-8 文件。底层 RPC 端点仍由 `LIKI_RPC_URL` 控制。
4. 六爻与黄历直接 POST JSON-RPC：默认 `https://liki.hk/jsonrpc`；用 `rpc.discover` 读取最终 schema。
5. 完成上述检查后进入路由。

## 路由

| 用户问题 | 入口 |
|---|---|
| 六爻 / 问吉凶 / 成败 / 何时有结果 | `app/liuyao-chart.md` |
| 奇门 / 方向 / 时机决策 | `app/qimen-chart.md` |
| 择日 / 黄历 / 吉日 | `app/auspicious.md` |

## 核心流程

| 步骤 | 条件 | 动作 | 产物 |
|---|---|---|---|
| 1 | 所有问卦 | `time.now` | 当前时间基准 |
| 2 | 六爻 | `liuyao.qigua` → `liuyao.chart` | 卦象与六因子 |
| 2 | 奇门 | `city_coords` → `solar_time` → `qimen_chart`；工具内部先投影稳定快照，失物等专占再 `query` | 真太阳时盘面、用神与解释候选 |
| 2 | 黄历 | `huangli.days` | 候选日与宜忌 |
| 3 | 已有数据 | 读取 app 卡与对应 domain 文档 | 解读规则 |
| 4 | 输出 | 按模板综合引擎因子 | 结论 + 依据链 |

## 硬边界

- 起卦、排盘、择日与应期候选来自工具或 RPC；LLM 只解释返回字段。
- 奇门事象路由在 Python 表完成；engine 只接收显式 `yong_shen`，不承接失物等专占结论。
- 奇门 `solar_time` 使用 `tianwen.time` 返回值；已有真太阳时则直接传入。
- 六爻吉凶只解读 app 卡定义的六个因子；卦盘中间字段仅用于展示。
- 医疗、法律、金融问题只给传统文化视角提示，并引导专业服务。

## 输出契约

- 先给一句话判断，再列用神 / 盘面 / 动爻或方法与关键因子。
- 专断因子优先；冲突因子并列解释，三个以上同向因子才形成综合判断。
- 应期只解释引擎返回且与所问对象相关的候选。
- 输出语言跟随用户；英文首次出现核心术语时括注英文。

工具或 JSON-RPC 参数错误按 schema 修正后重试；网络超时告知用户可重试；HTTP 403 更换 HTTP 客户端或请求头。反馈提交到 `https://liki.hk/api/feedback`，请求体使用 UTF-8。

## 交互与安全

- 参数不完整时给默认建议和编号选项；关键排盘结果先确认再深入。
- 仅服务问卦 / 择日话题；明显焦虑时引导专业帮助，避免宿命化表述。
