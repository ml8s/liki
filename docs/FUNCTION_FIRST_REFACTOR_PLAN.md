# Function-first time and domain refactor plan

本文记录时间、天文、干支和上层命理域的函数化改造方案。目标是把计划固化下来，
避免后续实施时忘记边界、顺序和验收标准。

## 当前进度（2026-10-22）

已完成：

1. wire 字段从 `utc` 收敛为 `instant`；
2. `instant` 接受 RFC3339 的 `Z` 或 numeric offset；
3. `timezone` 由 `mcpcontract` 统一加载为 `*time.Location`；
4. `mcpcontract` 不再引用 `timekernel`；
5. `timekernel` 已删除；
6. `tianwen` 已函数化真太阳时、农历、时辰、月支、节气状态和 DST 候选；
7. `ganzhi` 已函数化年柱、月柱、日柱、时柱和四柱；
8. `tianwen.ComputeBazi` 已改为组合干支函数；
9. 八字、紫微、六爻、奇门均已提供显式时间入口；
10. handler 均直接解析 `instant / timezone / longitude`；
11. 八字结构投影、旺衰、五行计数、扶抑、调候、格局和周期函数已拆分；
12. 六爻 `validate / base / strength / yongshen / advanced / guaci` pipeline 已拆分；
13. 紫微闰月、宫位、星曜、亮度、四化投影已函数化；
14. naming 的候选字池、组名、评估和姓氏匹配已是函数式 pipeline；
15. registry 收敛到 26 个概念；
16. `make check` 的 skill/packaging 指纹缓存已启用；
17. `make check`、`make test-engine`、`make test-counsel`、`make test-engine-mcp`、`make gate` 绿色。

## 目标

最终形态不再依赖大对象聚合，而是使用标准库时间和一组单一权威函数。

核心规则：

1. 运行时绝对时间直接使用 `time.Time`，wire 输入由边界归一化为 UTC。
2. 时区直接使用 `*time.Location`，wire 字段叫 `timezone`，必须是显式 IANA 时区。
3. `longitude` 是显式参数，只供真太阳时等需要地理位置的能力使用。
4. 不再实现 `TimeSet`、`TimeContext`、`ZonedDateTime`、`CivilDateTime`、`Offset`。
5. 需要什么派生事实，就调用对应函数。
6. 每个派生事实只能有一个权威函数。
7. 上层领域函数显式接收自己需要的最小参数。
8. `ComputeChart` 这类入口只负责编排，不直接实现所有规则。

wire 输入允许 canonical UTC：

```json
{
  "instant": "1990-05-20T04:00:00Z",
  "timezone": "Asia/Shanghai",
  "longitude": 116.407526
}
```

也允许等价 numeric offset：

```json
{
  "instant": "1990-05-20T12:00:00+08:00",
  "timezone": "Asia/Shanghai",
  "longitude": 116.407526
}
```

服务端立即归一化为 UTC `time.Time`。offset 只是序列化细节，不是时区身份。

## 当前原则

### 核心时间模型

长期保留的运行时类型只有标准库类型：

```go
time.Time
*time.Location
```

wire scalar 可以保留：

```go
type Instant string
```

它只表示 wire 字段 `instant` 的字符串形态，不是运行时时间对象。

### instant 解析

允许：

```text
1990-05-20T04:00:00Z
1990-05-20T12:00:00+08:00
1990-05-19T23:00:00-05:00
1990-05-20T04:00:00.123Z
```

拒绝：

```text
1990-05-20T04:00:00
1990-05-20 04:00:00Z
1990-05-20T04:00:00+0800
1990-05-20T04:00:00-00:00
```

解析器应统一为：

```go
func parseInstant(value string) (time.Time, error) {
    if strings.HasSuffix(value, "-00:00") {
        return time.Time{}, fmt.Errorf("instant uses unknown offset -00:00")
    }

    at, err := time.Parse(time.RFC3339Nano, value)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid instant %q: %w", value, err)
    }

    return at.UTC(), nil
}
```

该函数放在 wire 边界包内，所有接收 `instant` 的 handler 使用同一入口。

### timezone 解析

只接受显式 IANA 时区：

```text
UTC
Asia/Shanghai
America/New_York
```

拒绝：

```text
Local
empty
+08:00
UTC+8
CST
```

解析器应统一为：

