const origWrite = process.stdout.write.bind(process.stdout) as any;
let bytes = 0;
let writes = 0;

export function wrapStdout() {
  (process.stdout.write as any) = (chunk: any, enc?: any, cb?: any) => {
    const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk, enc);
    bytes += buf.length;
    writes++;
    emitMetric("write", { n: buf.length, ts: Date.now() });
    return origWrite(buf, enc, cb);
  };
}

export function summary(extra: Record<string, any> = {}) {
  emitMetric("summary", { bytes, writes, ...extra });
}

export function emitMetric(kind: string, data: Record<string, any>) {
  const payload = JSON.stringify({ kind, ...data, ts: Date.now() });
  origWrite("\n[METRIC]" + payload + "\n");
}
