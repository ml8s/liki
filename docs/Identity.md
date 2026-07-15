# liki.hk Identity

Version: 9.0 (Canonical)
Status: Active
Last Updated: 2026-07-15

---

# Identity

The public identity of the project is:

> **liki.hk**

All public-facing assets should consistently use this identity.

Including:

- Website
- GitHub
- README
- Documentation
- LLMS.txt
- Open Graph
- Schema.org
- AI Search
- Social Media

---

# Positioning

The canonical positioning of the project is:

> **AI Skills for Chinese Metaphysics**

This positioning is the single source of truth.

It should remain identical across all public surfaces.

---

# Canonical Definition

**liki.hk provides AI skills for Chinese Metaphysics.**

Current skills include:

- Mingli
- Qiming
- Wengua
- Fengshui

Additional skills may be introduced in the future without changing the project's identity or positioning.

---

# Terminology

## AI Skills

User-facing capabilities for specific Chinese Metaphysics tasks.

Examples include:

- Mingli
- Qiming
- Wengua
- Fengshui

---

## Metaphysics Engine

The stateless deterministic computation engine behind AI Skills.

It is responsible only for deterministic calculations.

---

## LLM

The reasoning layer.

It is responsible for understanding user intent, selecting skills, interpreting calculation results and generating responses.

---

# Product Model

```
liki.hk
        │
        ▼
AI Skills
        ├── Mingli
        ├── Qiming
        ├── Wengua
        ├── Fengshui
        └── ...
```

Users interact with AI Skills.

AI Skills are the public product.

---

# Architecture

liki.hk separates deterministic computation from AI reasoning.

```
                User
                  │
                  ▼
             AI Skills
                  │
                  ▼
        Metaphysics Engine
                  │
                  ▼
         Deterministic Results
                  │
                  ▼
         LLM Interpretation
                  │
                  ▼
          Structured Response
```

## Metaphysics Engine

The Metaphysics Engine is an independent component of the liki ecosystem.

It provides stateless deterministic computation for Chinese Metaphysics.

Its responsibilities include:

- Astronomical calculations
- Calendar calculations
- True Solar Time
- BaZi chart generation
- Ziwei chart generation
- Feng Shui calculations
- Almanac calculations
- Other rule-based calculations

The engine:

- Is stateless
- Is deterministic
- Produces reproducible results
- Performs calculations only
- Does not perform reasoning or interpretation
- Can be independently developed and deployed

AI Skills interact with the engine through well-defined interfaces, while the LLM is responsible for reasoning and interpretation.

## LLM

The LLM is responsible for:

- Understanding user intent
- Collecting required parameters
- Selecting the appropriate AI Skill
- Calling the Metaphysics Engine
- Interpreting calculation results
- Generating structured responses

This architecture ensures deterministic computation remains accurate, reproducible and independent of LLM hallucinations.

---

# Presentation

Every public surface should present the project consistently.

Identity

```
liki.hk
```

Positioning

```
AI Skills for Chinese Metaphysics
```

Definition

```
liki.hk provides AI skills for Chinese Metaphysics.
```

README, Website, Documentation, LLMS.txt, Schema.org and Social Media should all use the same identity, positioning and definition.

---

# Repository

Repository

```
ml8s/liki
```

Installation

```bash
npx skills add ml8s/liki
```

The repository name and installation command are technical identifiers.

The public identity remains:

```
liki.hk
```

---

# Design Principles

## One Identity

```
liki.hk
```

---

## One Positioning

```
AI Skills for Chinese Metaphysics
```

---

## One Definition

```
liki.hk provides AI skills for Chinese Metaphysics.
```

---

## One Product

Users interact with AI Skills.

---

## One Engine

AI Skills are powered by the Metaphysics Engine.

---

## One Architecture

Separate deterministic computation from AI reasoning.

---

# Vision

Advance the understanding and application of Chinese Metaphysics through trustworthy AI.

Every interpretation should be built upon deterministic and reproducible computation.