```go
func loadTimezone(value string) (*time.Location, error) {
    if value == "" {
        return nil, errors.New("timezone is required")
    }

    if strings.EqualFold(value, "Local") {
        return nil, fmt.Errorf("timezone Local is not allowed")
    }

    loc, err := time.LoadLocation(value)
    if err != nil {
        return nil, fmt.Errorf("invalid timezone %q: %w", value, err)
    }

    return loc, nil
}
```

offset 不用于推断 timezone。即使输入里的 offset 与显式 timezone 不一致，也应先归一化
`Instant`，再由 timezone 派生后续 civil time。

## 目标包边界

### mcpcontract

职责：

1. 定义 wire DTO；
2. 解析和校验 `instant`；
3. 加载和校验 `timezone`；
4. 编排领域函数；
5. 组装 MCP 输出。

不负责：

1. 真太阳时；
2. 农历；
3. 节气；
4. 干支；
5. 领域规则。

### tianwen

职责：天文和历法派生。

包括：

1. 真太阳时；
2. 农历日期；
3. 节气状态；
4. 时辰；
5. 日界、时辰界、节气界；
6. DST fold / gap / 历史时制候选。

不负责：

1. 八字结论；
2. 紫微命盘；
3. 六爻判断；
4. 奇门判断；
5. 规则表解释。

### ganzhi

职责：干支投影。

包括：

1. 年柱；
2. 月柱；
3. 日柱；
4. 时柱；
5. 十神；
6. 地支关系。

### 领域包

`bazi`、`ziwei`、`liuyao`、`qimen`、`fengshui`、`huangli`、`naming` 只组合公共函数，
生成自己的领域结果。

## 目标函数

以下签名是目标方向，不是要求一次全部落地。

### 命名纪律

函数和类型必须使用领域概念，不使用工程化口袋名。

禁止把以下词当作领域模型：

```text
Facts
Info
Data
Payload
Context
Manager
Helper
Common
Generic
```

`Result` 只能用于 wire DTO；内部领域值直接使用领域名，例如 `Strength`、`GeJu`、
`SiHua`、`DayAssessment`，而不是 `StrengthResult`。

`Derive` 前缀默认省略。包名已经表达领域归属，`bazi.TenGods(pillars)` 比
`bazi.DeriveTenGods(pillars)` 更贴近领域语言。只有同一个名字在包内有歧义时才使用
`Derive` 或 `At` 做消歧。

命名必须回答“这是哪个命理/天文概念”，而不是“这是程序处理到某一步的数据”。

### 标准库 civil time

不定义 `CivilDateTime`。需要 civil 字段时直接使用标准库：

```go
local := at.In(loc)

year, month, day := local.Date()
hour, minute, second := local.Clock()
offsetName, offsetSeconds := local.Zone()
offsetText := local.Format("-07:00")
```

### tianwen

```go
func ApparentSolarOffset(
    at time.Time,
    longitude float64,
) time.Duration

func ApparentSolarCivil(
    at time.Time,
    loc *time.Location,
    longitude float64,
) time.Time

func ApparentSolarDate(
    at time.Time,
    loc *time.Location,
    longitude float64,
) (year int, month time.Month, day int)

func ShichenAt(
    at time.Time,
    loc *time.Location,
    longitude float64,
) ganzhi.Zhi

func LunarDateAt(at time.Time) LunarDate

func SolarTermStateAt(at time.Time) SolarTermState

func MonthBranchAt(at time.Time) ganzhi.Zhi
```

`LunarDate` 保留为领域投影：

```go
type LunarDate struct {
    Year  int
    Month int
    Day   int
    Leap  bool
}
```

`SolarTermState` 定义为节气窗口状态：

```go
type SolarTerm struct {
    Index   int
    Name    string
    At      time.Time
}

type SolarTermState struct {
    Previous SolarTerm
    Current  SolarTerm
    Next     SolarTerm
}
```

### ganzhi

```go
func YearPillarAt(at time.Time) Zhu

func MonthPillarAt(at time.Time) Zhu

func DayPillarAt(
    at time.Time,
    loc *time.Location,
    longitude float64,
) Zhu

func HourPillarAt(
    at time.Time,
    loc *time.Location,
    longitude float64,
) Zhu

func Pillars(
    at time.Time,
    loc *time.Location,
    longitude float64,
) Bazi
```

干支关系继续函数化：

