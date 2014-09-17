package devices

import (
	"encoding/json"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"

	"github.com/ninjasphere/go-ninja/channels"
)

type MotionDeviceState struct {
	OnOff *bool `json:"on-off,omitempty"` //on is motion, off is no motion
}

func (c *MotionDeviceState) Clone() *MotionDeviceState {
	text, _ := json.Marshal(c)
	state := &MotionDeviceState{}
	json.Unmarshal(text, state)
	return state
}

type MotionDevice struct {
	bus           *ninja.DeviceBus
	state         *MotionDeviceState
	log           *logger.Logger
	motionChannel *channels.MotionChannel
}

func (d *MotionDevice) UpdateMotionState(on bool) error {
	d.state.OnOff = &on
	return d.motionChannel.SendState(&on)
}

func (d *MotionDevice) EnableMotionChannel() error {
	d.motionChannel = channels.NewMotionChannel(d)
	log.Infof("made motion channel")
	return d.bus.AddChannel(d.motionChannel, "motion", "motion")
}

func CreateMotionDevice(name string, bus *ninja.DeviceBus) (*MotionDevice, error) {

	motionDevice := &MotionDevice{
		bus: bus,
		log: logger.GetLogger("MotionDevice - " + name),
	}

	motionDevice.log.Infof("Created MotionDevice - " + name)

	return motionDevice, nil
}
