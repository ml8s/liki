# Liki context map

This map states the ownership boundary behind [SYSTEM.md](SYSTEM.md). A fact is
changed in exactly one context; other contexts consume its published projection.

| Context | Owns | Does not own | Main projection |
| --- | --- | --- | --- |
| Engine | astronomical time, calendars, deterministic chart facts, casting source facts | judgment, payment, user identity | MCP chart/casting facts |
| Counsel | domain factors, assertion rules, judgment, yongshen, naming selection | raw astronomy, persistence of charts, paywall | MCP judgment and deterministic reports |
| Product MCP | public tool surface, exposure policy, A2M paywall and fulfillment | deterministic computation and judgment rules | `hk.liki/liki` tool catalog and payment headers |
| Web BFF | user session, birth profile persistence, commerce, chat orchestration | deterministic facts and skill methodology | REST/browser data and Agent input |
| Agent Runtime | skill assembly, deployment contract, host adapter execution | domain rules, product tool surface | `liki-expert` deployment |

## Collaboration rules

1. Engine facts are deterministic inputs; Counsel may add judgment evidence but
   must not replace or reinterpret the engine's canonical time and chart facts.
2. Product MCP forwards tool contracts unchanged. Its only product decisions are
   exposure and paywall.
3. Web BFF is the only owner of persisted user identity and birth profile state.
4. Agent Runtime consumes a skill/deployment projection; it never becomes a
   second source of domain methodology.
5. Cross-context identifiers remain explicit: git SHA is the build identity,
   `out_trade_no` is the payment order identity, and profile ID is owned by Web
   BFF.

## Related language

Canonical domain terms and compatibility names are governed by
[DOMAIN_LANGUAGE.md](DOMAIN_LANGUAGE.md); runtime names are governed by
[DOMAIN_NAMING.md](DOMAIN_NAMING.md).
