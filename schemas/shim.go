package schemas

// Stolen shamelessly from paypal/gatt

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// cshim provides access to BLE via an external c executable.
type shim struct {
	sync.Mutex
	cmd *exec.Cmd
	io.Reader
	io.Writer
	bufio.Scanner
}

// newCShim starts the shim named file using the provided args.
func newCShim(file string, arg ...string) (*shim, error) {
	c := new(shim)
	var err error
	if file, err = exec.LookPath(file); err != nil {
		return nil, err
	}
	c.cmd = exec.Command(file, arg...)
	if c.Writer, err = c.cmd.StdinPipe(); err != nil {
		return nil, err
	}
	if c.Reader, err = c.cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = c.cmd.Start(); err != nil {
		return nil, err
	}

	c.Mutex = sync.Mutex{}
	c.Scanner = *bufio.NewScanner(c)

	return c, err
}

func (c *shim) Wait() error                { return c.cmd.Wait() }
func (c *shim) Close() error               { return c.cmd.Process.Kill() }
func (c *shim) Interrupt() error           { return c.cmd.Process.Signal(syscall.SIGINT) }
func (c *shim) Signal(sig os.Signal) error { return c.cmd.Process.Signal(sig) }
