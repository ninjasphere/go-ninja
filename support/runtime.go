package support

import (
	"fmt"
	"os"
	"os/signal"
)

// WaitUntilSignal forces the caller to wait until an OS signal is received. This is typically called at the base of the main function
// of a go process.
func WaitUntilSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	panic(fmt.Sprintf("killed by signal %v", s))
}
