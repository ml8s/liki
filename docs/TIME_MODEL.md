# Time model

The public engine time model has two boundaries:

1. **Recorded wall-clock input** goes to `tianwen.resolve` as a civil date/time
   (or lunar date, shichen, or range) plus an explicit IANA timezone and
   longitude.
2. **Chart and flow tools** accept only `standard_time` plus longitude. They do
   not accept `instant`, `civil_time`, `timezone`, or `fold`.

## Canonical input

Birth/event resolution:

```json
{
  "date": "1990-05-20",
  "time": "12:00",
  "timezone": "Asia/Shanghai",
  "longitude": 116.4
}
```

Chart and flow tools:

```json
{
  "standard_time": "1990-05-20T12:00:00+08:00",
  "longitude": 116.4
}
```

- `standard_time` is RFC3339 with a known offset. `Z` and numeric offsets
  such as `+08:00` / `-05:00` are accepted; `-00:00` is rejected because
  RFC3339 defines it as an unknown local offset.
- It is local **standard** time: instant plus the zone's standard offset, with
  no DST.
- The offset is not a timezone identity, but it does select the local standard
  representation used by date facts. Chart tools never infer an IANA zone from
  it. Two timestamps can identify the same instant while selecting different
  local standard dates.
- `longitude` is required only for capabilities that use local apparent solar
  time (LAST).
- `tianwen.resolve` is the only place where civil time and DST exist.

## DST boundary

`tianwen.resolve` converts civil wall time to an instant once:

- Normal time: one candidate.
- Fold: two candidates; the caller must disambiguate.
- Gap: error; no silent shift.

Every candidate is returned as `standard_time`. Downstream tools never apply
DST again.

## Projection matrix

| Fact | Projection | Rule |
| --- | --- | --- |
| Solar-term / 立春 boundary | absolute event | Compare `standard.UTC()` with the solar-term instant. The boundary is global. |
| Solar-term date label | standard date | Date of the event at the recorded place's standard offset. |
| Lunar date | local astronomical | New moons and 中气 are assigned to the recorded place's standard date. |
| Almanac date | standard date | 通书/黄历 uses the target standard date. |
| Bazi year/month pillar | solar-term event | The year changes at 立春; the month changes at 节. |
| Bazi day/hour pillar | local LAST | Local apparent solar time from `standard + longitude`; late 子时 rule applies. |
| 流日 / 流时 | standard date + LAST | The day label uses the standard date; the hour uses LAST. |
| 八宅 命卦 year | 立春 event | Derive the 立春 year from `standard.UTC()`. |
| 玄空 元运/年星 | 立春 event | Derive the 立春 year from `standard.UTC()`. |
| 奇门 起局/年命 | standard + LAST | 局 and 节气 use the event; the hour uses LAST; birth uses `birth_standard_time`. |
| Ziwei natal lunar date | local astronomical | Standard date + local-zone new moon/中气 calendar. |
| Ziwei flow lunar date | local astronomical | Same projection as natal. |
| Ziwei shichen | local LAST | Local apparent solar time. |

## Standard time and LAST

- Standard time: `instant + standard_offset`. It never contains DST.
- LAST: `standard + longitude_correction + equation_of_time`.
- Astronomy uses `standard.UTC()`.
- Date facts use `standard.Date()` at the recorded place's fixed offset.
- Day pillar changes at 00:00 LAST; the late 子时 uses the next day's day gan
  for the hour pillar.
- 流时 changes at the LAST shichen boundary.

## Lunar calendar

- A lunar month is leap when it contains no 中气 date in
  `[monthStart, nextMonthStart)`.
- A 中气 on the next month's start date belongs to the next month.
- New moons and 中气 are astronomical events; their **date assignment** uses
  the standard date.
- Date-to-JD arithmetic uses a timezone-free civil-date serial, so DST cannot
  create 23/25-hour "days".
- Solar terms use the ShouXing VSOP87 apparent longitude with the ShouXing ΔT
  table; new moons use Meeus lunation terms with the same ΔT.

The official UTC+8 Chinese calendar and the iztro `local_label` convention are
internal compatibility references for tests and oracles. They are not exposed
as 命理 API options. A fixed UTC+8 `LunarReference` is equivalent to the
official Chinese calendar; `Asia/Shanghai` is not identical during historical
DST.

## Output echo

Time-bearing tools echo the projections they actually used:

```json
"resolved": {
  "standard_time": "2026-02-04T11:00:00+08:00",
  "local_date": "2026-02-04",
  "last_date": "2026-02-04",
  "shichen": "午",
  "lichun_year": 2026,
  "boundary_day": true
}
```

Only include fields that participated in the computation. The echoed period
must be the same value passed to the compute stage; it is not recomputed in the
wire layer.

## Boundary rules

- 立春: the exact event decides.
- 节: the exact event decides the 节气 month.
- 农历: the local astronomical date changes at the standard midnight.
- 八字 day: the day pillar changes at 00:00 LAST.
- 流时: the shichen changes at the LAST 时辰 boundary.
- 奇门: 节气/置闰 use the event; the hour pillar uses LAST.

## Anti-patterns

- Do not pass `instant` or `civil_time` to chart tools.
- Do not apply DST a second time after `tianwen.resolve`.
- Do not infer an IANA timezone from a `standard_time` offset.
- Do not use a Beijing date label for a birth day/hour pillar.
- Do not use a Beijing date label for local 命理 facts; use the recorded
  place's standard time and, for LAST, its longitude.
- Do not pass a bare date to a boundary-sensitive chart tool.

## Enforcement

`tianwen.resolve` is the only tool that accepts civil time, lunar time, or an
IANA timezone. Chart and flow schemas require `standard_time`; the engine's
`TestTimeInputsCanonical` rejects reintroduced `instant`, `civil_time`,
`timezone`, `fold`, or bare-date inputs on those tools.
