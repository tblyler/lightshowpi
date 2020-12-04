package sound

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// FMConfig defines the tunables for an FM instance
type FMConfig struct {
	Frequency float32
}

// FM is a sound.Output implementation that outputs to an FM frequency
type FM struct {
	config *FMConfig
}

// NewFM creates a new FM instance with the given config
func NewFM(config FMConfig) *FM {
	return &FM{
		config: &config,
	}
}

// Play the given input to the FM frequency
func (f *FM) Play(ctx context.Context, filePath string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	soxCmd := exec.CommandContext(
		ctx,
		"sox",
		filePath,
		"-r", "22050",
		"-c", "1",
		"-b", "16",
		"-t", "wav",
		"-",
	)

	reader, err := soxCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open STDOUT pipe for converting the input file to the FM transmitter format: %w", err)
	}

	err = soxCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start conversion for the input file for the FM transmitter: %w", err)
	}

	fmCmd := exec.CommandContext(
		ctx,
		"fm_transmitter",
		"-f",
		fmt.Sprintf("%.1f", f.config.Frequency),
		"-",
	)

	writer, err := fmCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create STDIN pipe for FM transmitter: %w", err)
	}

	err = fmCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start FM transmitter command: %w", err)
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to write audio output to FM transmitter: %w", err)
	}

	err = soxCmd.Wait()
	if err != nil {
		return fmt.Errorf("audio conversion command resulted in an error: %w", err)
	}

	err = fmCmd.Wait()
	if err != nil {
		return fmt.Errorf("FM transmitter command resulted in an error: %w", err)
	}

	return nil
}
