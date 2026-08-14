<p align="center">
  <img alt="Liki" src="https://img.shields.io/badge/Liki-Professional_Skill_for_Chinese_Metaphysics-6d5acf?style=for-the-badge&logo=openai&logoColor=white&labelColor=30305c">
</p>

<p align="center">
  <strong>Liki — Professional Skill for Chinese Metaphysics（v3.10.0）</strong>
</p>

<p align="center">
  <code>npx skills add ml8s/liki</code>
</p>

<p align="center">
  <a href="https://github.com/ml8s/liki"><img src="https://img.shields.io/badge/GitHub-ml8s/liki-4a9e6b?style=flat&logo=github&logoColor=white&labelColor=30305c"></a>
  <a href="https://liki.hk"><img src="https://img.shields.io/badge/liki.hk-website-6d5acf?style=flat&logo=safari&logoColor=white&labelColor=30305c"></a>
  <a href="./README.md"><img src="https://img.shields.io/badge/中文-4a9e6b?style=flat&logo=readme&logoColor=white&labelColor=30305c"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-4a9e6b?style=flat"></a>
</p>

---

**Liki** is a professional Skill for Chinese Metaphysics: **BaZi, ZiWei, naming, Liuyao, Qimen, date selection, and Feng Shui** — 8 domains in your AI assistant. It is not "let the AI free-style fortune-telling"; it turns traditional metaphysics into an engineered tool with **executable process, verifiable conclusions, and traceable judgments**.

