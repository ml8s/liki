# Liki user guide

This guide keeps the full usage instructions. The README is only the quick entry point; this document explains flows, boundaries, and common questions.

## Before you start

Different domains need different inputs. In all cases, provide only what the current question requires and wait for follow-up questions when data is missing.

| Domain | Prepare |
|---|---|
| Destiny | Birth date, precise time if available, birth city, gender |
| Naming | Surname, gender, preferences, and birth data when the exact time is known |
| Divination | One concrete event, decision, or timing question |
| Feng shui | Facing, period date, and birth year plus gender for Bazhai |

## First question

Send context and the question in one message:

> Help me review marriage. Female, born 1992-03-15 14:30 in Guangzhou.

Missing information is acceptable. If you only know “morning” or do not know the hour, the Skill will ask follow-up questions or enter time calibration.

## Conversation style

The Skill works like a professional consultation: one topic at a time, conclusion first, evidence second.

| Goal | Ask |
|---|---|
| Evidence | `Why?` / `What is the basis?` |
| A specific year | `What about 2026?` / `What about the next three years?` |
| Another topic | `What about wealth?` / `What about health?` |
| Another person | `Are we compatible?` |
| Full report | `Please create my full life reading.` |

The same chart is reused across follow-up questions; you do not need to resend birth data.

## Destiny and Ziwei

The destiny domain calculates both Bazi and Ziwei, then reads assertions for the requested topic.

Topics include:

- Marriage: timing, quality, relationship dynamics.
- Career: direction, entrepreneurship, promotion, and difficult years.
- Wealth: income type, gains, losses, and financial windows.
- Health: weak organ tendencies and years needing care.
- Study: academic ability, examinations, and advancement.
- Personality, parents, siblings, children, relocation, and palace topics.

Output follows one structure:

1. A direct conclusion, or an explicit statement that evidence is insufficient.
2. Key destiny facts and assertion references.
3. Interpretation and practical suggestions.
4. Timing windows as conditional tendencies, not guarantees.

## Naming

The naming flow is:

1. Confirm single-character or two-character given name, style, avoidances, required characters, and duplication preference.
2. Use Bazi yongshen when the complete birth time is available.
3. Select candidate characters from the engine character pool.
4. Compose names and validate characters, phonetics, and Wu Xing (五行).
5. Present candidates, strengths, trade-offs, and representative rejected names.

Example:

> Baby naming. Male, born 2024-06-10 in Guangzhou, family name Chen.

For a foreigner’s Chinese name, the Skill first confirms a Chinese surname from phonetic candidates. If there is no reliable phonetic match, it explains the fallback instead of inventing a transliteration.

Without a complete birth time, the Skill does not default to noon. It uses an expected Wu Xing (五行) strategy and explicitly reports that yongshen was not evaluated.

## Divination

Divination selects one method by question type. It does not automatically combine methods.

| Goal | Default method | Example |
|---|---|---|
| Outcome or timing | Liuyao | `Can this project succeed?` |
| Strategy, direction, or action timing | QiMen | `Should I sign now?` |
| Date selection | Huangli | `Which day is good for moving?` |

If outcome and strategy are both requested, the Skill asks you to choose the main question first. If you explicitly ask for dual-method review, Liuyao reviews the outcome and QiMen reviews strategy independently.

Ordinary QiMen questions do not require method parameters. The default is hour scope, rotating plate, and chaibu hour determination. Advanced schools are used only when explicitly requested.

| Preferred method | Ask |
|---|---|
| Zhirun determination | `Use the zhirun chart for now.` |
| Luo Shu flying plate | `Use the Luo Shu flying plate for this matter.` |
| Ten-minute Kejia | `Use the ten-minute Kejia method.` |
| Twelve-minute ten-division | `Use the twelve-minute ten-division method.` |
| Jinhan Yujing | `Use Jinhan Yujing for today.` |

Each casting keeps an immutable snapshot and evidence references; follow-up questions reuse the same snapshot instead of recasting.

## Feng shui

Feng shui covers two orthogonal systems:

| System | Purpose |
|---|---|
| Bazhai | Ming gua, four auspicious and inauspicious directions, door / master / stove |
| Xuankong | period, sitting and facing stars, annual flying stars |

Bazhai needs birth year and gender, not the birth hour. Xuankong needs the period date and building facing, not the occupant’s birth date.

Balcony, main door, and primary daylight facing can differ. The Skill confirms which one defines the sitting before calculating. The sitting and facing mountains must be 180 degrees apart.

## FAQ

### What if I don't know the exact birth hour?

Destiny analysis asks follow-up questions or uses real events for calibration. Provide 2-3 candidate hours and 3-5 past events with years when possible; the Skill checks candidates and reports confidence. When evidence is insufficient, it reports insufficient evidence. Naming does not default the hour; it switches to an expected Wu Xing (五行) strategy.

Babies and teenagers skip calibration; the provided hour is used as given.

### Does it need internet?

By default, yes. The JSON-RPC engine performs calendar and chart calculations. Advanced users can run a private engine and set `LIKI_RPC_URL`.

### Is my birth data stored?

No. The Skill does not store birth data outside the conversation and does not ask for your real name. Birth data remains in the current conversation context.

### How should I interpret results?

Conclusions are conditional interpretations from a traditional cultural perspective. For health, legal, financial, safety, or major decisions, consult qualified professionals.

### How do I update?

The Skill checks its version at startup. When prompted, run:

```bash
npx skills add ml8s/liki -y
```

### How do I upgrade a self-hosted engine?

Keep the Skill `VERSION.txt` compatible with `info.version` returned by engine discovery. If the engine is older than required, the tool layer fails closed instead of calling old RPCs.

## Output principles

- Give a direct conclusion first, or state clearly that evidence is insufficient.
- Attach engine facts, assertion IDs, factors, or RPC evidence.
- Report unavailable fields as unavailable.
- Do not present timing windows as guarantees.
- Add professional boundaries for health, legal, financial, safety, or major life topics.