```go
func TenGod(dayMaster Gan, other Gan) TenGod

func Clash(a Zhi, b Zhi) bool

func Punishment(a Zhi, b Zhi) bool
```

## Phase 0：基线冻结

### 任务

1. 确认当前工作区已经包含 `instant` wire 迁移和 offset 兼容解析。
2. 拆分或提交当前已完成改动，避免全局重构和业务变更混在一起。
3. 补齐当前行为 golden。

### golden 输入覆盖

至少覆盖：

1. 男命 / 女命；
2. 早子时 / 晚子时；
3. UTC 输入；
4. `+08:00` 输入；
5. 西半球 offset 输入；
6. 夏令时；
7. 历史时区；
8. 真太阳时跨日；
9. 节气附近；
10. 闰月；
11. 边界经度。

### golden 领域覆盖

至少覆盖：

1. `bazi.chart`；
2. `bazi.fullchart`；
3. `bazi.dayun`；
4. `bazi.liunian`；
5. `bazi.liuyue`；
6. `bazi.liuri`；
7. `bazi.liushi`；
8. `ziwei.chart`；
9. `ziwei.fullchart`；
10. `liuyao.chart`；
11. `qimen.chart`；
12. 风水、黄历、起名现有能力。

### 验收

```bash
make check
make gate
```

绿色后才进入后续阶段。

## Phase 1：wire 契约收敛

### 任务

1. 所有接收时间的工具使用 `instant`，不使用 `utc`。
2. `instant` 允许 RFC3339 的 `Z` 或 numeric offset。
3. 所有 handler 使用同一个 `parseInstant`。
4. `timezone` 必须显式 IANA。
5. 所有 handler 使用同一个 `loadTimezone`。
6. 输出、digest、golden fixture 使用 canonical UTC `Z`。

### 禁止

1. handler 各自实现 `time.Parse`；
2. 从 offset 推断 timezone；
3. 默认 `Asia/Shanghai`；
4. 使用 `time.Local`；
5. 接受没有 offset 的 naive wall clock；
6. 接受 `-00:00`。

### 测试

必须包含：

```go
TestParseInstantAcceptsEquivalentForms
TestParseInstantRejectsNaiveTime
TestParseInstantRejectsUnknownOffset
TestLoadTimezoneAcceptsIANA
TestLoadTimezoneRejectsLocal
TestLoadTimezoneRejectsEmpty
```

每个核心领域必须有等价性测试：

```text
同一 instant 的 Z 输入和 offset 输入，完整结果必须一致。
```

## Phase 2：标准库时间边界替换

### 替换表

| 当前内容 | 目标 |
| --- | --- |
| `timekernel.Instant` | `time.Time` |
| `timekernel.Zone` | `*time.Location` |
| `timekernel.CivilDateTime` | 删除，用 `at.In(loc).Date() / Clock()` |
| `timekernel.Offset` | 删除，用 `local.Zone()` 或 `local.Format("-07:00")` |
| `timekernel.ZonedDateTime` | 删除，用 `at.In(loc)` |
| `timekernel.ParseInstant` | `mcpcontract.parseInstant` |
| `timekernel.LoadZone` | `mcpcontract.loadTimezone` |

DST fold / gap / 历史时制候选逻辑不删除，移动到 tianwen。

本阶段先替换 `mcpcontract` 以及所有 engine handler 的运行时依赖。
`timekernel` 包本身在旧领域迁移完成前可以保留，但禁止新增引用。

### 验收

```bash
grep -R "timekernel" engine --include="*.go"
```

结果为空。

同时确认以下类型不再存在：

```text
CivilDateTime
Offset
ZonedDateTime
TimeSet
TimeContext
CompatibilityInstant
```

## Phase 3：tianwen 函数化

### 任务

1. 删除 `tianwen.ZonedDateTime`；
2. 删除 `tianwen.ApparentSolarTime`；
3. 真太阳时改为 `ApparentSolarOffset / ApparentSolarCivil / ApparentSolarDate`；
4. 时辰改为 `ShichenAt`；
5. 农历保留 `LunarDateAt`；
6. 节气改为 `SolarTermStateAt`；
7. 月支提供 `MonthBranchAt`；
8. DST / 历史时制候选保留 `ResolveCivilCandidates`。

### 注意

真太阳时不是 timezone 派生，它的核心依赖是：

