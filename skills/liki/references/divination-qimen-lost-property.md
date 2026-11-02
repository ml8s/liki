# 奇门占失物

> 命理依据：《奇门遁甲秘笈大全·卷九·占失物》以日干落宫为失主、时干落宫为失物；《奇门遁甲元灵经·卷十八·占失物》补方向与内外远近。

公共边界见 `references/divination-ENTRY.md`「流程产物与边界」。

## 解释边界

1. 失主与失物读 `day_stem_palace` / `hour_stem_palace`。快照中的十干克应、门星克应、门破 / 门制与值符值使落宫是预留领域因子，当前失物表未消费。
2. 反吟为复得候选，时干落空亡为难复得候选；时干宫生日干宫且乘旺相气为复得候选，旺相状态由 engine 的 `gong_wang_shuai` / `hour_stem_palace.wang_shuai` 输出。日 / 时干的天盘与地盘层见 `day_stem_palace.facts` / `hour_stem_palace.facts`。墓、绝未入表。
3. 时干落宫方向由八宫方向表投影；内宫为近处候选，外宫为远处候选。
4. 无命中时输出「本表未命中」。
