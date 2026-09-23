<h1 align="center">Liki</h1>

<p align="center">
  A professional Skill for Chinese Metaphysics<br>
  懂命理，用 Liki。 For Chinese Metaphysics, use Liki.<br>
  Charts are calculated by a Go engine; judgments are explained from rule tables with traceable sources.<br>
  Bazi · Ziwei · Liuyao · QiMen · Date Selection · Feng Shui · Naming
</p>

<p align="center">
  <a href="./README.md"><img alt="中文" src="https://img.shields.io/badge/中文-4a9e6b?style=flat-square"></a>
  <a href="https://github.com/ml8s/liki/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/ml8s/liki/actions/workflows/ci.yml/badge.svg"></a>
  <a href="./LICENSE"><img alt="license" src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat-square"></a>
  <a href="https://liki.hk"><img alt="website" src="https://img.shields.io/badge/liki.hk-6d5acf?style=flat-square"></a>
</p>

## Install

```bash
npx skills add ml8s/liki
```

After installation, ask questions directly in an AI client that supports Agent Skills. The Skill checks its version at startup. When prompted, update with:

```bash
npx skills add ml8s/liki -y
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
- Judgments come from 799 assertion rules and preserve factor evidence and classical sources.
- 160 professional competition questions provide independent accuracy evaluation with isolated answers.
- Birth data remains in the current conversation; the Skill does not ask for real names or store data outside the session.
- Conclusions are conditional interpretations from a traditional cultural perspective. They are not medical, legal, investment, or major life advice.

## FAQ

### What if I don't know the exact birth hour?

The Skill asks follow-up questions or uses real events for calibration. Insufficient evidence is labeled explicitly; the hour is never silently defaulted.

### Does it need internet?

By default, yes. The JSON-RPC engine performs calendar and chart calculations. Advanced users can run a private engine and set `LIKI_RPC_URL`.

### Is my birth data stored?

No. Birth data remains in the current conversation context. It is not written to a local profile or submitted through feedback.

### How do I update?

Run `npx skills add ml8s/liki -y` when prompted. The Skill fails closed instead of calling incompatible old RPCs.

## Documentation

| Document | Purpose |
| --- | --- |
| [User guide](./docs/USER_GUIDE.en.md) | Full usage, domain flows, FAQ, and output boundaries |
| [README style](./docs/README_STYLE.md) | Structure, heading, and formatting contract for both READMEs |
| [Feedback model](./docs/FEEDBACK_MODEL.md) | Agent feedback privacy and contract |
| [Release model](./docs/RELEASE_MODEL.md) | CalVer runtime versions and SemVer releases |

## For developers

### Developer setup

```bash
make hooks         # install git hooks
make check         # all static checks (format + lint + schema + docs)
make gate          # local push gate (lint + check + test, ~3min)
make build-archive # pack the unified Liki skill
```

### Architecture

```text
skills/liki/
├── SKILL.md              # single skill entry: routing, safety, feedback
├── VERSION.txt           # single distribution version
├── FAQ.md                # runtime failure and recovery contract
├── natal/                # Birth-chart composite: Bazi + Ziwei + combined analysis
├── divination/           # Liuyao + QiMen + Huangli: ENTRY / TOOLS / app / domains / tools
├── fengshui/             # Bazhai + Xuankong: ENTRY / RPC / app / domains
└── naming/               # naming: ENTRY / RPC / app / domains
```

The repository root keeps `engine/`, `tests/`, and `scripts/` for the engine, evaluations, and build scripts; the installable package comes only from `skills/liki`. The call chain is fixed: `SKILL.md` → `ENTRY.md` → app card → Python tools or fixed RPC. Natal and divination use domain-local Python tools to orchestrate RPC, snapshots, factors, and assertions. Naming and feng shui have no local Python tool layer and use fixed JSON-RPC payloads.

### Engine image

The engine image is published with GitHub Releases: `docker pull ghcr.io/ml8s/liki-engine:latest`. Build from source with `engine/deploy/docker-compose.yml`.

### Domain contracts

| Contract | Purpose |
| --- | --- |
| [Natal tools](./skills/liki/natal/TOOLS.md) | Complete stdin payloads for five natal analysis tools |
| [Divination tools](./skills/liki/divination/TOOLS.md) | Liuyao, QiMen, and Huangli tool payloads |
| [Naming ENTRY](./skills/liki/naming/ENTRY.md) | Naming: yongshen-based character selection (engine-pro MCP tools) |
| [Feng shui ENTRY](./skills/liki/fengshui/ENTRY.md) | Feng shui: Bazhai, Xuankong and annual (engine MCP tools) |

### Tests and release

```bash
make test           # All tests (pytest + Go engine full suite)
make verify        # end-to-end integration tests
make golden # full golden suite
```

See [Release model](./docs/RELEASE_MODEL.md) for the layered model: `lint → check → test → verify → gate`.

Formal releases use SemVer tags; runtime compatibility uses CalVer. See [Release model](./docs/RELEASE_MODEL.md).

### Design principles

- Single responsibility: root entry, domain entry, app cards, domain knowledge, and tool layers do not replace each other.
- Single source of truth: tool contracts come from `skill-tools.json`; factors and assertions come from CSV tables.
- Explicit dual-system review: Bazi and Ziwei are calculated separately and conflicts are listed by evidence layer.
- Evaluation-driven: golden, functional, integration, skill-up smoke, and the 160-question benchmark run in separate layers.

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md). Update `CHANGELOG.md` and version contracts before submitting a PR. Release history is available in [CHANGELOG.md](./CHANGELOG.md).

> 懂命理，用 Liki。

## License and disclaimer

MIT. Conclusions are traditional cultural interpretations for reference only. They are not medical, legal, investment, or major life advice.
