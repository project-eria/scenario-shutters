package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	configmanager "github.com/project-eria/config-manager"
	"github.com/project-eria/logger"

	"github.com/jasonlvhit/gocron"
	sunrise "github.com/nathan-osman/go-sunrise"
	"github.com/project-eria/xaal-go/device"
	"github.com/project-eria/xaal-go/engine"
	"github.com/project-eria/xaal-go/schemas"
)

func version() string {
	return fmt.Sprintf("0.0.2 (engine %s)", engine.Version())
}

func setupDev(dev *device.Device) {
	dev.VendorID = "ERIA"
	dev.ProductID = "SenarioShutters"
	dev.Info = "senario.shutters"
	dev.Version = version()
}

const configFile = "scenario-shutters.json"

var (
	dev *device.Device
)

var config = struct {
	Lat         float64       `required:"true"`
	Long        float64       `required:"true"`
	OffsetOpen  time.Duration `default:"60"` // Minutes
	OffsetClose time.Duration `default:"60"` // Minutes
	Devices     map[string]string
	Schedules   []struct {
		Days []string
		Sets []struct {
			Shutters  []string
			OpenTime  string `required:"true"`
			CloseTime string `required:"true"`
		}
	}
}{}

func actionShutters(action string, shutters []string) {
	engine.SendRequest(dev, shutters, action, nil)
}

func schedule() {
	var (
		now         = time.Now()
		weekday     = now.Weekday().String()
		openOffset  = time.Minute * config.OffsetOpen
		closeOffset = time.Minute * config.OffsetClose
	)

	rise, set := sunrise.SunriseSunset(
		config.Lat, config.Long, // Navès
		now.Year(), now.Month(), now.Day(), // date
	)

	riseTime := rise.Add(openOffset).Format("15:04")
	setTime := set.Add(closeOffset).Format("15:04")

	openCloseScheduler.Clear()

	for _, schedule := range config.Schedules {
		if !include(schedule.Days, weekday) {
			continue
		}

		for _, set := range schedule.Sets {
			// Compile list of devices
			shutters := []string{}
			for _, shutter := range set.Shutters {
				addr, ok := config.Devices[shutter]
				if !ok {
					logger.Module("main").WithField("shutter", shutter).Error("Shutter not found in the devices list")
				} else {
					shutters = append(shutters, addr)
				}
			}

			// Set the opening time
			if set.OpenTime == "sunrise" {
				set.OpenTime = riseTime
			}
			logger.Module("main").WithFields(logger.Fields{"time": set.OpenTime, "shutter": set.Shutters}).Info("Opening time set for shutters")
			openCloseScheduler.Every(1).Day().At(set.OpenTime).Do(actionShutters, "up", shutters)

			// Set the closing time
			if set.CloseTime == "sunset" {
				set.CloseTime = setTime
			}
			logger.Module("main").WithFields(logger.Fields{"time": set.CloseTime, "shutter": set.Shutters}).Info("Closing time set for shutters")
			openCloseScheduler.Every(1).Day().At(set.CloseTime).Do(actionShutters, "down", shutters)

		}
		return
	}
}

var openCloseScheduler *gocron.Scheduler

func main() {
	defer os.Exit(0)
	_showVersion := flag.Bool("v", false, "Display the version")
	if !flag.Parsed() {
		flag.Parse()
	}

	// Show version (-v)
	if *_showVersion {
		fmt.Println(version())
		os.Exit(0)
	}

	logger.Module("main").Infof("Starting Scenario-Volets %s...", version())

	// Loading config
	cm, err := configmanager.Init(configFile, &config)
	if err != nil {
		if configmanager.IsFileMissing(err) {
			err = cm.Save()
			if err != nil {
				logger.Module("main").WithField("filename", configFile).Fatal(err)
			}
			logger.Module("main").Fatal("JSON Config file do not exists, created...")
		} else {
			logger.Module("main").WithField("filename", configFile).Fatal(err)
		}
	}

	if err := cm.Load(); err != nil {
		logger.Module("main").Fatal(err)
	}
	defer cm.Close()

	// xAAL engine starting
	engine.Init()

	go engine.Run()
	defer engine.Stop()

	dev = schemas.Basic("")
	setupDev(dev)
	engine.AddDevice(dev)

	// Configure the schedulers
	gocron.Every(1).Day().At("01:00").Do(schedule) // Compute open/close time, every morning
	gocron.Start()

	openCloseScheduler = gocron.NewScheduler()
	openCloseScheduler.Start()

	schedule() // Refresh the schedulers immediately

	// Monitor for config file changes and redo the scheduling
	go func() {
		for {
			cm.Next()
			schedule()
		}
	}()

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until keyboard interrupt is received.
	<-c
	runtime.Goexit()
}
