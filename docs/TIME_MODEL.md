# Time model

Engine time inputs are canonical and domain-derived. The engine never asks the
caller to choose a timezone projection per fact: it receives one absolute
instant plus a timezone (and longitude where true-solar time matters), then
projects the instant differently for each domain fact.

## Canonical input

```json
{
  "instant": "1990-05-20T12:00:00+08:00",
  "timezone": "Asia/Shanghai",
  "longitude": 116.4
}
```

- `instant` is an absolute RFC3339 instant; numeric offsets are accepted and
  normalized to UTC.
- `timezone` is an explicit IANA zone.
- `longitude` is required only for capabilities that use local apparent solar
  time (LAST).
- A calendar date is a derived projection, not the canonical input. Application
  layers may accept a date for convenience and convert it to an instant at a
  declared reference time before calling the engine.

## Projection matrix

| Fact | Projection | Rule |
| --- | --- | --- |
| Solar-term / 立春 boundary | absolute instant | Compare the instant with the solar-term instant. The boundary is global. |
| Solar-term date label | local civil date | Date label of the solar-term instant in the caller's zone. |
| Lunar date | local astronomical | New moons and 中气 are assigned to the caller's local civil date. |
| Almanac date | local civil date | 通书/黄历 uses the target local date. |
| Bazi year/month pillar boundary | absolute instant | The year changes at 立春; the month changes at 节. |
| Bazi day/hour pillar | local LAST | Local apparent solar time; the late 子时 rule applies. |
| 流时 shichen | local LAST | Local apparent solar time. |
| 八宅 命卦 year | absolute instant | Derive the 立春 year from the instant. |
| 玄空 元运/年星 | absolute instant | Derive the 立春 year from the instant. |
| 奇门 起局/年命 | absolute instant | The 局 and 年命 use the instant and the 立春 year. |
| Ziwei natal lunar date | local astronomical | Local civil date + local-zone new moon/中气 calendar. |
| Ziwei flow lunar date | local astronomical | Same projection as natal. |
| Ziwei shichen | local LAST | Local apparent solar time. |

`date` inputs are not part of the engine contract. A date is echoed as
`resolved.local_date`, `resolved.beijing_date`, or a domain period label.
`beijing_date` is a display echo, not the lunar projection.

## Local time and astronomical instants

命理 time facts are local; astronomical events are absolute instants.

- Canonical input: `instant + timezone + longitude`.
- Local civil date/time: `instant.In(timezone)`.
- Local lunar date: assign each new moon and 中气 to the local civil date.
- LAST / shichen: local apparent solar time from `timezone + longitude`.
- Solar terms, 立春, 冬至, new moons, and 中气 are global instants; only their
  date assignment uses the local zone.
- A lunar month is leap when it contains no 中气 date in
  `[monthStart, nextMonthStart)`. A 中气 on the next month's start date belongs
  to the next month.
- Date-to-JD arithmetic uses a timezone-free civil-date serial, so DST cannot
  create 23/25-hour "days".
- Solar terms use the ShouXing VSOP87 apparent longitude with the ShouXing
  ΔT table; new moons use Meeus lunation terms with the same ΔT.

The official UTC+8 Chinese calendar and the iztro `local_label` convention are
internal compatibility references for tests and oracles. They are not exposed
as 命理 API options. A fixed UTC+8 `LunarReference` is equivalent to the
official Chinese calendar; `Asia/Shanghai` is not identical during historical
DST.

## Output echo

Every time-bearing tool returns the projections it actually used:

```json
"resolved": {
  "absolute_instant": "2026-02-04T03:00:00Z",
  "local_date": "2026-02-03",
  "beijing_date": "2026-02-04",
  "last_date": "2026-02-03",
  "shichen": "亥",
  "year_boundary": "lichun",
  "resolved_year": 2025,
  "boundary_day": true
}
```

Only include fields that participated in the computation. The echoed period
must be the same value passed to the compute stage; it is not recomputed in the
wire layer.

## Boundary rules

- 立春: the exact instant decides. The 立春 date belongs to the new year for
  date-level projections; instant-level tools compare the instant.
- 节: the exact instant decides the 节气 month.
- 农历: the local astronomical date changes at the local civil midnight.
- 八字 day: the day pillar changes at 00:00 LAST; the late 子时 uses the next
  day's day gan for the hour pillar.
- 流时: the shichen changes at the LAST 时辰 boundary.

## Anti-patterns

- Do not convert the whole instant to Beijing time and derive every fact from
  that projection.
- Do not use a Beijing date label for a birth day/hour pillar.
- Do not use a Beijing date label for local 命理 facts; use the caller's
  timezone and, for LAST, longitude.
- Do not pass a bare date to a boundary-sensitive capability.

## Enforcement

All time-bearing engine tools use `instant + timezone` (plus `longitude` where
LAST is required). `TestTimeInputsCanonical` in `engine/internal/mcpcontract`
rejects reintroduced `date`, `year`, `lunar_year`, `target_lunar`, or `shi_zhi`
inputs.
