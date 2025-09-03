package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"apps/bubbletea/internal/instrument"
)

type model struct {
	vp      viewport.Model
	buf     []string
	maxBuf  int
	dropped int
	rate    int
	paused  bool
	seq     int
}

type tickMsg struct{}

type genMsg struct{ n int }

type togglePause struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), generate(m.rate))
}

func tick() tea.Cmd { return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} }) }

func generate(rate int) tea.Cmd {
	return func() tea.Msg {
		// lines per ~frame
		n := rate / 60
		if n < 1 { n = 1 }
		return genMsg{n: n}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.seq++
		seq := fmt.Sprintf("%d", m.seq)
		instrument.Emit("input", map[string]any{"key": msg.String(), "seq": seq})
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			instrument.Emit("state", map[string]any{"field": "paused", "value": m.paused, "seq": seq})
		}
	case tickMsg:
		// render tail
		start := 0
		if len(m.buf) > m.vp.Height {
			start = len(m.buf) - m.vp.Height
		}
		m.vp.SetContent(strings.Join(m.buf[start:], "\n"))
		// Mark flush; counting writer emits write, but we add logical flush for triplet
		instrument.Emit("flush", map[string]any{"seq": fmt.Sprintf("%d", m.seq)})
		return m, tick()
	case genMsg:
		if m.paused { return m, generate(m.rate) }
		for i := 0; i < msg.n; i++ {
			lvl := []string{"DEBUG","INFO","WARN","ERROR"}[rand.Intn(4)]
			line := fmt.Sprintf("%s %s msg=hello", time.Now().Format(time.RFC3339Nano), lvl)
			if len(m.buf) >= m.maxBuf {
				m.buf = m.buf[1:]
				m.dropped++
			}
			m.buf = append(m.buf, line)
		}
		return m, generate(m.rate)
	}
	return m, nil
}

func (m model) View() string { return m.vp.View() }

func main() {
	rate, _ := strconv.Atoi(os.Getenv("RATE"))
	if rate == 0 { rate = 10000 }
	m := model{
		vp:     viewport.Model{Width: 120, Height: 40},
		maxBuf: 20000,
		rate:   rate,
	}
	cw := &instrument.CountingWriter{W: os.Stdout}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(cw))
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	instrument.Emit("summary", map[string]any{"dropped": m.dropped, "bytes": cw.Bytes, "writes": cw.Writes})
}
