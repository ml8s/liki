# Golden fixtures

`tests/golden/<domain>/` 是「开发期生成器 + checked-in 期望值」两件套，只做**漂移检测**，
不是独立真值。独立口径在 `tests/fixtures/domain_oracle/`；三层分工见
`tests/README.md` 的 "Oracle & golden policy"。

## 契约

- 生成器只写 `engine/internal/engine/<domain>/testdata/*.json`；`liuyao` 另读
  `engine/internal/engine/liuyao/data/hexagrams.json` 作为输入。
- 每个生成器的默认输出都从仓库根解析（`Path(__file__).resolve().parents[3]`），
  因此**从任意 cwd 运行都落到同一位置**；`bazi` 可用 `--output`、`tianwen` 可用
  `OUTPUT` 环境变量覆盖。
- checked-in fixture 只由 `make golden` / `make test-engine` 读取。**不要手改 fixture**：
  改生成器后重新生成。
- 生成器是 dev-time 工具，依赖外部参考实现（`lunar-python`、`sxtwl`、Node 版
  bazi-calculator），**不进任何门禁**；`make golden` 只跑消费侧。
- `counsel_naming_baseline.json` 由 `counsel/tests/test_counsel_naming.py` 消费（命名域
  的唯一正确性回归锁）。

## 重新生成

```bash
python3 tests/golden/<domain>/generate.py
```

各域的依赖、口径与覆盖范围见该目录下的 `README.md`；`bazhai` / `huangli` / `liuyao` /
`xuankong` 不需要额外安装（只用标准库或 checked-in 输入）。
