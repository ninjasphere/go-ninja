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
// Driver objects. It contains that part of the Driver state that is common
// to most, if not all driver implementations and allows the driver implementor to
// reuse method implementations that typically do not need
// to vary across implementations.
//
// Example Usage
//
//	import "github.com/ninjasphere.com/support"
//
//	type acmeDriver {
// 		support.DriverSupport
// 		// driver specific state ...
// 	}
//
// 	func newAcmeDriver() (*acmeDriver, error) {
// 		driver := &acmeDriver{}
//
// 		err := driver.Init(info)
// 		if err != nil {
// 			return err
// 		}
//
// 		err = driver.Export(driver)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// For another example of how to use this type, refer to the FakeDriver module of the github.com/ninjasphere/go-ninja/fakedriver package.
//
type DriverSupport struct {
	Info        *model.Module
	Log         *logger.Logger
	Conn        *ninja.Connection
	sender      func(event string, payload interface{}) error
	initialized bool
}

// This method is called to initialize the members of a DriverSupport object
// and acquire a named connection to local message bus.
//
// This method initializes the following member variables: Info, Log and Conn.
//
// The logger will be named {id}.driver where id is the ID of the supplied Module model.
// The Connection will be named {id}.
//
// If initialization was not successful an non-nil error will be returned.
//
// self.Log will is guaranteed to be non-nil even if initialization fails, provided that
// the receiver (self) is not null.
//
func (self *DriverSupport) Init(info *model.Module) error {
	log := safeLog(self, info)

	if self == nil {
		return fmt.Errorf("assertion failed: receiver != nil")
	}

	self.Log = log

	if info == nil {
		return fmt.Errorf("invalid argument: info == nil")
	}

	if info.ID == "" {
		return fmt.Errorf("invalid argument: info.ID == \"\"")
	}

	conn, err := ninja.Connect(info.ID)

	self.Info = info
	self.Conn = conn

	self.initialized = (err == nil)

	return err
}

// Export the driver to the local MQTT bus. After this call completes the driver's Start method
// will be called, if it exists.
//
// The methods parameter is a reference to the full interface of the driver which implements any
// optional driver methods (such as Start and Stop) not implemented by the DriverSupport interface.
// If you do not specify this interface during the export, those methods will not be exposed
// to the RPC subsystem and so will not be called.
//
// This method should not be called until Init is called.
func (self *DriverSupport) Export(methods ninja.Driver) error {
	err := failIfNotInitialized(self)
	if err == nil {
		return self.Conn.ExportDriver(methods)
	} else {
		return err
	}
}

// Return the module info that describes the driver.
func (self *DriverSupport) GetModuleInfo() *model.Module {
	return self.Info
}

// Sends an event on one of the driver's event topics.
func (self *DriverSupport) SendEvent(event string, payload interface{}) error {
	err := failIfNotInitialized(self)
	if err == nil {
		if self.sender != nil {
			return self.sender(event, payload)
		} else {
			return fmt.Errorf("illegal state: driver has not been exported")
		}
	} else {
		return err
	}
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

// Return an error if the receiver has not been successfully initialized.
func failIfNotInitialized(self *DriverSupport) error {
	if self == nil || !self.initialized {
		return fmt.Errorf("illegal state: driver has not been initialized")
	} else {
		return nil
	}
}

// Given a possible nil or uninitialized module, always return
// a string that identifies the driver in some fashion.
func safeID(info *model.Module) string {
	if info == nil || info.ID == "" {
		return "{uninitialized-driver-id}"
	} else {
		return info.ID
	}
}

// this function will always return a logger that can be used even if the
// support object has not been initialized in the correct sequence or with
// the correct arguments.
func safeLog(self *DriverSupport, info *model.Module) *logger.Logger {
	if self == nil || self.Log == nil {
		return logger.GetLogger(fmt.Sprintf("%s.driver", safeID(info)))
	} else {
		return self.Log
	}
}
