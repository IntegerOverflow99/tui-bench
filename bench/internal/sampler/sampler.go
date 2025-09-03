package sampler

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"bench/internal/metrics"
)

type Sampler struct {
	pid int
	mu  sync.Mutex
	cpu []float64
	rss []float64
}

func New(pid int) *Sampler { return &Sampler{pid: pid} }

func (s *Sampler) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cpu, rss := psSample(s.pid)
			s.mu.Lock()
			s.cpu = append(s.cpu, cpu)
			s.rss = append(s.rss, rss)
			s.mu.Unlock()
		}
	}
}

func (s *Sampler) ReportCPU() (rep metrics.CPUReport) {
	s.mu.Lock(); defer s.mu.Unlock()
	var sum, max float64
	for _, v := range s.cpu { sum += v; if v > max { max = v } }
	if len(s.cpu) > 0 { rep.Avg = sum / float64(len(s.cpu)); rep.Max = max }
	return
}

func (s *Sampler) ReportMem() (rep metrics.MemReport) {
	s.mu.Lock(); defer s.mu.Unlock()
	var sum, max float64
	for _, v := range s.rss { sum += v; if v > max { max = v } }
	if len(s.rss) > 0 { rep.AvgMB = sum / float64(len(s.rss)); rep.MaxMB = max }
	return
}

// psSample uses `ps -o %cpu=,rss=` to sample CPU% and RSS KB on macOS/Linux.
func psSample(pid int) (cpu float64, rssMB float64) {
	cmd := exec.Command("ps", "-o", "%cpu=,rss=", "-p", strconv.Itoa(pid))
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	fields := strings.Fields(out.String())
	if len(fields) >= 2 {
		cpu, _ = strconv.ParseFloat(fields[0], 64)
		rssKB, _ := strconv.ParseFloat(fields[1], 64)
		rssMB = rssKB / 1024.0
	}
	return
}
