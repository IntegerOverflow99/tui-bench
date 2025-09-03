# Build and Run

These instructions assume macOS with Go 1.21+ and Node 18+.

## Go harness (bench)

- Install deps and build:

```sh
cd bench
go mod tidy
# optional build: go build ./cmd/benchctl
```

## Bubble Tea apps (Go)

```sh
cd apps/bubbletea
go mod tidy
# optional build: go build ./cmd/logstream
# optional build: go build ./cmd/netdash
```

## Ink apps (Node/TypeScript)

```sh
cd apps/ink
pnpm install
# dev runs (no build needed):
pnpm run dev:logstream
pnpm run dev:netdash
```

## Run harness (examples)

- Bubble Tea logstream (10s, 10k lines/s, 120x40):

```sh
cd bench
./benchctl -app "../apps/bubbletea/cmd/logstream/logstream" -scenario logstream -cols 120 -rows 40 -rate 10000 -dur 10s
```

- Ink logstream (driven via node):

```sh
cd bench
./benchctl -app "node ../apps/ink/src/logstream.tsx" -scenario logstream -cols 120 -rows 40 -rate 10000 -dur 10s
```

- Bubble Tea netdash:

```sh
cd bench
URL=https://example.com ./benchctl -app "../apps/bubbletea/cmd/netdash/netdash" -scenario netdash -dur 10s
```

- Ink netdash:

```sh
cd bench
URL=https://example.com ./benchctl -app "node ../apps/ink/src/netdash.tsx" -scenario netdash -dur 10s
```

Note: The harness parses lines prefixed with `[METRIC]` for metrics; other output is ignored.
