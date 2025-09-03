package instrument

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

var enable = os.Getenv("METRICS_JSON") == "1"

func Emit(kind string, m map[string]any) {
	if !enable { return }
	m["kind"] = kind
	m["ts"] = time.Now().Format(time.RFC3339Nano)
	b, _ := json.Marshal(m)
	fmt.Printf("\n[METRIC]%s\n", string(b))
}

type CountingWriter struct {
	W      io.Writer
	Bytes  uint64
	Writes uint64
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.W.Write(p)
	atomic.AddUint64(&cw.Bytes, uint64(n))
	atomic.AddUint64(&cw.Writes, 1)
	Emit("write", map[string]any{"n": n})
	return n, err
}
