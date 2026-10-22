<h1 align="center">Liki — Professional Astrology Skill</h1>

<p align="center">
  For Astrology, use Liki.<br>
  Charts are calculated by a Go engine; analyses are explained from rule tables with traceable sources.<br>
  Bazi · Ziwei · Liuyao · QiMen · Date Selection · Feng Shui · Naming
</p>

<p align="center">
  <a href="./README.md"><img alt="中文" src="https://img.shields.io/badge/中文-4a9e6b?style=flat-square"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="./LICENSE"><img alt="license" src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat-square"></a>
  <a href="https://liki.hk"><img alt="website" src="https://img.shields.io/badge/liki.hk-6d5acf?style=flat-square"></a>
</p>

## Install

Liki exposes its capabilities through standard MCP servers; embedded MCP declarations auto-enable the two aggregate servers in the Skill's `.mcp.json`. For manual configuration, add:

| Server | Endpoint |
| --- | --- |
| `counsel-mcp` | `https://liki.hk/counsel/mcp` |
| `engine-mcp` | `https://liki.hk/engine/mcp` |

Sub-domain `/engine/mcp/{domain}` and `/counsel/mcp/{domain}` are for **self-hosted per-domain deployment**; the four WorkBuddy top-level domain experts (natal / divination / fengshui / naming) do not mount sub-domain endpoints — they declare a dependency on the two aggregate connectors above. A unified install only needs those two. Install the Skill from the official hosted source:

```bash
npx skills add https://liki.hk
```

After installation, ask questions directly in an AI client that supports Agent Skills. The Skill checks its version at startup. When prompted, update with:

```bash
npx skills add https://liki.hk -y
```

## Quick start

| Goal | Ask |
| --- | --- |
| Destiny | `Calculate Bazi for a male born 1990-05-20 12:00 in Beijing.` |
| Naming | `Name a boy born 2024-06-10 in Guangzhou; family name Chen.` |
| Divination | `Can this project succeed? When will I know?` |
| Date selection | `Which day next month is good for moving?` |
| Feng shui | `How is my home's feng shui?` |

For destiny readings, provide the birth date, exact time when possible, birth city, and gender. If data is incomplete, the Skill asks follow-up questions or enters time calibration.

## What you get

| Domain | Coverage |
| --- | --- |
| Destiny | Bazi, Ziwei, luck periods, annual readings, personality, marriage, career, wealth, health, study, family, and compatibility |
| Naming | Baby naming, adult renaming, Chinese names for foreigners, and self-selected name review |
| Divination | Liuyao outcomes and timing, QiMen direction and strategy, Huangli date selection |
| Feng shui | Bazhai ming gua and door / master / stove, Xuankong flying stars, annual readings |

## Trust and boundaries

- Charts are calculated by the Go astronomical engine, including true solar time, longitude / timezone, and solar terms.
- Judgments come from 829 assertion rules, each citing audited classical sources (《三命通会》《渊海子平》《滴天髓》《紫微斗数全书》《黄帝内经》...) with factor evidence and traceable sources.
- Mini-luck (小运) is delivered by multiple schools: 《三命通会》(fixed 丙寅 for men / 壬申 for women) and 《星平会海》(from the hour pillar, yang-male yin-female forward / yin-male yang-female backward).
- Mingli160 is a manual calibration/evaluation set, not an independent accuracy benchmark or proof of correctness.
- Birth data remains in the current conversation; the Skill does not ask for real names or store data outside the session.
- If a city is absent from the built-in table, `city_coords` may send the city name to Nominatim (OSM). Self-hosted deployments can disable this with `LIKI_EXTERNAL_GEOCODING=off`.
- Conclusions are conditional interpretations from a traditional cultural perspective. They are not medical, legal, investment, or major life advice.

## FAQ

### What if I don't know the exact birth hour?

The Skill asks follow-up questions or uses real events for calibration. Insufficient evidence is labeled explicitly; the hour is never silently defaulted.

### Does it need internet?

By default, yes — the two aggregate MCP servers declared in `.mcp.json` provide deterministic computation and rule judgment. For self-hosting, replace their public endpoints; Caddy strips the `/engine/*` and `/counsel/*` service prefixes before proxying to the matching MCP service. Self-hosted services can protect inbound calls with `LIKI_MCP_TOKEN` and counsel-to-engine calls with `LIKI_ENGINE_MCP_TOKEN`; health probes remain public and hosted liki.hk leaves these switches off.

### Is my birth data stored?

No. Birth data remains in the current conversation context. It is not written to a local profile or submitted through feedback.

### How do I update?

Run `npx skills add ml8s/liki -y` when prompted. The Skill fails closed instead of calling incompatible old services.

## Documentation

