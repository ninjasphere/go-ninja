// The package is intended to provide objects used to build implementations of the Driver and Device interfaces
// since there is typically only one possible implementation for many of these methods. This reduces
// the amount of boiler plate code that driver implementors need to write in order to implement the required
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
//
//		// after this time, the framework will start calling methods on the driver.
//
//		return driver
// 	}
//
// For another example of how to use this type, refer to the FakeDriver module of the github.com/ninjasphere/go-ninja/fakedriver package.
//
type DriverSupport struct {
	Info   *model.Module
	Log    *logger.Logger
	Conn   *ninja.Connection
	sender func(event string, payload interface{}) error
}

// This method is called to initialize the Info, Log and Conn members
// of the receiving DriverSupport object and to acquire a named
// connection to the local message bus.
//
// Info is initialized with the supplied *model.Module argument
// which must be non-nil and must have a non-empty ID whose value is
// member referred to here as {id}.
//
// Log is initialized with a Logger member named "{id}.driver".
//
// Conn is initialized with the results of a call to ninja.Connect
// passing {id} as the client id parameter.
// This connection will log to "{id}.connection".
//
// If initialization was not successful for any reason, either because
// the supplied info object was incomplete or because the connection
// attemped failed, the method will return a non-nil error object and
// the receiver should not be used for any further operations.
//
// However, to avoid the need for the caller to acquire its own logging
// object, and provided the receiver itself is not nil, the Log member of
// the receiver will be initialized with a valid Logger even if initialization
// itself fails.
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

	self.Info = info

	conn, err := ninja.Connect(info.ID)
	self.Conn = conn

	return err
}

// Export the driver to the local message bus. After this call completes the driver's Start method
// , and other supported methods, will be called.
//
// The methods parameter is a reference to the full interface of the driver which implements any
// optional driver methods (such as Start and Stop) not implemented by the DriverSupport interface.
// If you do not specify this interface during the export, those methods will not be exposed
// to the RPC subsystem and so will not be called.
//
// This method should not be called until Init has been successfully called. Otherwise, it will
// return a non-nil error.
func (self *DriverSupport) Export(methods ninja.Driver) error {
	err := failIfNotInitialized(self)
	if err == nil {
		return self.Conn.ExportDriver(methods)
	} else {
		return err
	}
}

// Return the module info that describes the driver. This will be nil unless the Init
// method has been called.
func (self *DriverSupport) GetModuleInfo() *model.Module {
	return self.Info
}

// This method can be used by the driver itself to emit a payload on one
// of its own event topics. This method should not be called until both
// the Init and Export methods have been called.
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

// This method is used to receive a reference to the event handler that the driver
// should use to emit events. Consumers of the DriverSupport object should not
// need to override this method, but should instead call SendEvent method as required
// to make use of the handler.
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
	if self == nil ||
		self.Info == nil ||
		self.Log == nil ||
		self.Conn == nil {
		return fmt.Errorf("illegal state: driver has not been successfully initialized")
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
