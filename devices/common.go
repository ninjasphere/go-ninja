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
