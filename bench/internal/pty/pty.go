package pty

import (
	"context"
	"os"
	"os/exec"

	ptylib "github.com/creack/pty"
)

// SpawnWithSize starts cmd in a PTY of size cols x rows and returns the cmd and the PTY file.
func SpawnWithSize(ctx context.Context, cmd *exec.Cmd, cols, rows int) (*exec.Cmd, *os.File, error) {
	ws := &ptylib.Winsize{Cols: uint16(cols), Rows: uint16(rows)}
	ptmx, err := ptylib.StartWithSize(cmd, ws)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		<-ctx.Done()
		_ = ptmx.Close() // close when context cancelled
	}()
	return cmd, ptmx, nil
}
