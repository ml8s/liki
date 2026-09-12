# Liki testing layers

`tests/` 有确定性功能测试与两类独立的 agent 评测。不要用 160 题准确率基准做功能回归，也不要把功能 smoke 当作命理准确率证明。

## Functional rule-engine tests: `functional/`

`functional/` 是规则引擎功能测试，不负责证明全部命理断语正确。它锁定确定性行为：

```text
factor      因子求值与机械算子行为
assertion   断语表加载、条件匹配、命中与禁止行为
scenario    场景别名展开与默认领域过滤
conflict    多断语共存 / 互斥边界
```

数据与执行分离：

| 文件 / 目录 | 说明 |
|---|---|
| `functional/manifest.json` | 功能样本索引与最低覆盖契约 |
| `functional/factors/cases.json` | 因子与机械算子行为样本 |
| `functional/assertions/cases.json` | 断语匹配行为样本 |
| `functional/scenarios/cases.json` | 场景别名与领域过滤样本 |
| `functional/conflicts/cases.json` | 多断语行为样本 |
| `test_functional.py` | runner，只加载、执行、比较 |

`fixtures/domain_oracle/` 继续作为跨 Python / Go 的领域真值表与锚点数据源；本目录不复制这些大表。
修改因子表、断语表或场景映射前先运行：

```bash
make test-functional
```

## Accuracy benchmark: `benchmark/mingli160`

MingLi-Bench 保持独立：160 道命理师大赛真题按命盘分组为 32 个 case，每盘 4-6 题。它用金标准答案判分，回答的问题是：

> 当前八字 / 八紫命理能力有没有回归？

```bash
make benchmark-mingli160
```

关键文件：

| 文件 | 说明 |
|---|---|
| `benchmark/mingli160/eval.yaml` | 160 题 skill-up accuracy benchmark 配置 |
| `benchmark/mingli160/evals/cases/pan01..32.yaml` | 32 个命盘分组 case，不含答案 |
| `benchmark/mingli160/grade-case.py` | 金标准答案 judge |
| `benchmark/mingli160/answers.json` / `groups.json` / `cats.json` | 判分数据，运行时隔离 |
| `benchmark/mingli160/run.sh` | 起本地 engine、隔离答案、运行评测、恢复答案 |

本地模型密钥放在 `benchmark/mingli160/evals/.zhipu.local.env`。

## Behavior smoke: `skillup/`

`skillup/` 是跨领域功能测试。它不判开放解释的优劣，而是判 agent 是否遵守硬契约：

- 是否走正确领域；
- 是否调用正确 RPC；
- 是否引用 engine 事实；
- 是否处理 fallback；
- 缺输入时是否拒绝排盘；
- 是否在候选集外编造事实。

```bash
make skillup-smoke-validate
make skillup-smoke
make skillup-smoke-bazi
make skillup-smoke-divination
make skillup-smoke-fengshui
make skillup-smoke-naming
```

初始覆盖：

| 域 | case 数 | 重点 |
|---|---:|---|
| 八字 / 八紫 | 4 | 排盘、用神、缺输入、双盘合参 |
| 问卦 | 6 | 六爻路由、奇门路由、晚子时口径 |
| 风水 | 5 | 八宅命卦、门主灶、玄空宅盘、流年、缺输入 |
| 起名 | 6 | 受控姓氏、多候选、fallback、无出生时间、自选名校验、非拉丁边界 |

`grade.py` 只做 script judge，检查必要事实、RPC 收据和禁止行为。
本套件不进入 `pre-push`；模型 key 放在 `skillup/evals/.local.env`。
问卦 smoke 的 Docker 镜像必须预装 `jsonschema>=4,<5`；不要依赖 agent 临时联网安装。

## Stable checks

确定性测试仍然走：

```bash
make check
make test-functional
make test
make test-engine
make test-integration
make pre-push
```
