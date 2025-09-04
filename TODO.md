# TODO / Next Steps

## Done recently

- Ink netdash wrapper script added at `apps/ink/bin/netdash`.
- Write-hooked flush implemented in both Bubble Tea and Ink; removed timer-based flush in Ink.
- Input/state/flush latency triplet with shared seq wired across both stacks.
- benchctl: added `-mirror` flag to tee PTY to stdout while parsing metrics.
- Metrics rename: `bytes` -> `bytesWritten` in reports.
- Netdash: added theme toggle on key `t` in both stacks.
- Harness input driver: auto-inject `t` presses during measurement for `netdash`.
- Aggregator: computes p50/p95/p99 for input→state, state→flush, and input→flush latencies.
- Sampler: CPU% and RSS MB sampling via `ps` integrated into report.

1. Latency fidelity and metrics

- Frame time measurement: instrument and report render durations per frame (extend apps + aggregator).
- Aggregator: already reports input latency quantiles; extend to frame-time quantiles.

1. Input driver and scenarios

- Add flags to benchctl to load scripts from JSON/YAML and to choose scenario-specific scripts (logstream vs netdash).
- Add more keys: page up/down, filter typing, search, resize injections.
- Netdash: auto `t` injection during measurement (DONE).

1. Resource metrics

- Go: emit runtime/metrics snapshots, goroutine count, GC pauses, allocs/sec.
- Node: add perf_hooks.monitorEventLoopDelay, process.memoryUsage, V8 heap stats.
- Already present: CPU% and RSS MB sampling in harness (DONE).

1. Throughput/backpressure

- Wrap stdout with bounded channels to simulate backpressure; track dropped frames/writes.
- Record bytes/sec and writes/sec by interval, not just totals.
- Already present: total bytesWritten and writes in report (DONE).

1. Reporting

- Add bench/internal/report to aggregate quantiles and export CSV/JSON; include run config and run ID.
- Optional small charts or markdown summary.
- Already present: JSON report printed by benchctl (basic), includes window times and resource metrics.

1. Scenario parity and UX

- Match visible layout between stacks; verify resize behavior.
- Theme toggle added in both netdash apps (DONE).

1. CI and reproducibility

- Add scripts to run multiple iterations with warmup/measure/cooldown; save artifacts per run.

1. Netdash improvements

- Add jitter to schedule; method/path options; status histogram; error rate.
- Errors are counted; consider adding rate and histogram visualizations.

1. Logstream improvements

- Add filter/search; configurable buffer caps and drop policy; visibility-only rendering toggle.
