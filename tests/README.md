# Liki 评测（MingLi-Bench，160 题）

Liki 的独立评测体系：160 道命理师大赛真题按命盘分组为 **32 个 case**（每命盘 4-6 题，模拟真实用户"同一命盘连问多题"），用 [skill-up](https://github.com/alibaba/skill-up) + qwen_code 自动评测。

## 文件清单（现行结构）

| 文件 | 说明 |
|------|------|
| `evals/eval.yaml` | skill-up 评测配置（qwen_code + Docker 沙箱；`LIKI_RPC_URL` 注入本地引擎） |
| `evals/cases/pan01..32.yaml` | 32 个命盘分组 case（题目，不含答案） |
| `grade-case.py` | skill-up script judge（自包含：内嵌盘例与答案，提取"题N 答案：X"逐题比对） |
| `answers.json` / `groups.json` / `cats.json` | 标准答案与映射（判分用；评测前自发 stash 隔离） |
| `run-qwen.sh` | 一键运行：起本地引擎→移答案→评测→恢复→判分 |

## 运行

前置：模型 key（`OPENAI_API_KEY`，OpenAI-compatible；DeepSeek 或智谱均可）已配置。智谱可设置 `ZHIPU_API_KEY` 或 `ZHIPUAI_API_KEY`；默认使用 `glm-4-flash-250414`，可用 `OPENAI_MODEL` 覆盖。
本地密钥可放在 `tests/evals/.zhipu.local.env`，该文件已加入 `.gitignore`。

```bash
bash tests/run-qwen.sh --parallelism 16    # 一键评测 + 判分
```

评测特点：
- **本地引擎**：`run-qwen.sh` 自动起本仓 `engine/` 的引擎（`LIKI_RPC_URL` 注入容器），脱离生产 liki.hk，可重复复现
- **答案隔离**：评测前把答案文件 stash 移出 skill 目录，agent 容器物理读不到
- **判分**：skill-up script judge（`grade-case.py`，由各 case 的 `judge.script_path` 引用）随评测完成

## 评测正确性

见根 README「实现机制」「开发者」两节（MingLi-Bench 交叉验证、数据驱动命理锚定）。
