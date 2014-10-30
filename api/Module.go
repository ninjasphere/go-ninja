package ninja

import (
	"github.com/ninjasphere/go-ninja/model"
)

//
// Driver implementors should provide an implementation of this interface.
//
// The methods of this interface are used by the RPC framework to learn about
// the driver and also to configure the driver with a callback that the
// driver can use to emit events that comply with the driver service
// specification (http://schema.ninjablocks.com/service/driver)
//
// Implementors can inherit default implementations of these methods
// by including an anonymous struct member of type support.DriverSupport.
//
type Module interface {
	GetModuleInfo() *model.Module
	SetEventHandler(func(event string, payload interface{}) error)
}

// A default implementation of this method is provided by support.DriverSupport.
type LogControl interface {
	// Sets the current log level for the driver process, specified as a string.
	//
	// This method is intended to called remotely in cases where it is necessary
	// to adjust the logging levels in order to acquire additional diagnostic information
	// from the driver
	//
	// Available logging levels are as per http://godoc.org/github.com/juju/loggo#Level
	SetLogLevel(level string) error
}
