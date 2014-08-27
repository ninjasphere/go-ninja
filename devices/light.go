package devices

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/logger"

	"github.com/ninjasphere/go-ninja/channels"
)

var log = logger.GetLogger("LightDevice")

type LightDeviceState struct {
	OnOff      *bool                `json:"on-off,omitempty"`
	Color      *channels.ColorState `json:"color,omitempty"`
	Brightness *float64             `json:"brightness,omitempty"`
	Transition *int                 `json:"transition,omitempty"`
}

func (c *LightDeviceState) Clone() *LightDeviceState {
	text, _ := json.Marshal(c)
	state := &LightDeviceState{}
	json.Unmarshal(text, state)
	return state
}

type LightBatchChannel struct {
	light *LightDevice
}

func (c *LightBatchChannel) SetBatch(state *LightDeviceState, reply *interface{}) error {
	return c.light.SetBatch(state)
}

type LightDevice struct {
	sync.Mutex

	// SetLightState is required, and should actually set the state on the physical light
	ApplyLightState func(state *LightDeviceState) error

	// The following three are optional, and are used instead of ApplyLightState
	// if only a single channel state is being set.
	ApplyOnOff      func(state bool) error
	ApplyBrightness func(state float64) error
	ApplyColor      func(state *channels.ColorState) error
	ApplyTransition func(state int) error

	bus   *ninja.DeviceBus
	state *LightDeviceState
	batch bool
	log   *logger.Logger

	onOff      *channels.OnOffChannel
	brightness *channels.BrightnessChannel
	color      *channels.ColorChannel
	transition *channels.TransitionChannel
}

func (d *LightDevice) SetLightState(state *LightDeviceState) error {
	d.Lock()

	var err error

	d.state = state

	if state.OnOff != nil {
		err = d.onOff.SendEvent("state", *state.OnOff)
	}

	if err != nil {
		return fmt.Errorf("Failed emitting on-off state: %s", err)
	}

	if state.Brightness != nil {
		err = d.brightness.SendEvent("state", *state.Brightness)
	}

	if err != nil {
		return fmt.Errorf("Failed emitting brightness state: %s", err)
	}

	if state.Color != nil {
		err = d.color.SendEvent("state", *state.Color)
	}

	if err != nil {
		return fmt.Errorf("Failed emitting color state: %s", err)
	}

	if state.Transition != nil {
		err = d.transition.SendEvent("state", *state.Transition)
	}

	if err != nil {
		return fmt.Errorf("Failed emitting transition state: %s", err)
	}

	d.Unlock()
	return nil
}

func (d *LightDevice) SetBatch(state *LightDeviceState) error {
	d.Lock()

	mergedState := d.state.Clone()
	if state.OnOff != nil {
		mergedState.OnOff = state.OnOff
	}
	if state.Brightness != nil {
		mergedState.Brightness = state.Brightness
	}
	if state.Color != nil {
		mergedState.Color = state.Color
	}
	if state.Transition != nil {
		mergedState.Transition = state.Transition
	}
	err := d.ApplyLightState(mergedState)

	d.Unlock()
	return err
}

func (d *LightDevice) SetOnOff(state bool) error {
	d.Lock()

	var err error

	if state == true {
		d.log.Infof("Turning Off")
	} else {
		d.log.Infof("Turning On")
	}

	if d.ApplyOnOff != nil {
		err = d.ApplyOnOff(state)
	} else {
		lightState := d.state.Clone()
		lightState.OnOff = &state

		err = d.ApplyLightState(lightState)
	}

	d.Unlock()
	return err
}

func (d *LightDevice) SetBrightness(state float64) error {
	if d.brightness == nil {
		return fmt.Errorf("This device does not have a brightness channel")
	}

	d.Lock()

	var err error

	d.log.Infof("Setting brightness to %f", state)

	if d.ApplyBrightness != nil {
		err = d.ApplyBrightness(state)
	} else {
		lightState := d.state.Clone()
		lightState.Brightness = &state

		err = d.ApplyLightState(lightState)
	}

	d.Unlock()
	return err
}

func (d *LightDevice) SetColor(state *channels.ColorState) error {
	if d.color == nil {
		return fmt.Errorf("This device does not have a color channel")
	}

	d.Lock()

	var err error

	json, _ := json.Marshal(state)
	d.log.Infof("Setting Color to %s", json)

	if d.ApplyColor != nil {
		err = d.ApplyColor(state)
	} else {
		lightState := d.state.Clone()
		lightState.Color = state

		err = d.ApplyLightState(lightState)
	}

	d.Unlock()
	return err
}

func (d *LightDevice) SetTransition(state int) error {
	if d.transition == nil {
		return fmt.Errorf("This device does not have a transition channel")
	}

	d.Lock()

	d.state.Transition = &state

	d.Unlock()
	return nil
}

func (d *LightDevice) ToggleOnOff() error {
	if d.state.OnOff == nil {
		d.log.Warningf("On-off channel is in an unknown state for toggling. Setting to off.")
		return d.SetOnOff(false)
	}
	return d.SetOnOff(!*d.state.OnOff)
}

func (d *LightDevice) EnableOnOffChannel() error {
	d.onOff = channels.NewOnOffChannel(d)
	return d.bus.AddChannel(d.onOff, "on-off", "on-off")
}

func (d *LightDevice) EnableBrightnessChannel() error {
	d.brightness = channels.NewBrightnessChannel(d)
	return d.bus.AddChannel(d.brightness, "brightness", "brightness")
}

func (d *LightDevice) EnableColorChannel() error {
	d.color = channels.NewColorChannel(d)
	return d.bus.AddChannel(d.color, "color", "color")
}

func (d *LightDevice) EnableTransitionChannel() error {
	d.transition = channels.NewTransitionChannel(d)
	return d.bus.AddChannel(d.transition, "transition", "transition")
}

func CreateLightDevice(name string, bus *ninja.DeviceBus) (*LightDevice, error) {

	light := &LightDevice{
		bus: bus,
		log: logger.GetLogger("LightDevice - " + name),
	}

	bus.AddChannel(&LightBatchChannel{light}, "core.batching", "batch")

	light.log.Infof("Created")

	return light, nil
}
