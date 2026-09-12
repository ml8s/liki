# Rule Engine Functional Suite

本目录是规则引擎和工具层的**功能测试 / 行为契约测试**，不是命理正确性 oracle，也不是 golden 样本集。

## 覆盖目标

```text
factors      因子求值与机械算子行为
assertions   断语表加载、条件匹配、命中与禁止行为
scenarios    场景别名展开与默认领域过滤
conflicts    同盘多断语的选择边界
```

## 目录

```text
manifest.json              功能样本索引与最低覆盖契约
factors/cases.json         因子与机械算子行为样本
assertions/cases.json      断语匹配行为样本
scenarios/cases.json       场景与领域过滤样本
conflicts/cases.json       多断语共存 / 互斥行为样本
```

`tests/fixtures/domain_oracle/` 是另一类数据：跨 Python / Go 的领域真值表和锚点数据源。  
本目录不复制这些大表，也不把机械执行结果冒充命理正确性证明。

## 边界

1. 这些测试回答“规则是否按当前契约正确执行”。
2. 这些测试不回答“每条命理断语在命理上是否正确”。
3. `basis` 表示锁定该行为的原因，不表示完备的命理证明。
4. 生产 CSV 全量结构检查继续由 `make check` 负责。
5. 命理正确性如需继续加强，应另建独立 domain oracle，按高风险簇设计，不用 CSV 行数撑规模。

## 运行

```bash
make test-functional
```
