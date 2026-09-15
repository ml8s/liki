# Tianwen calendar golden

历法是所有排盘域的地基。这里用 **lunar-python + sxtwl** 两套独立实现做
开发期共识；只有两源完全一致的农历日期和节月才写入 checked-in fixture。
CI 只运行 fixture，不联网、不安装参考库。

覆盖 2024-2027 年每个真实农历月的初一是月末（含 2025 闰六月）。测试锁定：

- `SolarToLunar`
- `LunarToGregorian` 往返
- `JianYue` 节气月支

重新生成需要：

```bash
python3 -m pip install lunar-python==1.4.8 sxtwl==2.0.7
PYTHONPATH="$PYTHONPATH" make golden-tianwen-generate
```
