# TUI Bench

A small cross-stack benchmark to compare terminal UIs built with:

- Go: Charm's Bubble Tea/Bubbles
- Node/React: Ink

It includes:

- A benchmark harness (Go) that launches apps in a PTY, drives scripted inputs, and collects metrics from app-emitted lines.
- Two scenarios implemented in each stack:
  - logstream: high-throughput log line generation with backpressure.
  - netdash: simple network dashboard that periodically pings a URL and shows latency.

## Structure

- bench/ Go benchmark harness
- apps/
  - bubbletea/ Go apps (Bubble Tea)
  - ink/ Node/React apps (Ink + TypeScript)

## Quick start (macOS)

Prereqs:

- Go 1.21+
- Node.js 18+ (or 20+ recommended)

Install Go deps and build harness:

```sh
cd bench
go mod tidy
go build ./cmd/benchctl
```

Install Ink app deps:

```sh
cd ../apps/ink
pnpm install
```

Build and run Bubble Tea logstream directly (optional sanity check):

```sh
cd ../bubbletea
go mod tidy
go run ./cmd/logstream
```

Run the bench against Bubble Tea logstream (120x40 PTY, 10k lines/sec, 10s duration):

```sh
cd ../../bench
./benchctl -app "./../apps/bubbletea/logstream" -scenario logstream -cols 120 -rows 40 -rate 10000 -dur 10s
```

Run the bench against Ink logstream:

```sh
cd ../../bench
./benchctl -app "node ../apps/ink/src/logstream.tsx" -scenario logstream -cols 120 -rows 40 -rate 10000 -dur 10s
```

Notes:

- The harness expects apps to emit metrics lines prefixed with `[METRIC]` followed by a JSON object.
- For the Ink TypeScript entry points, the package uses `tsx` for zero-config execution; you can also run via `npm run logstream`.

## Scenarios

- logstream: Generate many lines/sec. Space toggles pause. Ctrl+C to quit.
- netdash: Periodically HEAD a URL (default [https://example.com](https://example.com)) and render a tiny ASCII sparkline of recent latencies.

## Output

The harness prints a JSON report summarizing metrics observed during the measurement window (after warmup):

- bytesPerSec, writesPerSec (from app stdout wrapper)
- cpuPct avg/max, rssMB avg/max (sampled via ps on macOS)
- app counters like dropped lines

More metrics can be added easily by emitting additional `[METRIC]{...}` events.

See BUILD.md for step-by-step build/run commands, and TODO.md for next steps.
