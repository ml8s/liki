# Liki skill-up 功能 smoke

这一组测试和 `tests/benchmark/mingli160/` 完全解耦。

| 套件 | 目标 | 运行 |
|---|---|---|
| `tests/benchmark/mingli160/` | 160 题命理答案准确率基准 | `make benchmark-mingli160` |
| `tests/skillup/` | 领域功能、RPC 调用、路由、fallback 与越权边界 | `make skillup-smoke` |

功能 smoke 不判断开放式解释是否“更好”，只判断硬行为：

1. 是否调用正确 RPC；
2. 是否引用 engine 返回的关键事实；
3. 是否走对六爻 / 奇门 / 八宅 / 玄空等领域入口；
4. 是否正确表达 fallback；
5. 是否在缺少输入时拒绝排盘；
6. 是否没有在 engine 候选之外创造事实。

每个 case 最后必须输出机器可读收据：

```text
SMOKE_ID: <case-id>
RPC_CALLS: <rpc1,rpc2|none>
EVIDENCE: <关键事实>
BOUNDARY: <领域边界说明>
```

`grade.py` 读取 skill-up 注入的 final output / transcript，并按内嵌规则判分。  
运行前需要本地模型 key；本套件不进入 `make pre-push`。

### 运行前置

- 本地 engine：runner 会通过 `scripts/local-engine.sh` 自动启动。
- 模型 key：放入 `tests/skillup/evals/.local.env`，或导出 `OPENAI_API_KEY` / `ZHIPU_API_KEY`。
- Docker 镜像：`liki-qwen:1` 必须包含 `qwen_code` 运行时；问卦 smoke 还要求 Python 依赖 `jsonschema>=4,<5`。未预装时请构建派生镜像，不要让 agent 在评测中临时联网安装依赖。

```bash
make skillup-smoke-validate        # 只校验 case / config
make skillup-smoke                 # 跑全部功能 smoke
make skillup-smoke-bazi            # 只跑八字
make skillup-smoke-divination      # 只跑问卦
make skillup-smoke-fengshui        # 只跑风水
make skillup-smoke-naming          # 只跑起名
```

配置校验不执行模型，也不启动真实 agent：

```bash
make skillup-smoke-validate
```
