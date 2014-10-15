// The package is intended to provide objects used to build implementations of the Driver and Device interfaces
// since there is typically only one possible implementation for many of these methods. This will help reduce
// the amount of boiler plate that driver implementors will need to write in order to implement the required
// interfaces.
package support

import (
	"fmt"

	"github.com/juju/loggo"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
)

//
// The DriverSupport object is intended to be used as an anonymous member of
// Driver objects. It contains that part of the driver state that is common
// to most, if not all drivers and allows the driver implementor to
// reuse method implementations that typically do not need
// to vary across implementations.
//
// For an example of how to use this type, refer to the FakeDriver module of the github.com/ninjasphere/go-ninja/fakedriver package.
//
type DriverSupport struct {
	Info   *model.Module
	Log    *logger.Logger
	Conn   *ninja.Connection
	sender func(event string, payload interface{}) error
}

// This method is called to initialize the members of the driver support
// object and acquire a connection to the mqtt broker. The info.ID
// is used to acquire the driver's logger and MQTT connection.
func (self *DriverSupport) Init(info *model.Module) error {
	self.Info = info
	self.Log = logger.GetLogger(self.Info.ID+".driver")
	conn, err := ninja.Connect(self.Info.ID)
	self.Conn = conn
	return err
}

// Export the driver to the local MQTT bus. After this call complete's the driver's Start method
// will be called, if implemented.
//
// The methods parameter is a reference to the full interface of the driver which implements any
// optional driver methods (such as Start and Stop) not implemented by the DriverSupport interface.
// If you do not specify this interface during the export, those methods will not be exposed
// to the RPC subsystem and so will not be called.
func (self *DriverSupport) Export(methods ninja.Driver) error {
     if methods == nil {
		methods = self
     }
     return self.Conn.ExportDriver(methods)
}

// Return the module info that describes the driver.
func (self *DriverSupport) GetModuleInfo() *model.Module {
	return self.Info
}

// Sends an event on one of the driver's event topics.
func (self *DriverSupport) SendEvent(event string, payload interface{}) error {
	return self.sender(event, payload)
}

// Configure the event sender. The event sender is used by the driver implementation
// to emit events relating to the life cycle of driver itself.
func (self *DriverSupport) SetEventHandler(handler func(event string, payload interface{}) error) {
	// FIXME: this method should probably be renamed to SetEventSender.
	self.sender = handler
}

// Configure the log level of the root logger for the driver process.
func (self *DriverSupport) SetLogLevel(level string) error {
	// FIXME: maybe move this implementation into the logger package
	parsed, ok := loggo.ParseLevel(level)
	if ok && parsed != loggo.UNSPECIFIED {
		loggo.GetLogger("").SetLogLevel(parsed)
		return nil
	} else {
		return fmt.Errorf("%s is not a valid logging level")
	}
}
