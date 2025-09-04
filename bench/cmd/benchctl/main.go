package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"bench/internal/metrics"
	"bench/internal/pty"
	"bench/internal/sampler"
	"bench/internal/driver"
)

func main() {
	var app string
	var scenario string
	var cols, rows int
	var dur, warmup time.Duration
	var rate int
	var cwd string
	var mirror bool
	flag.StringVar(&app, "app", "", "App command to run (binary or 'node script.js')")
	flag.StringVar(&scenario, "scenario", "logstream", "Scenario: logstream|netdash")
	flag.IntVar(&cols, "cols", 120, "PTY columns")
	flag.IntVar(&rows, "rows", 40, "PTY rows")
	flag.DurationVar(&dur, "dur", 10*time.Second, "Measurement duration")
	flag.DurationVar(&warmup, "warmup", 3*time.Second, "Warmup duration")
	flag.IntVar(&rate, "rate", 10000, "Log lines/sec or update rate")
	flag.StringVar(&cwd, "cwd", "", "Working directory for the app")
	flag.BoolVar(&mirror, "mirror", false, "Mirror the app's PTY output to stdout while parsing metrics")
	flag.Parse()

	if app == "" {
		log.Fatal("missing -app")
	}

	// Split command for exec if needed
	var cmd *exec.Cmd
	parts := strings.Fields(app)
	if len(parts) == 1 {
		cmd = exec.Command(parts[0])
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	// Env
	env := os.Environ()
	env = append(env, "SCENARIO="+scenario)
	env = append(env, fmt.Sprintf("RATE=%d", rate))
	env = append(env, "METRICS_JSON=1")
	cmd.Env = env
	if cwd != "" { cmd.Dir = cwd }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proc, tty, err := pty.SpawnWithSize(ctx, cmd, cols, rows)
	if err != nil {
		log.Fatalf("spawn: %v", err)
	}
	defer proc.Process.Kill()

	// Sampler
	s := sampler.New(proc.Process.Pid)
	go s.Run(ctx, 200*time.Millisecond)

	agg := metrics.NewAggregator()

	// Collect metrics lines prefixed with [METRIC]
	if mirror {
		pr, pw := io.Pipe()
		// Copy raw PTY to stdout and to a pipe for metrics parsing
		go func() {
			defer pw.Close()
			mw := io.MultiWriter(os.Stdout, pw)
			_, _ = io.Copy(mw, tty)
		}()
		go func() {
			scanner := bufio.NewScanner(pr)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "[METRIC]") {
					var m map[string]any
					if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "[METRIC]")), &m); err == nil {
						agg.Add(m)
					}
				}
			}
		}()
	} else {
		go func() {
			scanner := bufio.NewScanner(tty)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "[METRIC]") {
					var m map[string]any
					if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "[METRIC]")), &m); err == nil {
						agg.Add(m)
					}
				}
			}
		}()
	}

	// Drive a simple default input script for latency measurement
	go driver.Run(ctx, tty, driver.Script{Steps: []driver.Step{
		{At: 2 * time.Second, Keys: driver.KeysDown(5)},
		{At: 3 * time.Second, Keys: driver.KeysUp(3)},
		{At: 4 * time.Second, Keys: driver.KeysSpace()},
		{At: 6 * time.Second, Keys: driver.KeysSpace()},
		{At: 7 * time.Second, Keys: driver.KeysEnter()},
	}})

	time.Sleep(warmup)
	agg.Start()

	// For netdash scenario, inject 't' keypresses periodically during the measurement window
	if scenario == "netdash" {
		go func() {
			// Press immediately at measurement start, then every second
			_, _ = io.WriteString(tty, "t")
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_, _ = io.WriteString(tty, "t")
				}
			}
		}()
	}
	time.Sleep(dur)
	agg.Stop()
	cancel()
	// small grace period
	time.Sleep(300 * time.Millisecond)

	rep := agg.Report()
	rep.CPU = s.ReportCPU()
	rep.Memory = s.ReportMem()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rep)
}
