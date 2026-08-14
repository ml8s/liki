# Liki 评测（MingLi-Bench，160 题）

Liki 的独立评测体系：160 道命理师大赛真题按命盘分组为 **32 个 case**（每命盘 4-6 题，模拟真实用户"同一命盘连问多题"），用 [skill-up](https://github.com/alibaba/skill-up) + qwen-code + 轻量模型自动评测。

## 文件清单

| 文件 | 说明 |
|------|------|
| `cases-grouped/pan01..32.yaml` | 32 个命盘分组 case（题目，不含答案） |
| `answers.json` / `groups.json` / `cats.json` | 标准答案与映射（判分用） |
| `eval-grouped-qwen.yaml` | skill-up 评测配置（engine: qwen_code，Docker 沙箱） |
| `grade-grouped.py` | 自动判分脚本（提取"题N 答案：X"逐题比对） |
| `run-qwen.sh` | 一键运行：移答案→评测→恢复→判分 |

## 运行

前置：`skill-up` 已安装、模型 key 已配置（qwen_code 走 OPENAI_API_KEY/OPENAI_BASE_URL，兼容 deepseek 等）。

```bash
cd skills/liki
skill-up validate tests/eval-grouped-qwen.yaml   # 校验 32 个 case
bash tests/run-qwen.sh --parallelism 16          # 一键评测 + 判分
```

## 隔离原理（防作弊）

- 答案文件（`answers.json`/`groups.json`/`cats.json`）与题目分离，case 文件内不含答案
- `run-qwen.sh` 运行前把答案移出 skill 目录 → 评测容器挂载不到 → 恢复 → 判分
- 即使答案公开在仓库，**评测运行时 agent 也物理读不到答案**，判分独立无人工干预

## 诚实声明

正确率随每轮优化迭代，**不以单一数字固化在文档**（过时即误导）——最新评测数据见发布帖与 `tests/evals/` 判分明细（每轮全量评测自动存档）。评测的价值在于：流程纪律可观测、每轮改动的回归数据可复现。命理结论为传统文化视角，仅供参考。