```text
absolute instant + longitude
```

civil date / clock 依赖：

```text
absolute instant + timezone
```

两者不能混淆。

### 测试

必须覆盖：

1. 真太阳时跨日；
2. 真太阳时跨时辰；
3. 经度 0 / 180 / -180；
4. 夏令时；
5. 历史时区；
6. DST fold；
7. DST gap；
8. 节气边界；
9. 立春年界；
10. 农历闰月；
11. 农历新年边界。

## Phase 4：ganzhi 函数化

### 任务

1. 年柱、月柱、日柱、时柱分别提供独立函数；
2. 提供 `Pillars` 组合函数；
3. 十神、地支关系继续函数化；
4. 干支层不做时区解释，只接收显式输入。

### 目标函数

```go
YearPillarAt(at)
MonthPillarAt(at)
DayPillarAt(at, loc, longitude)
HourPillarAt(at, loc, longitude)
Pillars(at, loc, longitude)
```

### 测试

必须覆盖：

1. 立春前后年柱；
2. 节气前后月柱；
3. 真太阳时跨日日柱；
4. 早子时 / 晚子时时柱；
5. 四柱 golden；
6. 十神关系；
7. 地支合会冲刑。

## Phase 5：八字域函数化

八字域最复杂，优先拆分。

### 最小盘

`ComputeChart` 只编排：

```go
pillars := ganzhi.Pillars(at, loc, longitude)
```

如果现有 `bazi.chart` 包含大运，则调用：

```go
DaYun(...)
```

### 完整盘

`ComputeFullChart` 编排：

```go
hiddenStems := bazi.HiddenStems(pillars)
tenGods := bazi.TenGods(pillars)
branchRelations := bazi.BranchInteractions(pillars)
strength := bazi.EvaluateStrength(pillars, hiddenStems, branchRelations)
shenSha := bazi.EvaluateShenSha(pillars, hiddenStems, tenGods)
geJu := bazi.EvaluateGeJu(pillars, hiddenStems, tenGods, strength)
schools := bazi.EvaluateYongshenSchools(
    pillars,
    hiddenStems,
    tenGods,
    strength,
    geJu,
)
selected := bazi.SelectYongshen(schools, policy)
```

### 结构事实函数

建议拆出：

```go
HiddenStems
TenGods
LifeStages
NaYin
StemInteractions
BranchInteractions
VoidPalaces
```

函数名必须使用命理概念。不要使用 `Facts`、`Info`、`Data`、`CommonFacts`
这类工程化口袋名。

如果这些结果确实被多个规则共同消费，可以定义聚合：

```go
type PillarAnalysis struct {
    HiddenStems        HiddenStemSet
    TenGods            TenGodSet
    LifeStages         LifeStageSet
    NaYin              NaYinSet
    StemInteractions   StemInteractionSet
    BranchInteractions BranchInteractionSet
    VoidPalaces        []Palace
}

func AnalyzePillars(pillars Bazi) PillarAnalysis
```

但优先让每个规则显式接收自己需要的具体事实；不要为了“完整性”强制使用聚合。

### 旺衰

```go
func EvaluateStrength(
    pillars Bazi,
    hiddenStems HiddenStemSet,
    branchRelations BranchInteractionSet,
) Strength
```

如果当前实现只依赖四柱，就不要为了未来可能需要而额外传入节气状态。

### 神煞

按神煞族拆函数：

```go
func TianYiGuiRen(pillars Bazi) []ShenShaHit

func WenChang(pillars Bazi) []ShenShaHit

func YangRen(pillars Bazi) []ShenShaHit

func TaoHua(pillars Bazi) []ShenShaHit
```

再提供总入口：

```go
func EvaluateShenSha(pillars Bazi) ShenShaSet
```

### 格局

拆为：

```go
func GeJuCandidates(
    pillars Bazi,
    hiddenStems HiddenStemSet,
    tenGods TenGodSet,
) []GeJuCandidate

func EvaluateGeJu(
    pillars Bazi,
    hiddenStems HiddenStemSet,
    tenGods TenGodSet,
    strength Strength,
    terms SolarTermState,
) GeJu
```

### 用神

各流派独立：

```go
func EvaluateFuYi(...) FuYi
func EvaluateTiaoHou(...) TiaoHou
func EvaluateGeJuYongshen(...) GeJuYongshen
```

