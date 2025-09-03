const origWrite = process.stdout.write.bind(process.stdout) as any;
let bytes = 0;
let writes = 0;
let pendingSeq: string | null = null;

export function wrapStdout() {
  (process.stdout.write as any) = (chunk: any, enc?: any, cb?: any) => {
    const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk, enc);
    bytes += buf.length;
    writes++;
    emitMetric("write", { n: buf.length, ts: Date.now() });
    const ret = origWrite(buf, enc, cb);
    if (pendingSeq) {
      emitMetric("flush", { seq: pendingSeq, ts: Date.now() });
      pendingSeq = null;
    }
    return ret;
  };
}

export function summary(extra: Record<string, any> = {}) {
  emitMetric("summary", { bytes, writes, ...extra });
}

export function emitMetric(kind: string, data: Record<string, any>) {
  const payload = JSON.stringify({ kind, ...data, ts: Date.now() });
  origWrite("\n[METRIC]" + payload + "\n");
}

// Set the sequence to be flushed on the next write
export function setPendingFlushSeq(seq: string) {
  pendingSeq = seq;
}
