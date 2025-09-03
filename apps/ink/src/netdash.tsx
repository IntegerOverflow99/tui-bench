import React, { useEffect, useMemo, useRef, useState } from "react";
import { render, Box, Text, useInput } from "ink";
import { wrapStdout, summary, emitMetric } from "./instrument/ioWrap";

wrapStdout();

const URL = process.env.URL || "https://example.com";

function bars(vals: number[]) {
  const glyphs = ["▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"];
  const max = Math.max(1, ...vals);
  return vals
    .map(
      (v) =>
        glyphs[
          Math.min(
            glyphs.length - 1,
            Math.max(0, Math.round((v / max) * (glyphs.length - 1)))
          )
        ]
    )
    .join("");
}

function App() {
  const [lat, setLat] = useState<number[]>([]);
  const [errs, setErrs] = useState(0);
  const seq = useRef(0);

  useEffect(() => {
    let alive = true;
    const tick = async () => {
      const t0 = performance.now();
      try {
        const ctrl = new AbortController();
        const id = setTimeout(() => ctrl.abort(), 5000);
        await fetch(URL, { method: "HEAD", signal: ctrl.signal });
        clearTimeout(id);
        const ms = performance.now() - t0;
        if (!alive) return;
        setLat((prev) => {
          const next = [...prev, ms];
          if (next.length > 80) next.shift();
          return next;
        });
      } catch {
        if (!alive) return;
        setErrs((e) => e + 1);
      } finally {
        setTimeout(tick, 1000);
      }
    };
    tick();
    return () => {
      alive = false;
    };
  }, []);

  useInput((input, key) => {
    seq.current += 1;
    const s = String(seq.current);
    emitMetric("input", { key: key.return ? "enter" : input, seq: s });
    if ((key.ctrl && input === "c") || input === "q") {
      summary({ errors: errs });
      process.exit(0);
    }
  });

  const spark = useMemo(() => bars(lat), [lat]);
  const avg = useMemo(
    () => (lat.length ? lat.reduce((a, b) => a + b, 0) / lat.length : 0),
    [lat]
  );
  const max = useMemo(() => (lat.length ? Math.max(...lat) : 0), [lat]);

  return (
    <Box flexDirection="column">
      <Text>
        Netdash {URL} errors={errs}
      </Text>
      <Text>latency(ms)</Text>
      <Text>{spark}</Text>
      <Text>
        n={lat.length} avg={avg.toFixed(1)} max={max.toFixed(1)}
      </Text>
    </Box>
  );
}

const inst = render(<App />);
setInterval(() => emitMetric("flush", { seq: String(Date.now()) }), 50);
