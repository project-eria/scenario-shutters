package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
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

var config = struct {
	Lat       float64 `required:"true"`
	Long      float64 `required:"true"`
	Devices   map[string]string
	Schedules []struct {
		Days  []string
		Open  []timeAction
		Close []timeAction
	}
}{}

type timeAction struct {
	Shutters []string
	Time     string `required:"true"`
}

var (
	_dev       *device.Device
	_validSun  = regexp.MustCompile(`^(sunrise|sunset)([\+\-]\d+)$`)
	_validHour = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	_rise      time.Time
	_set       time.Time
)

func actionShutters(action string, shutters []string) {
	engine.SendRequest(_dev, shutters, action, nil)
}

func schedule() {
	var (
		now     = time.Now()
		weekday = now.Weekday().String()
	)

	_rise, _set = sunrise.SunriseSunset(
		config.Lat, config.Long, // Navès
		now.Year(), now.Month(), now.Day(), // date
	)

	openCloseScheduler.Clear()

	for _, schedule := range config.Schedules {
		if !include(schedule.Days, weekday) {
			continue
		}

		for _, set := range schedule.Open {
			setTimeAction("up", set)
		}

		for _, set := range schedule.Close {
			setTimeAction("down", set)
		}
		return
	}
}

func setTimeAction(action string, details timeAction) {
	// Compile list of devices
	shutters := getShuttersAddresses(&details.Shutters)

	if _validSun.MatchString(details.Time) {
		var sunTime time.Time

		res := _validSun.FindStringSubmatch(details.Time)

		if res[1] == "sunrise" {
			sunTime = _rise
		} else {
			sunTime = _set
		}

		offset, _ := time.ParseDuration(res[2] + "m")

		details.Time = sunTime.Add(offset).Format("15:04")
	} else if !_validHour.MatchString(details.Time) {
		logger.Module("main").WithFields(logger.Fields{"time": details.Time, "shutter": details.Shutters}).Warn("Incorrect time set for shutters, ignoring")
		return
	}

	logger.Module("main").WithFields(logger.Fields{"time": details.Time, "shutter": details.Shutters}).Infof("%s time set for shutters", action)
	openCloseScheduler.Every(1).Day().At(details.Time).Do(actionShutters, action, shutters)
}

func getShuttersAddresses(shutters *[]string) []string {
	addresses := []string{}
	for _, shutter := range *shutters {
		addr, ok := config.Devices[shutter]
		if !ok {
			logger.Module("main").WithField("shutter", shutter).Error("Shutter not found in the devices list")
		} else {
			addresses = append(addresses, addr)
		}
	}
	return addresses
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

	_dev = schemas.Basic("")
	setupDev(_dev)
	engine.AddDevice(_dev)

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
