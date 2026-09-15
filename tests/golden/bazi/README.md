# BaZi multi-oracle golden

本目录只做开发期 fixture 生成，不引入 Python / Node 运行时依赖。CI 和
`make test-engine` 只读取生成后的 checked-in fixture。

## 校验范围

`multi_oracle_golden.json` 覆盖 2024-2027 年每年 24 个节气的前后锚点，
共 192 个四柱 case。锚点取节气时间的前后约 25 小时；如落到 23 点，
再避开 1 小时，防止把“晚子时换日”流派差异混入硬 golden。

边界策略：

| 层 | 结论 |
|---|---|
| 精确节气边界 | 继续由 `bazi_golden*.json` 用 lunar-typescript 锚定。 |
| 多源共识边界 | 本目录用三个独立实现的共识 fixture，覆盖所有 24 节气前后。 |
| 大运 / 三元 | 仍由现有 golden 覆盖；三个参考库没有共同可靠契约。 |
| 旺衰 / 用神 / 解释 | 不做多源投票，走经典文献与领域 oracle。 |

## Oracle 来源

| Oracle | 项目 | 版本 | 用途 |
|---|---|---|---|
| `lunar-python` | [6tail/lunar-python](https://github.com/6tail/lunar-python) | 1.4.8 | 主参考 |
| `sxtwl` | [sxtwl_cpp / PyPI sxtwl](https://pypi.org/project/sxtwl/) | 2.0.7 | 独立历法实现 |
| `bazi-calculator` | [shenshuge/bazi-calculator](https://github.com/shenshuge/bazi-calculator) | 1.0.0 | 独立四柱实现，2000-2099 |

只有三个 oracle 完全一致的字段才写入 `expected.pillars`。任何不一致会
让 generator 显式失败；不能静默挑选多数派。

## 重新生成

开发机准备：

```bash
python3 -m pip install lunar-python==1.4.8 sxtwl==2.0.7
git clone https://github.com/shenshuge/bazi-calculator.git /tmp/bazi-calculator
npm --prefix /tmp/bazi-calculator install
npm --prefix /tmp/bazi-calculator run build
```

生成：

```bash
BAZI_CALCULATOR_DIR=/tmp/bazi-calculator \
  make golden-bazi-generate
```

只跑 checked-in 数据：

```bash
make golden-bazi
```
