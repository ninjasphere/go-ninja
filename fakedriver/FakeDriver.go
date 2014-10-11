package main

import (
	"fmt"
	"log"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
)

var info = ninja.LoadModuleInfo("./package.json")

/*model.Module{
	ID:          "com.ninjablocks.fakedriver",
	Name:        "Fake Driver",
	Version:     "1.0.2",
	Description: "Just used to test go-ninja",
	Author:      "Elliot Shepherd <elliot@ninjablocks.com>",
	License:     "MIT",
}*/

type FakeDriver struct {
	config    *FakeDriverConfig
	conn      *ninja.Connection
	sendEvent func(event string, payload interface{}) error
}

type FakeDriverConfig struct {
	NumberOfLights       int
	NumberOfMediaPlayers int
}

func defaultConfig() *FakeDriverConfig {
	return &FakeDriverConfig{
		NumberOfLights:       5,
		NumberOfMediaPlayers: 1,
	}
}

func NewFakeDriver() (*FakeDriver, error) {

	conn, err := ninja.Connect("FakeDriver")

	if err != nil {
		log.Fatalf("Failed to create fake driver: %s", err)
	}

	driver := &FakeDriver{
		conn: conn,
	}

	err = conn.ExportDriver(driver)

	if err != nil {
		log.Fatalf("Failed to export fake driver: %s", err)
	}

	/*go func() {
		time.Sleep(time.Second)
		driver.Start(nil, nil, nil)
	}()*/

	return driver, nil
}

func (d *FakeDriver) Start(config *FakeDriverConfig) error {
	log.Printf("Fake Driver Starting with config %v", config)

	d.config = config

	for i := 0; i < d.config.NumberOfLights; i++ {
		log.Print("Creating new fake light")
		device := NewFakeLight(d, i)

		err := d.conn.ExportDevice(device)
		if err != nil {
			log.Fatalf("Failed to export fake light %d: %s", i, err)
		}

		err = d.conn.ExportChannel(device, device.onOffChannel, "on-off")
		if err != nil {
			log.Fatalf("Failed to export fake light on off channel %d: %s", i, err)
		}

		err = d.conn.ExportChannel(device, device.brightnessChannel, "brightness")
		if err != nil {
			log.Fatalf("Failed to export fake light brightness channel %d: %s", i, err)
		}

		err = d.conn.ExportChannel(device, device.colorChannel, "color")
		if err != nil {
			log.Fatalf("Failed to export fake color channel %d: %s", i, err)
		}
	}

	// Bump the config prop by one... to test it updates
	config.NumberOfLights++

	if d.config.NumberOfMediaPlayers == 0 {
		d.config.NumberOfMediaPlayers = 1
	}

	for i := 0; i < d.config.NumberOfMediaPlayers; i++ {
		log.Print("Creating new fake media player")
		_, err := NewFakeMediaPlayer(d, d.conn, i)
		if err != nil {
			log.Fatalf("failed to create fake media player")
		}
	}

	return d.sendEvent("config", config)
}

func (d *FakeDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")
}

type In struct {
	Name string
}

type Out struct {
	Age  int
	Name string
}

func (d *FakeDriver) Blarg(in *In) (*Out, error) {
	log.Printf("GOT INCOMING! %s", in.Name)
	return &Out{
		Name: in.Name,
		Age:  30,
	}, nil
}

func (d *FakeDriver) GetModuleInfo() *model.Module {
	return info
}

func (d *FakeDriver) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.sendEvent = sendEvent
}
