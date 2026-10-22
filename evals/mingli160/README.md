# Mingli160 calibration/evaluation set

This dataset holds 32 birth records and their 160-question grouping. It informed
the natal assertion table during development, so it is **not held out** and must
not be reported as independent accuracy.

Run the manual report with:

```bash
make eval-mingli160
```

The report measures assertion coverage and zero-hit cases. It does not grade
answers and is deliberately outside `make test`, `make gate`, and
`make release-gate`. If a calibration change modifies source or skill assets,
run `make gate` afterward. A real accuracy claim needs a frozen assertion
version and a second, tuning-untouched question set.
