# Liki testing layers

`tests/` 有确定性功能测试与两类独立的 agent 评测。不要用 160 题准确率基准做功能回归，也不要把功能 smoke 当作命理准确率证明。

## Unified skill structure: `skills/liki/`

仓库现在只有一个可安装 skill：`skills/liki`。根 `SKILL.md` 只负责产品路由、全局 RPC、安全和 feedback；四个领域入口是：

```text
bazi/ENTRY.md
divination/ENTRY.md
fengshui/ENTRY.md
naming/ENTRY.md
```

结构契约由 `tests/test_unified_skill_structure.py` 和 `tests/check.sh` 锁定：

- 全 skill 只有一个 `SKILL.md`；
- 四个领域都有 `ENTRY.md`；
- `VERSION`、`feedback.py`、`feedback.schema.json` 不重复；
- 根入口保持轻量；
- 旧 `liki-*` skill 名不出现在安装内容中；
- `bazi/TOOLS.md` 与 `divination/TOOLS.md` 覆盖全部 Python 工具；
- `naming/RPC.md` 与 `fengshui/RPC.md` 覆盖全部固定 RPC。

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

## External golden calibration: `golden/bazi/`

八字四柱有独立多源共识 golden。`golden/bazi/` 的 generator 只在开发机运行，
它用 `6tail/lunar-python`、`sxtwl` 和 `shenshuge/bazi-calculator` 三方完全
一致的四柱生成 checked-in fixture；CI 只运行 fixture，不访问网络，也不安装
这三个 oracle。

当前覆盖 2024-2027 年全部 24 节气的前后锚点，共 192 例。精确分钟边界、
晚子时和海外时区仍由 engine 内的 `bazi_golden*.json` 锁定；旺衰、用神和断语
不做开源项目多数派投票。

```bash
make golden-bazi            # 只跑 checked-in 数据
make golden-bazi-generate   # 开发机重新生成；依赖见目录 README
```

其他 engine 域使用同一套原则补齐了数据驱动 golden：

| 域 | fixture | 覆盖 |
|---|---|---|
| `tianwen` | lunar-python + sxtwl 共识 | 2024-2027 农历月初 / 月末、闰月与节月支，99 例 |
| `liuyao` | 京房八宫领域 oracle | 64 卦卦名、宫位、世应全量 |
| `huangli` | lunar-python + sxtwl + 建除黄黑道规则 | 90 个日期，覆盖 15 类事项、12 建除 |
| `bazhai` | 通行命卦公式 checked-in oracle | 1900-2099 × 男 / 女，400 例 |
| `xuankong` | 三元九运飞星矩阵 | 九运 × 12 对正向下卦山向，108 张完整盘 |
| `qimen` | 已有 atopx 外部锚点 + 置闰 / 飞盘 / 山向等多套 golden | 方法矩阵与流派边界继续锁定 |
| `ziwei` | 已有 iztro 兼容流日 / 流时 / 流分与亮度 golden | 流盘边界与确定性契约继续锁定 |

一键运行确定性 engine golden：

```bash
make golden-engine
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
