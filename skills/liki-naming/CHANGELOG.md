# Changelog

## 2026.09.01.7

- Rename the public field to `kangxi_stroke`; Wuge calculations consume Kangxi strokes.
- Keep Unihan as the source-table naming boundary.

## 2026.09.01.6

- Rename generated source tables to `unihan_*` and expose the Wuge stroke field for Wuge calculations.

## 2026.09.01.5

- Wuge true path now uses generated Kangxi strokes for surnames and candidate characters.
- `qiming.char` returns modern `stroke`, `kangxi_stroke`, and the form used by Wuge.