- **For users**: every conclusion has four guarantees — chart casting never relies on the model (astronomical engine), judgments never rely on improvisation (46 truth tables, 701 rules with classical references), process never skips steps (gate-based checklists), and conclusions can be re-verified against your real life events (calibration).
- **For developers**: all judgment rules are CSV truth tables (readable, editable, reviewable), the evaluation harness ships with the skill (160 questions, answer-isolated, auto-graded, reproducible), and engine + skill are cleanly separated ([liki-engine](https://github.com/ml8s/liki-engine) is open source).

## What You Can Do

| Scenario | Example input | What Liki does |
|----------|--------------|----------------|
| Read a BaZi chart | `1990-05-20 12:00 Beijing, male` | True solar time correction → chart → strength/yongshen → geju → domain rules |
| Ask about career/marriage/wealth | `How's my career luck lately?` | Domain routing → spouse/officer star lookup → BaZi+ZiWei cross-check → liunian candidates |
| ZiWei chart | `ZiWei chart for 1988-03-15 Shanghai, female` | Lunar chart → 12 palaces & sihua → daxian/liunian |
| Name a baby | `Name my baby boy, born 2024-06-10 Guangzhou, surname Chen` | Yongshen five-element → wuge-sancai → candidate characters |
| Date selection / divination | `Should we move house tomorrow?` | Huangli + Bazhai/Xuankong → auspicious judgment |
| Full report | `Write me a full life report` | BaZi+ZiWei full pipeline → combined report |

**The difference from "AI that just tells fortunes"**: ask "what happened in 1996" — a plain AI answers from memory; Liki first calls the engine to chart the 1996 annual pillar, looks up truth-table rules, then gives event candidates **with metaphysical evidence at every step**.

## Why You Can Trust Liki

Reliability is not a slogan — five hard mechanisms:

**① Astronomical engine, no model charting** — BaZi/ZiWei charting is done by the open-source [liki-engine](https://github.com/ml8s/liki-engine): true solar time, DST, lat-lon timezone — all computed astronomically. The model only interprets; it is **forbidden to derive chart data itself**.

**② Truth-table-driven judgments** — 46 judgment tables (BaZi 26 + ZiWei 20, **701 rules**), each with a **classical reference column** (from 《渊海子平》《子平真诠》《滴天髓》《三命通会》《紫微斗数全书》 etc.). Rules are rules — readable, editable, reviewable. Plus 7 yearly tables (annual event domains, including annual star spirits).

**③ Gate-based execution, no skipped steps** — all question types follow a unified pipeline (Phase 0-8): routing → hour determination → chart snapshot → strength/yongshen → domain lookup → ZiWei cross-check → calibration. Each phase has a fill-in checklist (□ fill-in, not check-off); incomplete = invalid conclusion.

**④ Metaphysical depth: contextual factors eliminate rule conflicts** — the hard part of Chinese metaphysics is "multiple rules on one chart contradicting each other." Liki solves this systematically with contextual factors:
- **Month-branch master element defines personality** (《子平真诠》"the month branch is the outline; the ge-shen rules the nature") — personality questions get one primary face; ten-god strength rules are secondary only
- **Star-palace co-reference** (《三命通会》) — spouse star exposed (star auspicious) + spouse palace clashed (palace inauspicious) = marriage possible but turbulent
- **Annual star spirits** (《协纪辨方书》) — Bingfu/Sangmen/Diaoke/Dahao looked up by Tai-sui year, triggered when they land on natal pillars
- **Zero-conflict verification**: scanning all 19 domains for "contradictory rules hitting the same chart" → add contextual factors → zero collisions

**⑤ Calibration with real life events** — before charting, calibrate the birth hour with known events; before concluding, verify against 3-5 known life periods ("this period's triggers → what happened"), conclusion stands only if ≥2 periods match.

## Quick Start

```bash
npx skills add ml8s/liki
```

Then talk to your AI assistant:

> BaZi chart, 1990-05-20 12:00 Beijing, male
> ZiWei chart for 1988-03-15 Shanghai, female
> Name my baby, 2024-06-10 Guangzhou, male, surname Chen
> Should we move house tomorrow?

Or request a full report:

> Write me a full life report

The AI assistant runs the full BaZi+ZiWei pipeline and outputs a combined judgment + BaZi report + ZiWei report.

## Evaluation: Reproducible System, Data in Release Posts

Liki maintains an independent evaluation harness on [MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) (160 questions from real fortune-telling contests): **answers separated from questions, Docker sandbox isolation (the agent physically cannot read answers), skill-up auto-grading**, and every change is regression-comparable.

**Accuracy numbers / iteration records are not published in this README** — see release posts and CHANGELOG (to keep the README from becoming a promo page).

**Reproduce** (eval config and grading scripts ship with the skill; answer files are separated and moved out of the mounted directory at runtime):

```bash
cd skills/liki
skill-up validate tests/eval-grouped-qwen.yaml    # validate 32 cases
bash tests/run-qwen.sh --parallelism 16          # move answers → eval → restore → auto-grade
```

> The eval harness itself (160 questions, answer isolation, auto-grading, reproducibility) is the evidence of engineering rigor; any number ships with its eval config and is reproducible — no inflated claims.

## Architecture

```
┌─ Process layer   SKILL.md Phase 0-8 (gate-based) + app/ 13 cards (per-scenario routing)
├─ Rule layer      tools/ 46 truth tables (701 rules + classical refs) + factor engine (497 factor definitions)
├─ Tool layer      tools/ 5 functions (full_paipan / make_factors / query …)
└─ Engine layer    liki-engine (open-source JSON-RPC astronomical computation)
```

Data pipeline: **engine charts (astronomical) → factor generation (truth tables) → rule lookup (classical refs) → gate process → dual-system cross-check → calibration** — every link is auditable.

## Project Structure

```
├── SKILL.md    ← Execution backbone (Phase 0-8 routing + common rules + adjudication)
├── app/        ← Application layer (13 cards: chart/marriage/health/career/wealth/study/personality/family/divination/naming/fengshui/compatibility)
├── domains/    ← Domain layer (8 domains, rule + methodology docs)
├── tools/      ← Inference engine (charting / factors / rule lookup)
│   ├── bazi/       ← BaZi rule tables, 26 (incl. yearly_* 7)
│   ├── ziwei/      ← ZiWei rule tables, 20
│   ├── factors/    ← Factor definitions (natal 497 + liunian factors)
│   ├── paipan.py   ← Charting (full_paipan / liunian)
│   └── duanyu.py   ← Rule lookup (query / match)
├── tests/      ← Evaluation harness (160 grouped cases, answer separation, grading scripts)
└── webapp/     ← Web integration pipeline
```

## Design Principles

- **Domain/application separation** — domain layer holds rules (symbol→reality translation) and methods (analysis flow); application layer holds process cards. Fix knowledge without touching process; adjust process without touching applications.
- **Truth-table-driven** — CSV truth tables + classical reference column: rules are readable, editable, reviewable — not model memory.
- **Contextual factors eliminate conflicts** — month-branch master, star-palace co-reference, annual star spirits — zero-collision verified across domains.
- **Gate-based execution** — fill-in checklists force every step on paper; the model cannot skip.
- **Dual-system cross-validation** — BaZi leads, ZiWei reviews; conflicts resolved with explicit evidence.
- **Calibration loop** — conclusions verified against the subject's real life events; re-derive if unmatched.
- **Semantic versioning** — VERSION + CHANGELOG, each version records changes and evaluation data.
- **Honest evaluation** — independent auto-grading, answer isolation, public data, no inflated scores.

## References

This project's design references the following open-source projects:

- [weizeW/mingli-skills](https://github.com/weizeW/mingli-skills) — Four-dimensional cross-validation framework
- [jinchenma94/bazi-skill](https://github.com/jinchenma94/bazi-skill) — Classical summary design, historical event calibration
- [dzcmemory-web/bazi-ziwei-skill](https://github.com/dzcmemory-web/bazi-ziwei-skill) — Bazi+ZiWei cross-validation mode
- [shizhilya/yuan](https://github.com/shizhilya/yuan) — Conclusion-first output design
- [hhszzzz/taibu](https://github.com/hhszzzz/taibu) — Agent-friendly design
- [SylarLong/iztro](https://github.com/SylarLong/iztro) — Ziwei Doushu charting engine
- [ai-freer/fortune-skill](https://github.com/ai-freer/fortune-skill) — Three-layer computation architecture
- [yanouyuan-bit/bazi-roundtable](https://github.com/yanouyuan-bit/bazi-roundtable) — Multi-school review and conclusion strength labeling
- [DestinyLinker/MingLi-Bench](https://github.com/DestinyLinker/MingLi-Bench) — Metaphysics benchmark reference
- [2021291696/high-confidence-mingli-skill](https://github.com/2021291696/high-confidence-mingli-skill) — Confidence system, personality profile inference

## Links

- Website: [liki.hk](https://liki.hk)
- GitHub: [ml8s/liki](https://github.com/ml8s/liki)

## Disclaimer

Liki provides Chinese Astrology analysis from a traditional cultural perspective for research and reference only. Its conclusions **do not constitute** medical diagnosis, legal advice, financial forecasts, or major life decisions. Please maintain a rational and proactive outlook.

---
