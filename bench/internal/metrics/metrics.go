package metrics

import (
	"sync"
	"time"
	"sort"
)

type Report struct {
	Bytes       uint64            `json:"bytes,omitempty"`
	Writes      uint64            `json:"writes,omitempty"`
	Counters    map[string]uint64 `json:"counters,omitempty"`
	CPU         CPUReport         `json:"cpu,omitempty"`
	Memory      MemReport         `json:"memory,omitempty"`
	WindowStart time.Time         `json:"windowStart"`
	WindowEnd   time.Time         `json:"windowEnd"`
	Latency     struct {
		InputToState map[string]float64 `json:"inputToStateMs,omitempty"`
		StateToFlush map[string]float64 `json:"stateToFlushMs,omitempty"`
		InputToFlush map[string]float64 `json:"inputToFlushMs,omitempty"`
	} `json:"latency,omitempty"`
}

type CPUReport struct {
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
}

type MemReport struct {
	AvgMB float64 `json:"avgMB"`
	MaxMB float64 `json:"maxMB"`
}

type Aggregator struct {
	mu       sync.Mutex
	measuring bool
	start    time.Time
	end      time.Time
	bytes    uint64
	writes   uint64
	counters map[string]uint64
	// latency matching by seq
	in   map[string]time.Time
	st   map[string]time.Time
	fl   map[string]time.Time
}

func NewAggregator() *Aggregator {
	return &Aggregator{counters: map[string]uint64{}, in: map[string]time.Time{}, st: map[string]time.Time{}, fl: map[string]time.Time{}}
}

func (a *Aggregator) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.measuring = true
	a.start = time.Now()
}

func (a *Aggregator) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.measuring = false
	a.end = time.Now()
}

func (a *Aggregator) Add(m map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.measuring {
		return
	}
	if kind, _ := m["kind"].(string); kind == "write" {
		if n, ok := m["n"].(float64); ok {
			a.bytes += uint64(n)
		}
		a.writes++
	}
	if kind, _ := m["kind"].(string); kind == "counter" {
		if name, ok := m["name"].(string); ok {
			if v, ok := m["value"].(float64); ok {
				a.counters[name] += uint64(v)
			}
		}
	}
	// latency triplet
	if seq, _ := m["seq"].(string); seq != "" {
		ts := time.Now()
		// Prefer event-provided timestamp
		if tss, ok := m["ts"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339Nano, tss); err == nil {
				ts = parsed
			}
		} else if tsm, ok := m["ts"].(float64); ok {
			// milliseconds since epoch
			sec := int64(tsm) / 1000
			nsec := (int64(tsm) % 1000) * int64(time.Millisecond)
			ts = time.Unix(sec, nsec)
		}
		switch m["kind"] {
		case "input":
			a.in[seq] = ts
		case "state":
			a.st[seq] = ts
		case "flush":
			a.fl[seq] = ts
		}
	}
}

func (a *Aggregator) Report() Report {
	a.mu.Lock()
	defer a.mu.Unlock()
	rep := Report{
		Bytes:       a.bytes,
		Writes:      a.writes,
		Counters:    a.counters,
		WindowStart: a.start,
		WindowEnd:   a.end,
	}
	// Compute simple averages for latency (ms); quantiles can be added later
	itos := []float64{}
	stfl := []float64{}
	itfl := []float64{}
	for seq, t0 := range a.in {
		if t1, ok := a.st[seq]; ok {
			itos = append(itos, float64(t1.Sub(t0).Milliseconds()))
		}
		if t2, ok := a.fl[seq]; ok {
			itfl = append(itfl, float64(t2.Sub(t0).Milliseconds()))
		}
	}
	for seq, t1 := range a.st {
		if t2, ok := a.fl[seq]; ok {
			stfl = append(stfl, float64(t2.Sub(t1).Milliseconds()))
		}
	}
	rep.Latency.InputToState = summarize(itos)
	rep.Latency.StateToFlush = summarize(stfl)
	rep.Latency.InputToFlush = summarize(itfl)
	return rep
}

func summarize(vals []float64) map[string]float64 {
	if len(vals) == 0 { return nil }
	// naive p50/p95/p99 with sort
	cp := append([]float64(nil), vals...)
	sort.Float64s(cp)
	n := float64(len(cp))
	pick := func(p float64) float64 {
		if len(cp) == 0 { return 0 }
		idx := int(p*n) - 1
		if idx < 0 { idx = 0 }
		if idx >= len(cp) { idx = len(cp)-1 }
		return cp[idx]
	}
	return map[string]float64{
		"p50": pick(0.50),
		"p95": pick(0.95),
		"p99": pick(0.99),
	}
}
