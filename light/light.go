package light

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

// BasicLight describes a light that simply can be turned on/off, like a relay
type BasicLight interface {
	On() error
	Off() error
	State() (isOn bool, err error)
	Close() error
}

// BasicLightConductorSchedule says which basic lights should be on/off for the second that maps to their index
type BasicLightConductorSchedule struct {
	On  [][]string
	Off [][]string
}

// NewBasicLightConductorScheduleFromFile creates a new asicLightConductorSchedule instance from a file path
func NewBasicLightConductorScheduleFromFile(filePath string) (*BasicLightConductorSchedule, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open basic light conductor schedule file at %s: %w", filePath, err)
	}

	defer file.Close()

	csvReader := csv.NewReader(file)

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get header for basic light conductor schedule file %s: %w", filePath, err)
	}

	schedule := &BasicLightConductorSchedule{}
	for lineNumber := 2; true; lineNumber++ {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("failed to get record at line %d of basic light conductor schedule file %s: %w", lineNumber, filePath, err)
		}

		onRow := []string{}
		offRow := []string{}

		for columnNumber, val := range record {
			switch val {
			case "on":
				onRow = append(onRow, header[columnNumber])
			case "off":
				offRow = append(offRow, header[columnNumber])

			default:
				return nil, fmt.Errorf(
					"invalid setting %s at line %d column %d of basic light conductor schedule file %s: %w",
					val,
					lineNumber,
					columnNumber+1,
					filePath,
					err,
				)
			}
		}

		schedule.On = append(schedule.On, onRow)
		schedule.Off = append(schedule.Off, offRow)
	}

	return schedule, nil
}

// BasicLightConductor turns basic lights on/off for a given schedule when told to conduct
type BasicLightConductor struct {
	basicLights map[string]BasicLight
	schedule    *BasicLightConductorSchedule
}

// NewBasicLightConductor creates a basic light conductor instance for the given basic lights and schedule
func NewBasicLightConductor(basicLights map[string]BasicLight, schedule *BasicLightConductorSchedule) *BasicLightConductor {
	return &BasicLightConductor{
		basicLights: basicLights,
		schedule:    schedule,
	}
}

func (blc *BasicLightConductor) setStatesForSecondInterval(i int) error {
	errGroup := errgroup.Group{}

	errGroup.Go(func() error {
		for _, alias := range blc.schedule.Off[i] {
			basicLight, ok := blc.basicLights[alias]
			if !ok {
				return fmt.Errorf("invalid alias %s for basic light provided at schedule index %d", alias, i)
			}

			err := basicLight.Off()
			if err != nil {
				return fmt.Errorf("failed to turn off basic light alias %s: %w", alias, err)
			}
		}

		return nil
	})

	errGroup.Go(func() error {
		for _, alias := range blc.schedule.On[i] {
			basicLight, ok := blc.basicLights[alias]
			if !ok {
				return fmt.Errorf("invalid alias %s for basic light provided at schedule index %d", alias, i)
			}

			err := basicLight.On()
			if err != nil {
				return fmt.Errorf("failed to turn on basic light alias %s: %w", alias, err)
			}
		}

		return nil
	})

	return errGroup.Wait()
}

// Start begins turning the basic lights on/off for the given schedule
func (blc *BasicLightConductor) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	scheduleLen := len(blc.schedule.Off)
	if onLen := len(blc.schedule.On); scheduleLen != onLen {
		return fmt.Errorf("schedule for on/off light aliases do not have the same amount of entries! on: %d off: %d", onLen, scheduleLen)
	}

	err := blc.setStatesForSecondInterval(0)
	if err != nil {
		return err
	}

	for i := 1; i < scheduleLen; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			err = blc.setStatesForSecondInterval(i)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
