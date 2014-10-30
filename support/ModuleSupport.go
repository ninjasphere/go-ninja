package support

import (
	"fmt"

	"github.com/juju/loggo"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
)

// ModuleSupport is contains implementations of methods that are common to all modules whether they
// be apps or drivers. It provides access to, the module information, the logger and the connection
// to the local message bus.
type ModuleSupport struct {
	Info   *model.Module
	Log    *logger.Logger
	Conn   *ninja.Connection
	sender func(event string, payload interface{}) error
}

// This method is called to initialize the Info, Log and Conn members
// of the receiving ModuleSupport object and to acquire a named
// connection to the local message bus.
//
// Info is initialized with the supplied *model.Module argument
// which must be non-nil and must have a non-empty ID whose value is
// member referred to here as {id}.
//
// Log is initialized with a Logger member named "{id}.module".
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
func (self *ModuleSupport) Init(info *model.Module) error {
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

// Return the module info that describes the module. This will be nil unless the Init
// method has been called.
func (self *ModuleSupport) GetModuleInfo() *model.Module {
	return self.Info
}

// This method can be used by the module itself to emit a payload on one
// of its own event topics. This method should not be called until both
// the Init and Export methods have been called.
func (self *ModuleSupport) SendEvent(event string, payload interface{}) error {
	err := failIfNotInitialized(self)
	if err == nil {
		if self.sender != nil {
			return self.sender(event, payload)
		} else {
			return fmt.Errorf("illegal state: module has not been exported")
		}
	} else {
		return err
	}
}

// This method is used to receive a reference to the event handler that the module
// should use to emit events. Consumers of the ModuleSupport object should not
// need to override this method, but should instead call SendEvent method as required
// to make use of the handler.
func (self *ModuleSupport) SetEventHandler(handler func(event string, payload interface{}) error) {
	// FIXME: this method should probably be renamed to SetEventSender.
	self.sender = handler
}

// Configure the og level of the root logger for the module's process.
func (self *ModuleSupport) SetLogLevel(level string) error {
	// FIXME: maybe move this implementation into the logger package
	parsed, ok := loggo.ParseLevel(level)
	if ok && parsed != loggo.UNSPECIFIED {
		loggo.GetLogger("").SetLogLevel(parsed)
		safeLog(self, nil).Logf(parsed, "Log level has been reset to %s", level)
		return nil
	} else {
		return fmt.Errorf("%s is not a valid logging level")
	}
}

// Return an error if the receiver has not been successfully initialized.
func failIfNotInitialized(self *ModuleSupport) error {
	if self == nil ||
		self.Info == nil ||
		self.Log == nil ||
		self.Conn == nil {
		return fmt.Errorf("illegal state: module has not been successfully initialized")
	} else {
		return nil
	}
}

// Given a possible nil or uninitialized module, always return
// a string that identifies the module in some fashion.
func safeID(info *model.Module) string {
	if info == nil || info.ID == "" {
		return "{uninitialized-module-id}"
	} else {
		return info.ID
	}
}

// this function will always return a logger that can be used even if the
// support object has not been initialized in the correct sequence or with
// the correct arguments.
func safeLog(self *ModuleSupport, info *model.Module) *logger.Logger {
	if self == nil || self.Log == nil {
		return logger.GetLogger(fmt.Sprintf("%s.module", safeID(info)))
	} else {
		return self.Log
	}
}
