package devices

import (
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
)

type baseDevice struct {
	log       *logger.Logger
	driver    ninja.Driver
	info      *model.Device
	conn      *ninja.Connection
	sendEvent func(event string, payload interface{}) error
}

func (d *baseDevice) GetDeviceInfo() *model.Device {
	return d.info
}

func (d *baseDevice) GetDriver() ninja.Driver {
	return d.driver
}

func (d *baseDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.sendEvent = sendEvent
}

func (d *baseDevice) Log() *logger.Logger {
	return d.log
}

// Re-expressed with public members so that drivers that don't want to use the Device structs devices.*
// can do so without re-implementing all the Device methods themselves.

type BaseDevice struct {
	Driver ninja.Driver
	Info   *model.Device

	Conn *ninja.Connection
	Log_ *logger.Logger

	SendEvent func(event string, payload interface{}) error
}

func (d *BaseDevice) GetDeviceInfo() *model.Device {
	return d.Info
}

func (d *BaseDevice) GetDriver() ninja.Driver {
	return d.Driver
}

func (d *BaseDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.SendEvent = sendEvent
}

func (d *BaseDevice) Log() *logger.Logger {
	return d.Log_
}
