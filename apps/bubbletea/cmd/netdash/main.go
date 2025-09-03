package main

import (
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/viewport"

    "apps/bubbletea/internal/instrument"
)

type model struct {
    vp      viewport.Model
    url     string
    lat     []time.Duration
    maxN    int
    errc    int
    period  time.Duration
    seq     int
}

type tick struct{}
type sample struct{ d time.Duration; err error }

func (m model) Init() tea.Cmd {
    return tea.Batch(poll(m.url), tickCmd())
}

func tickCmd() tea.Cmd { return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg { return tick{} }) }

func poll(u string) tea.Cmd {
    return func() tea.Msg {
        start := time.Now()
        req, _ := http.NewRequest("HEAD", u, nil)
        req.Header.Set("User-Agent", "tui-bench/0.1")
        client := &http.Client{ Timeout: 5 * time.Second }
        _, err := client.Do(req)
        d := time.Since(start)
        return sample{d: d, err: err}
    }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        m.seq++
        seq := fmt.Sprintf("%d", m.seq)
        instrument.Emit("input", map[string]any{"key": msg.String(), "seq": seq})
        if msg.String() == "ctrl+c" || msg.String() == "q" { return m, tea.Quit }
    case tick:
        // render
        content := m.render()
        m.vp.SetContent(content)
        return m, tickCmd()
    case sample:
        if msg.err != nil { m.errc++ } else {
            m.lat = append(m.lat, msg.d)
            if len(m.lat) > m.maxN { m.lat = m.lat[1:] }
        }
        return m, poll(m.url)
    }
    return m, nil
}

func (m model) View() string { return m.vp.View() }

func (m model) render() string {
    // sparkline over last N latencies in ms
    if len(m.lat) == 0 {
        return fmt.Sprintf("Netdash %s\n(no data yet)\nerrors=%d", m.url, m.errc)
    }
    ms := make([]float64, len(m.lat))
    max := 1.0
    for i, d := range m.lat { v := float64(d.Milliseconds()); ms[i] = v; if v > max { max = v } }
    bars := []rune("▁▂▃▄▅▆▇█")
    var b strings.Builder
    b.WriteString(fmt.Sprintf("Netdash %s period=%s errors=%d\n", m.url, m.period, m.errc))
    b.WriteString("latency(ms)\n")
    for _, v := range ms {
        idx := int((v/max)*float64(len(bars)-1) + 0.5)
        if idx < 0 { idx = 0 }
        if idx >= len(bars) { idx = len(bars)-1 }
        b.WriteRune(bars[idx])
    }
    b.WriteString("\n")
    if n := len(ms); n > 0 {
        // simple stats
        sum := 0.0
        for _, v := range ms { sum += v }
        avg := sum / float64(n)
        b.WriteString(fmt.Sprintf("n=%d avg=%.1f max=%.1f\n", n, avg, max))
    }
    return b.String()
}

func main() {
    url := os.Getenv("URL")
    if url == "" { url = "https://example.com" }
    periodMs, _ := strconv.Atoi(os.Getenv("PERIOD_MS"))
    if periodMs == 0 { periodMs = 1000 }
    m := model{
        vp:     viewport.Model{Width: 120, Height: 40},
        url:    url,
        maxN:   80,
        period: time.Duration(periodMs) * time.Millisecond,
    }

    cw := &instrument.CountingWriter{W: os.Stdout}
    p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(cw))
    if _, err := p.Run(); err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }
    instrument.Emit("summary", map[string]any{"bytes": cw.Bytes, "writes": cw.Writes, "errors": m.errc})
}
