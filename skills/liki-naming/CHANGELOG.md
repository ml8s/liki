# Changelog

## 2026.09.02.1

- Remove Sancai-Wuge naming rules.
- Replace the legacy naming build endpoint with `qiming.compose`.
- Return candidate character pools from `qiming.pick`.
- Resolve all character facts from the qiming database during composition and evaluation.
- Return only generated names from `qiming.compose`; `qiming.check` provides character facts.
- Infer single/double names from the presence of `second`; `qiming.check` accepts `given_names`.
- Resolve radical-based element inference while generating the runtime character table.
- Normalize absent traditional and radical values.
- Project the source character table into a six-field runtime table without the unused traditional-form field.
- Project Kangxi source data into a three-field runtime table.

## 2026.09.01.7

- Rename the public field to `kangxi_stroke`; Wuge calculations consume Kangxi strokes.
- Keep Unihan as the source-table naming boundary.

## 2026.09.01.6

- Rename generated source tables to `unihan_*` and expose the Wuge stroke field for Wuge calculations.

## 2026.09.01.5

- Wuge true path now uses generated Kangxi strokes for surnames and candidate characters.
- `qiming.char` returns modern `stroke`, `kangxi_stroke`, and the form used by Wuge.