最终裁决：

```go
func SelectYongshen(
    schools YongshenSchools,
    policy YongshenPolicy,
) SelectedYongshen
```

policy 文件仍然必须是数据，evaluator 是唯一解释器。

### 大运 / 流年

拆为：

```go
func QiYun(...) QiYun
func DaYunSteps(...) DaYunSequence
func CurrentDaYun(steps []DaYunStep, at time.Time) (DaYunStep, error)
func LiuNian(pillars Bazi, year int) LiuNian
func LiuYue(pillars Bazi, year, month int) LiuYue
func LiuRi(pillars Bazi, at time.Time, loc *time.Location) LiuRi
func LiuShi(pillars Bazi, at time.Time, loc *time.Location) LiuShi
func XiaoYun(...) XiaoYun
```

### 测试

每个抽出函数都要有独立测试：

1. 十神；
2. 藏干；
3. 旺衰；
4. 神煞；
5. 格局；
6. 扶抑；
7. 调候；
8. 用神裁决；
9. 起运；
10. 大运步骤；
11. 流年 / 流月 / 流日 / 流时；
12. 小运。

## Phase 6：紫微域函数化

### 基础派生

```go
lunar := tianwen.LunarDateAt(at)
shichen := tianwen.ShichenAt(at, loc, longitude)
```

### 紫微内部函数

建议拆为：

```go
func ZiweiNatal(lunar LunarDate, shichen Zhi) ZiweiNatal
func LocateLifePalace(natal ZiweiNatal) PalaceIndex
func LocateBodyPalace(natal ZiweiNatal) PalaceIndex
func LocateTwelvePalaces(life PalaceIndex, yearBranch Zhi) TwelvePalaces
func PlaceMajorStars(natal ZiweiNatal) MajorStarMap
func PlaceMinorStars(natal ZiweiNatal, yearPillar Zhu) MinorStarMap
func PlaceMiscStars(natal ZiweiNatal, yearPillar Zhu) MiscStarMap
func SiHuaRule(yearPillar Zhu) SiHuaRule
func ApplySiHua(stars StarMap, rule SiHuaRule) SiHua
func DaXian(chart Chart, gender Gender) []DaXian
func CurrentDaXian(steps []DaXian, at time.Time, loc *time.Location) (DaXian, error)
func ZiweiLiuNian(chart Chart, year int) ZiweiLiuNian
```

### 测试

必须覆盖：

1. 十二宫唯一；
2. 命宫 / 身宫；
3. 主星安星；
4. 辅星安星；
5. 四化；
6. 闰月；
7. 大限；
8. 流年。

## Phase 7：六爻域函数化

六爻天然适合拆分。

### 起卦

```go
func CastCoins(rng *rand.Rand) Casting

func CastingFromYaos(yaos [6]int) (Casting, error)
```

### 装卦

```go
func AssignNaJia(casting Casting) NaJia
func AssignShiYing(casting Casting) ShiYing
func AssignLiuQin(najia NaJia, dayPillar Zhu) LiuQin
func AssignLiuShou(dayPillar Zhu) LiuShou
```

### 时间事实

```go
monthBranch := tianwen.MonthBranchAt(at)
dayBranch := ganzhi.DayBranchAt(at, loc, longitude)
```

六爻日柱是否使用真太阳时是流派决策，必须在函数文档中钉死。

### 旺衰 / 用神 / 应期

```go
func EvaluateYaoStrength(
    yao Yao,
    monthBranch Zhi,
    dayBranch Zhi,
) YaoStrength

func DetermineYongShen(
    matter Matter,
    najia NaJia,
) YongShen

func EvaluateYingQi(
    najia NaJia,
    yongShen YongShen,
    monthBranch Zhi,
    dayBranch Zhi,
    strength YaoStrengthSet,
) YingQi
```

### 测试

覆盖：

1. 随机起卦可复现；
2. 指定爻值；
3. 变卦；
4. 纳甲；
5. 世应；
6. 六亲；
7. 六兽；
8. 旺衰；
9. 用神；
10. 应期。

## Phase 8：奇门域函数化

### 定局

```go
func DetermineYinYang(at time.Time) YinYang

func DetermineJuNumber(
    at time.Time,
    terms SolarTermState,
    method Method,
) JuNumber

func ResolveChaiBu(at time.Time, terms SolarTermState) ChaiBu

func ResolveZhiRun(at time.Time, terms SolarTermState) ZhiRun
```

