# 六爻吉凶判定

> 经典依据：《增删卜易》《卜筮正宗》

## 断语查询

调用 Python 工具层查询断语：
```python
from tools.liuyao.duanyu import query, evaluate_factors

factors = evaluate_factors(chart, yong_shen)
results = query('career', factors)
```

## 引擎输出字段

- `wang_shuai`: 用神旺衰（旺/相/休/囚/死）
- `yue_po`: 用神月破（0/1）
- `xun_kong`: 用神旬空（0/1）
- `dong_yao_relations`: 动爻与用神的关系（数组，枚举：生用/克用/比和/冲用/生原神/克原神/生忌神/克忌神）
- `patterns`: 独立格局（六冲/六合/反吟/伏吟/进退/三刑/六神）

## CSV 断语表

- `enum_general.csv`: 综合断语表（用神旺衰×主要动爻关系×格局 → 结论）
- `yingqi.csv`: 应期断语（用神状态×修饰 → 应期）
