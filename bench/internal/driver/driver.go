package driver

import (
    "context"
    "io"
    "strings"
    "time"
)

// Step represents a synthetic input fired at a relative time offset.
type Step struct {
    At   time.Duration
    Keys string // raw bytes to write (may include ANSI sequences)
}

type Script struct {
    Steps []Step
}

// Run executes the script by writing keys to w at scheduled times.
func Run(ctx context.Context, w io.Writer, s Script) {
    start := time.Now()
    for _, st := range s.Steps {
        select {
        case <-ctx.Done():
            return
        case <-time.After(st.At - time.Since(start)):
            _, _ = io.WriteString(w, st.Keys)
        }
    }
}

// Helpers to build common key sequences
const (
    keyUp    = "\x1b[A"
    keyDown  = "\x1b[B"
    keyRight = "\x1b[C"
    keyLeft  = "\x1b[D"
)

func KeysUp(n int) string    { return strings.Repeat(keyUp, n) }
func KeysDown(n int) string  { return strings.Repeat(keyDown, n) }
func KeysRight(n int) string { return strings.Repeat(keyRight, n) }
func KeysLeft(n int) string  { return strings.Repeat(keyLeft, n) }
func KeysType(s string) string { return s }
func KeysEnter() string        { return "\r" }
func KeysSpace() string        { return " " }