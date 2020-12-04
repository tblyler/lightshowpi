package rpi

import (
	"errors"
	"fmt"
	"sync"

	"github.com/warthog618/gpiod"
)

const (
	// DefaultChipName to try to use if one is not provided
	DefaultChipName = "gpiochip0"
)

var (
	// ErrLineAlreadyRequested denotes that the requested line is not available because it has already been requested and not Closed
	ErrLineAlreadyRequested = errors.New("line has already been requested")

	// ErrLineNotRequested denotes that the given line cannot be acted on because it was not previously requested
	ErrLineNotRequested = errors.New("line was not previously requested")
)

// ManagerConfig configuration for a Manager
type ManagerConfig struct {
	ChipName string
}

// GetChipName from the config
func (mc *ManagerConfig) GetChipName() string {
	if mc.ChipName == "" {
		return DefaultChipName
	}

	return mc.ChipName
}

// Manager for overarching raspberry pi GPIO management
type Manager struct {
	config    *ManagerConfig
	chip      *gpiod.Chip
	lines     map[int]*gpiod.Line
	linesLock sync.Mutex
}

// NewManager creates a new Manager instance for the given config
func NewManager(config ManagerConfig) (*Manager, error) {
	chip, err := gpiod.NewChip(config.GetChipName())
	if err != nil {
		return nil, fmt.Errorf("failed to open GPIO chip %s: %w", config.GetChipName(), err)
	}

	return &Manager{
		config: &config,
		chip:   chip,
		lines:  make(map[int]*gpiod.Line),
	}, nil
}

// GetBasicLight for the given pin if available
func (m *Manager) GetBasicLight(pin int) (*BasicLight, error) {
	line, err := m.requestLine(pin)
	if err != nil {
		return nil, err
	}

	return &BasicLight{
		pin:     pin,
		line:    line,
		manager: m,
	}, nil
}

func (m *Manager) requestLine(pin int) (*gpiod.Line, error) {
	m.linesLock.Lock()
	defer m.linesLock.Unlock()

	_, exists := m.lines[pin]
	if exists {
		return nil, fmt.Errorf("%w pin %d", ErrLineAlreadyRequested, pin)
	}

	line, err := m.chip.RequestLine(pin, gpiod.AsOutput(0))
	if err != nil {
		return nil, fmt.Errorf("failed to request line for pin %d: %w", pin, err)
	}

	m.lines[pin] = line

	return line, nil
}

func (m *Manager) closeLine(pin int) error {
	m.linesLock.Lock()
	defer m.linesLock.Unlock()

	return m.closeLineHelper(pin)
}

func (m *Manager) closeLineHelper(pin int) error {
	line := m.lines[pin]
	if line == nil {
		return fmt.Errorf("%w pin %d", ErrLineNotRequested, pin)
	}

	delete(m.lines, pin)

	return line.Close()
}

// Close this Manager and all of its associated lines
func (m *Manager) Close() error {
	m.linesLock.Lock()
	defer m.linesLock.Unlock()

	for pin := range m.lines {
		m.closeLineHelper(pin)
	}

	return m.chip.Close()
}
