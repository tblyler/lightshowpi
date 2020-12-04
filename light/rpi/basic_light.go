package rpi

import (
	"fmt"
	"sync"

	"github.com/warthog618/gpiod"
)

// BasicLight is a BasicLight implementation for the Raspberry Pi
type BasicLight struct {
	pin      int
	line     *gpiod.Line
	manager  *Manager
	isOn     bool
	isOnLock sync.Mutex
}

// On turns on the light
func (bl *BasicLight) On() error {
	bl.isOnLock.Lock()
	defer bl.isOnLock.Unlock()

	if bl.isOn {
		return nil
	}

	err := bl.line.SetValue(0)
	if err != nil {
		return fmt.Errorf("failed to set pin %d to low (on): %w", bl.pin, err)
	}

	bl.isOn = true

	return nil
}

// Off turns off the light
func (bl *BasicLight) Off() error {
	bl.isOnLock.Lock()
	defer bl.isOnLock.Unlock()

	if !bl.isOn {
		return nil
	}

	err := bl.line.SetValue(1)
	if err != nil {
		return fmt.Errorf("failed to set pin %d to high (on): %w", bl.pin, err)
	}

	bl.isOn = false

	return nil
}

// State of whether the light is on or off
func (bl *BasicLight) State() (isOn bool, err error) {
	bl.isOnLock.Lock()
	defer bl.isOnLock.Unlock()

	return bl.isOn, nil
}

// Close underlying open connection
func (bl *BasicLight) Close() error {
	return bl.manager.closeLine(bl.pin)
}
