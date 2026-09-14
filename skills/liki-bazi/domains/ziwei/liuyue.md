# 紫微流月周期契约

## 输入口径

- `ziwei.liuyue` 只接受已解析完成的农历日期：`target_lunar.year / month / day / leap`。
- 不接受公历日期；公历必须先由 tianwen 历法链路转成农历。
- 不接受裸月份：同一农历年的六月与闰六月是两个不同真实月份。
- 不接受 half 或 fixLeap 参数：闰月前后半由完整农历日期决定。

## 周期解析

| 真实农历目标 | `calendar_period.half` | `flow_month` |
|---|---|---|
| 普通月任意一日 | whole | 同一农历月 |
| 闰月 1–15 日 | `first` | 同一闰月 |
| 闰月 16 日至月末 | `second` | 下一农历月 |

`resolved_period.calendar_period` 描述真实时间所在周期；`resolved_period.flow_month` 是排盘实际采用的流月。LLM 必须显式读取 `flow_month`，不得自行用日期改算。

## Fail closed

- 农历月不存在时报错；闰月不存在时不得回退到普通月。
- 农历日超过真实月末时报错；真实月末由 tianwen 计算。
- 缺 `day` 或 `leap` 时报错；“2025 闰六月”未给日期时不能默认取前半。

## Skill 边界

当前 `liki-bazi` 工具层仍只封装 `full_paipan`、`query`、`yearly_range`、`bond` 等高阶工具；`ziwei.liuyue` 是引擎级 API。若后续暴露流月工具，schema 必须复用这里的 `target_lunar` 与 `resolved_period` 契约，Python 不得判断 `day > 15`。