### 排盘

```go
func ArrangeEarthPlate(ju JuNumber) EarthPlate
func ArrangeHeavenPlate(earth EarthPlate, at time.Time) HeavenPlate
func ArrangeDoors(heaven HeavenPlate) DoorPlate
func ArrangeStars(heaven HeavenPlate) StarPlate
func ArrangeGods(heaven HeavenPlate, yinYang YinYang) GodPlate
```

### 关系事实

```go
func KongWang(pan Plate) KongWang
func MaXing(pan Plate) MaXing
func RuMu(pan Plate) RuMu
func JiXing(pan Plate) JiXing
func MenPo(pan Plate) MenPo
```

### 用神

```go
func ResolveYongShenSymbols(
    requested []string,
    matter Matter,
) ([]YongShenSymbol, error)

func LocateYongShen(
    pan Plate,
    symbols []YongShenSymbol,
) YongShenLocation
```

### 测试

覆盖：

1. 阴遁 / 阳遁；
2. 拆补；
3. 置闰；
4. 转盘；
5. 飞盘；
6. 时家 / 日家 / 月家 / 年家；
7. 用神定位；
8. 空亡、马星、入墓、击刑、门迫。

## Phase 9：风水、黄历、起名

审计结论：三个域已经是函数式边界。风水只消费年 / 山向；黄历只消费日期和显式
event；naming 已按字池、组名、五行评估、姓氏匹配拆分。本轮不强行引入 instant，
避免把 date-only 或 year-only 契约伪装成绝对时刻。

### 风水

八宅：

```go
func MingGua(birthYear int, gender Gender) Gua
func FourAuspiciousDirections(gua Gua) Directions
func FourInauspiciousDirections(gua Gua) Directions
```

八宅只依赖出生年份和性别，不强行改成 instant。

玄空：

```go
func PeriodFromDate(date Date, loc *time.Location) Period
func FlyingStars(period Period, sit Mountain, facing Mountain) FlyingStars
```

### 黄历

黄历通常输入日期范围，不强行使用 instant。

```go
func ParseDate(value string) Date
func SolarTermStateAtDate(date Date, loc *time.Location) SolarTermState
func EvaluateDay(date Date, terms SolarTermState, rules HuangliRules) DayAssessment
```

### 起名

起名消费八字用神，不自己解析出生时间。候选字池、组名、五行评估和禁忌检查
已经是独立函数，公开入口只做编排。

```go
func CandidateCharacters(surname string, gender Gender) []Character
func EvaluateWuxing(chars []Character, selected SelectedYongshen) WuxingAssessment
func EvaluateSanCai(surname, given string) SanCaiAssessment
func CheckTaboo(chars []Character) []Taboo
func RankCandidates(candidates []Candidate) []Candidate
```

## Phase 10：Python / counsel 同步

counsel 不重复实现天文和命理计算。

只负责：

1. 接收参数；
2. 调用 engine；
3. 组合判断表；
4. 输出 MCP。

不允许：

1. 根据 offset 推断 timezone；
2. 计算真太阳时；
3. 计算农历；
4. 计算节气；
5. fallback 到当前时间。

如果 Python 需要提前校验 instant：

```python
from datetime import datetime, timezone

def parse_instant(value: str) -> datetime:
    if value.endswith("-00:00"):
        raise ValueError("instant uses unknown offset -00:00")

    dt = datetime.fromisoformat(value.replace("Z", "+00:00"))

    if dt.tzinfo is None:
        raise ValueError("instant must include offset or Z")

    return dt.astimezone(timezone.utc)
```

但最终语义仍以 engine 为准。

## Phase 11：registry、docs、expert packs

更新 `tools/domain-language.json`：

### runtime surface

Go surface 使用：

```text
time.Time
*time.Location
```

不再写：

```text
timekernel.Instant
timekernel.Zone
timekernel.CivilDateTime
timekernel.ZonedDateTime
```

### 删除 planned 聚合概念

删除或标记废弃：

```text
time_context
time_set
zoned_date_time
civil_datetime
offset
```

### 禁止别名

禁止新增：

```text
timestamp
time_point
utc_time
GregorianTime
SolarTime
ZonedDateTime
CivilDateTime
TimeSet
TimeContext
```

