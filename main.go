package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tblyler/lightshowpi/config"
	"github.com/tblyler/lightshowpi/light"
	"github.com/tblyler/lightshowpi/light/rpi"
	"github.com/tblyler/lightshowpi/sound"
	"golang.org/x/sync/errgroup"
)

func readConfig(filePath string) (*config.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}

	defer file.Close()

	rawConfigData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	config, err := config.ParseConfig(rawConfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return config, nil
}

func main() {
	err := func() error {
		cfg, err := readConfig("/etc/lightshowpi.toml")
		if err != nil {
			return err
		}

		piManager, err := rpi.NewManager(*cfg.RPIManagerConfig())
		if err != nil {
			return err
		}

		defer piManager.Close()

		basicLights := map[string]light.BasicLight{}

		for alias, info := range cfg.RaspberryPi.BasicLights {
			basicLight, err := piManager.GetBasicLight(info.Pin)
			if err != nil {
				return fmt.Errorf("failed to get raspberry pi basic light %s for pin %d: %w", alias, info.Pin, err)
			}

			basicLights[alias] = basicLight
		}

		basicLightConductorSchedule, err := light.NewBasicLightConductorScheduleFromFile(cfg.LightSchedulePath)
		if err != nil {
			return err
		}

		basicLightConductor := light.NewBasicLightConductor(basicLights, basicLightConductorSchedule)
		soundOuput := sound.NewFM(cfg.FMOutput)

		errGroup, errGroupCtx := errgroup.WithContext(context.TODO())
		errGroup.Go(func() error {
			return basicLightConductor.Start(errGroupCtx)
		})

		errGroup.Go(func() error {
			return soundOuput.Play(errGroupCtx, cfg.SongPath)
		})

		return errGroup.Wait()
	}()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
