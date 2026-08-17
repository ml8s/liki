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
- `dong_sheng`: 动爻生用（0/1）
- `dong_ke`: 动爻克用（0/1）
- `patterns`: 特殊格局数组（类型/子类型/爻位/断语）

## CSV 断语表

- `career.csv`: 事业断语（用官鬼相关因子）
- `wealth.csv`: 财运断语（用妻财相关因子）
- `relationship.csv`: 感情断语（用财官相关因子）
- `academic.csv`: 学业断语（用父母相关因子）
- `travel.csv`: 出行断语（用世爻相关因子）
- `home.csv`: 住宅断语（用父母相关因子）
- `legal.csv`: 法律断语（用官鬼相关因子）
- `family.csv`: 家庭断语（按亲属取用神）
- `health.csv`: 健康断语（用世爻相关因子）
- `weather.csv`: 天时断语（用父母相关因子）
- `general.csv`: 通用断语（按问题取用神）
- `patterns.csv`: 特殊格局断语（10类）
- `yingqi.csv`: 应期断语