| Document | Purpose |
| --- | --- |
| [User guide](./docs/USER_GUIDE.en.md) | Full usage, domain flows, FAQ, and output boundaries |
| [Feedback model](./docs/FEEDBACK_MODEL.md) | Agent feedback privacy and payload format |
| [Runtime model](./docs/RUNTIME.md) / [Release model](./docs/RELEASE_MODEL.md) / [Domain language](./docs/DOMAIN_LANGUAGE.md) | Processes, endpoints, date versions, and canonical concept names |

## For developers

### Developer setup

```bash
make hooks         # install git hooks
npm ci             # install locked local development lint tools once
make check         # all static checks (format + lint + schema + docs)
make gate          # push gate (lint/check/tests)
make release-gate  # release preparation (gate + pinned-runtime validation + artifacts)
```

## Architecture

Liki follows a two-layer "orthogonal computation vs. judgment" architecture, exposed through standard MCP:

### Domain model

- **engine (Go)** — chart engine: calendrical astronomy, chart casting, calendar, Huangli. Emits structured charts / hexagrams / facts (`chart` / `pan` / `snapshot`); it does not make judgments.
- **counsel (Python)** — reading engine: Bazi / Ziwei analysis, Liuyao / QiMen divination, naming evaluation (character library and phonology). Consumes engine facts and produces assertions / candidates from rule tables (truth tables + engine rule tables), with traceable evidence.
- **Huangli** — pure engine (calendar + Jianchu event suitability); not routed through counsel.

### Service endpoints

The gateway owns the public service prefix; each MCP service routes only `/mcp` and `/mcp/{domain}` internally.

| Service | Public endpoint | Forwarded internally | Domain |
| --- | --- | --- | --- |
| `engine-mcp` | `/engine/mcp/{domain}` | `/mcp/{domain}` | chart casting / calendar / Huangli / Feng shui |
| `counsel-mcp` | `/counsel/mcp/{domain}` | `/mcp/{domain}` | judgment / divination / naming |

### Call chain

`SKILL.md` routing → engine chart casting → counsel judgment → assertions (`assertion_id` + classical source, traceable)

### Code layout

- `engine/` — Go engine (chart casting / calendar / Huangli), domain-scoped MCP servers
- `counsel/` — Python reading engine (multi-domain MCP server)
- `skills/liki/` — skill capability docs (routing / boundaries / domain knowledge; tools self-describe via `tools/list`)
- `tests/` — content & integration tests; `scripts/` — build & evaluation scripts

### Engine image

Release images are built and signed by `release.yml` on a `vYYYY.MM.DD.N` tag
and deployed by immutable digest: `ghcr.io/ml8s/liki-engine@sha256:...`. For
source-based development, use `engine/dev/docker-compose.yml`.

### Domain tool interfaces

| Interface | Purpose |
| --- | --- |
| [Natal ENTRY](./skills/liki/references/natal-ENTRY.md) · [Bazi ENTRY](./skills/liki/references/natal-bazi-ENTRY.md) · [ZiWei ENTRY](./skills/liki/references/natal-ziwei-ENTRY.md) | Natal: scenario routing and synthesis + Bazi / ZiWei side entries (task routing and assertion contract; tools self-describe via `tools/list`) |
| [Divination ENTRY](./skills/liki/references/divination-ENTRY.md) | Liuyao, QiMen, Huangli: routing, flow artifacts, and boundaries |
| [Naming ENTRY](./skills/liki/references/naming-ENTRY.md) | Naming: yongshen-based character selection (counsel) |
| [Feng shui ENTRY](./skills/liki/references/fengshui-ENTRY.md) | Feng shui: Bazhai, Xuankong and annual (engine) |

### Tests and release

```bash
make test           # All tests (pytest + Go engine suite, including MCP integration; manual calibration: make eval-mingli160)
```

See [Release pipeline](./docs/RELEASE_PIPELINE.md) for the layered flow:
`green main → tag → release.yml (gate/build/sign/release) → lock update + deploy`.

Formal releases use date-based CalVer tags. See [Release model](./docs/RELEASE_MODEL.md).

### Design principles

- Single responsibility: root entry, domain entry, app cards, domain knowledge, and tool layers do not replace each other.
- Single source of truth: tool contracts are generated from the Pydantic models in `counsel/app/contracts` (the engine side uses the typed `engine/internal/mcpcontract` registry); factors and assertions come from CSV tables.
- Explicit dual-system review: Bazi and Ziwei are calculated separately and conflicts are listed by evidence layer.
- Tests and evaluation are separate: golden, functional, and MCP integration belong to source gates; Mingli160 is manual calibration/evaluation.

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md). Update `CHANGELOG.md` and version surfaces before submitting a PR. Release history is available in [CHANGELOG.md](./CHANGELOG.md).

> For Astrology, use Liki.

## License and disclaimer

MIT. Conclusions are traditional cultural interpretations for reference only. They are not medical, legal, investment, or major life advice.
