# TODO / Next Steps

1. Latency fidelity and metrics

- Bubble Tea & Ink: ensure input/state/flush seqs are tightly correlated with the rendered frame; consider hooking precise write completion instead of interval-based flush in Ink.
- Aggregator: expand to compute p50/p95/p99 for frame time and input latency using explicit timestamps from apps.

1. Input driver and scenarios

- Add flags to benchctl to load scripts from JSON/YAML and to choose scenario-specific scripts (logstream vs netdash).
- Add more keys: page up/down, filter typing, search, resize injections.

1. Resource metrics

- Go: emit runtime/metrics snapshots, goroutine count, GC pauses, allocs/sec.
- Node: add perf_hooks.monitorEventLoopDelay, process.memoryUsage, V8 heap stats.

1. Throughput/backpressure

- Wrap stdout with bounded channels to simulate backpressure; track dropped frames/writes.
- Record bytes/sec and writes/sec by interval, not just totals.

1. Reporting

- Add bench/internal/report to aggregate quantiles and export CSV/JSON; include run config and run ID.
- Optional small charts or markdown summary.

1. Scenario parity and UX

- Match visible layout between stacks; add theme toggle; verify resize behavior.

1. CI and reproducibility

- Add scripts to run multiple iterations with warmup/measure/cooldown; save artifacts per run.

1. Netdash improvements

- Add jitter to schedule; method/path options; status histogram; error rate.

1. Logstream improvements

- Add filter/search; configurable buffer caps and drop policy; visibility-only rendering toggle.
