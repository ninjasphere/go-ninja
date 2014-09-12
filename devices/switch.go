package devices

import (
	"encoding/json"
	"fmt"
	"sync"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"

	"github.com/ninjasphere/go-ninja/channels"
)

type SwitchDeviceState struct {
	OnOff *bool `json:"on-off,omitempty"`
}

func (c *SwitchDeviceState) Clone() *SwitchDeviceState {
	text, _ := json.Marshal(c)
	state := &SwitchDeviceState{}
	json.Unmarshal(text, state)
	return state
}

type SwitchBatchChannel struct {
	switchDevice *SwitchDevice
}

func (c *SwitchBatchChannel) SetBatch(message mqtt.Message, state *SwitchDeviceState, reply *interface{}) error {
	return c.switchDevice.SetBatch(state)
}

type SwitchDevice struct {
	sync.Mutex

	// ApplySwitchState is required, and should actually set the state on the physical switch
	ApplySwitchState func(state *SwitchDeviceState) error

	// The following three are optional, and are used instead of ApplyLightState
	// if only a single channel state is being set.
	ApplyOnOff func(state bool) error

	bus          *ninja.DeviceBus
	state        *SwitchDeviceState
	batch        bool
	log          *logger.Logger
	onOffChannel *channels.OnOffChannel
}

func (d *SwitchDevice) SetSwitchState(state *SwitchDeviceState) error {
	d.Lock()
	defer d.Unlock()

	d.state = state

	if state.OnOff != nil {
		if err := d.onOffChannel.SendEvent("state", *state.OnOff); err != nil {
			return fmt.Errorf("Failed emitting on-off state: %s", err)
		}
	}

	return nil
}

func (d *SwitchDevice) UpdateSwitchOnOffState(on bool) error {
	return d.onOffChannel.SendState(&on)
}

func (d *SwitchDevice) SetBatch(state *SwitchDeviceState) error {
	d.Lock()
	defer d.Unlock()

	mergedState := d.state.Clone()
	if state.OnOff != nil {
		mergedState.OnOff = state.OnOff
	}

	return d.ApplySwitchState(mergedState)
}

func (d *SwitchDevice) SetOnOff(state bool) error {
	d.Lock()
	defer d.Unlock()

	var err error

	if state {
		d.log.Infof("Turning On")
	} else {
		d.log.Infof("Turning Off")
	}

	if d.ApplyOnOff != nil {
		err = d.ApplyOnOff(state)
	} else {
		switchState := d.state.Clone()
		switchState.OnOff = &state

		err = d.ApplySwitchState(switchState)
	}

	return err
}

func (d *SwitchDevice) ToggleOnOff() error {
	if d.state.OnOff == nil {
		d.log.Warningf("On-off channel is in an unknown state for toggling. Setting to off.")
		return d.SetOnOff(false)
	}
	return d.SetOnOff(!*d.state.OnOff)
}

func (d *SwitchDevice) EnableOnOffChannel() error {
	d.onOffChannel = channels.NewOnOffChannel(d)
	return d.bus.AddChannel(d.onOffChannel, "on-off", "on-off")
}

func CreateSwitchDevice(name string, bus *ninja.DeviceBus) (*SwitchDevice, error) {

	switchDevice := &SwitchDevice{
		bus: bus,
		log: logger.GetLogger("SwitchDevice - " + name),
	}

	err := bus.AddChannel(&SwitchBatchChannel{switchDevice}, "core.batching", "core.batching")
	if err != nil {
		return nil, fmt.Errorf("Failed to create batch channel: %s", err)
	}

	switchDevice.log.Infof("Created")

	return switchDevice, nil
}