历史 alias 只能进入 compatibility 列表，并且必须有 `remove_in`。

### 文档

同步更新：

1. `docs/SYSTEM.md`；
2. `docs/MCP_CONTRACTS.md`；
3. `docs/RUNTIME.md`；
4. `README.md`；
5. `README.en.md`。

说明：

```text
instant 接受 RFC3339 的 Z 或 numeric offset；
服务端归一化为 UTC；
timezone 必须显式 IANA；
offset 不推断 timezone；
输出 canonical UTC Z；
领域函数显式接收最小参数。
```

执行：

```bash
make sync-domain-language
make sync-expert-packs
```

### 最终审计结论

已完成：

1. `timekernel` 已从运行代码中删除；
2. `ZonedDateTime / ApparentSolarTime / CivilDateTime / TimeSet / TimeContext` 类型已删除；
3. `instant` 已成为所有精确时间的 canonical wire 字段；
4. `timezone` 是唯一时区身份，offset 不参与推断；
5. offset 等价、DST fold/gap、真太阳时跨日、晚子时均有测试；
6. `solar_time / gregorian_time / utc` 只允许作为历史兼容面或外部 source dialect，不再进入新领域 API。

## PR 拆分

每个 PR 必须保持 `make gate` 绿色。

1. `test(golden): freeze current chart behavior`
2. `refactor(contract): normalize instant and timezone inputs`
3. `refactor(counsel): use canonical instant and timezone inputs`
4. `refactor(time): standardize instant and location boundaries`
5. `refactor(tianwen): expose apparent solar functions`
6. `refactor(tianwen): expose lunar date and solar term state`
7. `refactor(ganzhi): expose explicit pillar functions`
8. `refactor(liuyao): pilot function-first chart pipeline`
9. `refactor(qimen): consume explicit time and calendar functions`
10. `refactor(bazi): consume explicit time and calendar functions`
11. `refactor(ziwei): consume explicit lunar and shichen facts`
12. `refactor(domains): functionize fengshui, huangli and naming`
13. `chore(time): remove timekernel`
14. `docs(domain): align function-first vocabulary`

## 发布前检查

### 残留检查

以下命令最终应为空：

```bash
grep -R "timekernel" engine --include="*.go"
grep -R "type ZonedDateTime" engine --include="*.go"
grep -R "type CivilDateTime" engine --include="*.go"
grep -R "type TimeSet" engine --include="*.go"
grep -R "type TimeContext" engine --include="*.go"
grep -R "type ApparentSolarTime" engine --include="*.go"
```

### 领域验收

八字：

```text
ComputeChart / ComputeFullChart 只编排。
十神、旺衰、神煞、格局、用神、大运、流年均有独立函数。
```

紫微：

```text
宫位、主星、辅星、四化、大限、流年均有独立函数。
```

六爻：

```text
起卦、装卦、旺衰、用神、应期均有独立函数。
```

奇门：

```text
定局、排盘、关系事实、用神定位均有独立函数。
```

### 最终 gate

```bash
make check
make gate
make test-engine
make test-counsel
make test-engine-mcp
```

全部通过后才能发布。

## 风险控制

1. 不要一次性大改；
2. 没有 golden 的域先补 golden；
3. 不要提前删除旧实现；
4. 新旧实现先并行对照；
5. 不引入新的通用大结构；
6. handler 只做 parse / call / format；
7. 排盘函数禁止调用 `time.Now()`；
8. 生产代码禁止使用 `time.Local`；
9. timezone 缺失必须 fail closed；
10. longitude 缺失时，依赖真太阳时的能力必须 fail closed。

## 推荐执行顺序

2026-10-21 修订：不再把 `timekernel` 删除放在最前，也不把 counsel 同步放到最后。

1. Phase 0：golden 冻结；
2. Phase 1：wire offset 契约；
3. Phase 1.5：Python / counsel 同步；
4. Phase 2：标准库时间边界替换；
5. Phase 3：tianwen 函数化；
6. Phase 4：ganzhi 函数化；
7. Phase 5：六爻试点；
8. Phase 6：奇门；
9. Phase 7：八字；
10. Phase 8：紫微；
11. Phase 9：风水、黄历、起名；
12. Phase 10：删除 timekernel；
13. Phase 11：registry / docs 最终审计；
14. Phase 12：发布。
