import React, { useEffect, useMemo, useRef, useState } from "react";
import { render, Box, Text, useInput } from "ink";
import { wrapStdout, summary, emitMetric } from "./instrument/ioWrap";

wrapStdout();

const RATE = parseInt(process.env.RATE || "10000", 10);

function App() {
  const seq = useRef(0);
  const [buf, setBuf] = useState<string[]>([]);
  const [dropped, setDropped] = useState(0);
  const paused = useRef(false);

  useEffect(() => {
    const id = setInterval(() => {
      if (paused.current) return;
      const n = Math.max(1, Math.floor(RATE / 60));
      setBuf((prev) => {
        const next = prev.slice();
        for (let i = 0; i < n; i++) {
          const ts = new Date().toISOString();
          const lvl = ["DEBUG", "INFO", "WARN", "ERROR"][
            Math.floor(Math.random() * 4)
          ];
          next.push(`${ts} ${lvl} msg=hello`);
        }
        const maxBuf = 20000;
        let drop = 0;
        while (next.length > maxBuf) {
          next.shift();
          drop++;
        }
        if (drop) setDropped((d) => d + drop);
        return next;
      });
    }, 16);
    return () => clearInterval(id);
  }, []);

  useEffect(() => () => summary({ dropped }), [dropped]);

  useInput((input, key) => {
    seq.current += 1;
    const s = String(seq.current);
    emitMetric("input", { key: key.return ? "enter" : input, seq: s });
    if ((key.ctrl && input === "c") || input === "q") {
      summary({ dropped });
      process.exit(0);
    }
    if (input === " ") paused.current = !paused.current;
    // State change example
    if (input === " ")
      emitMetric("state", { field: "paused", value: paused.current, seq: s });
  });

  const lines = useMemo(() => {
    const start = Math.max(0, buf.length - 200);
    return buf.slice(start);
  }, [buf]);

  return (
    <Box flexDirection="column">
      <Text>
        Logstream | rate={RATE} dropped={dropped}
      </Text>
      {lines.map((l, i) => (
        <Text key={i}>{l}</Text>
      ))}
    </Box>
  );
}

const instance = render(<App />);
// After each render frame, schedule a flush mark tied to the latest seq
// This is approximate but good enough to correlate input->flush at frame granularity
setInterval(() => {
  emitMetric("flush", { seq: String(Date.now()) });
}, 50);
