package main

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/kelvins/sunrisesunset"
	"github.com/project-eria/eria-base"
	"github.com/project-eria/logger"
	"github.com/project-eria/xaal-go"
	"github.com/project-eria/xaal-go/device"
	"github.com/project-eria/xaal-go/schemas"
	"github.com/project-eria/xaal-go/utils"
)

var (
	// Version is a placeholder that will receive the git tag version during build time
	Version = "-"
)

func setupDev(dev *device.Device) {
	dev.VendorID = "ERIA"
	dev.ProductID = "SenarioShutters"
	dev.Info = "senario.shutters"
	dev.Version = Version
}

const configFile = "scenario-shutters.json"

var config = struct {
	Lat     float64 `required:"true"`
	Long    float64 `required:"true"`
	Devices map[string]string
	Events  []struct {
		Label string
		Value bool
	}
	Open  []TimeSchedule
	Close []TimeSchedule
}{}

type TimeSchedule struct {
	Days    []string
	Actions []ActionSchedule
}

type ActionSchedule struct {
	Shutters []string
	Time     string `required:"true"`
	Min      string
	Max      string
}

var (
	_dev       *device.Device
	_validSun  = regexp.MustCompile(`^(sunrise|sunset)([\+\-]\d+)?$`)
	_validHour = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	_rise      time.Time
	_set       time.Time
	_location  *time.Location
)

func Scheduleshutters(action string, shutters []string) {
	xaal.SendRequest(_dev, shutters, action, nil)
}

func schedule() {
	var (
		now     = time.Now()
		weekday = now.Weekday().String()
		offset  float64
		err     error
		event   string
	)

	// Find if there an event has been triggered
	event = findCurrentEvent()

	tzOffset := now.Format("-07") // Get the timezone offset
	offset, err = strconv.ParseFloat(tzOffset, 64)
	if err != nil {
		logger.Module("main").Error(err)
		return
	}
	p := sunrisesunset.Parameters{
		Latitude:  config.Lat,
		Longitude: config.Long,
		UtcOffset: offset,
		Date:      now,
	}
	// Calculate the sunrise and sunset times
	_rise, _set, err = p.GetSunriseSunset()
	// Clean the date component for future comparisons
	_rise, _ = time.Parse("15:04", _rise.Format("15:04"))
	_set, _ = time.Parse("15:04", _set.Format("15:04"))

	if err != nil {
		logger.Module("main").Error(err)
		return
	}
	openCloseScheduler.Clear()

	setActions("up", config.Open, weekday, event)
	setActions("down", config.Close, weekday, event)
}

func findCurrentEvent() string {
	for _, event := range config.Events {
		if event.Value == true {
			return event.Label
		}
	}
	return ""
}

func setActions(action string, timeSchedules []TimeSchedule, weekday string, event string) {
	var (
		actions []ActionSchedule
	)

	for _, timeSchedule := range timeSchedules {
		// Default to weekday
		if _, in := utils.SliceContains(&timeSchedule.Days, weekday); in {
			actions = timeSchedule.Actions
			if event == "" {
				break // stop the loop
			}
		}

		if event != "" {
			// Search for event schedule
			if _, in := utils.SliceContains(&timeSchedule.Days, event); in {
				actions = timeSchedule.Actions
				break // stop the loop
			}
		}
	}

	for _, set := range actions {
		setAction(action, set)
	}
}

func setAction(action string, details ActionSchedule) {
	// Compile list of devices
	shutters := getShuttersAddresses(&details.Shutters)

	if _validSun.MatchString(details.Time) {
		var sunTime time.Time

		res := _validSun.FindStringSubmatch(details.Time)

		if res[1] == "sunrise" {
			if res[2] != "" {
				offset, _ := time.ParseDuration(res[2] + "m")
				_rise = _rise.Add(offset)
			}

			if minTime, err := time.Parse("15:04", details.Min); err == nil && minTime.After(_rise) {
				sunTime = minTime
			} else {
				if err != nil && details.Min != "" {
					logger.Module("main").WithFields(logger.Fields{"min": details.Min}).Warn("Incorrect min value, ignoring")
				}
				sunTime = _rise
			}
		} else {
			if res[2] != "" {
				offset, _ := time.ParseDuration(res[2] + "m")
				_set = _set.Add(offset)
			}
			if maxTime, err := time.Parse("15:04", details.Max); err == nil && maxTime.Before(_set) {
				sunTime = maxTime
			} else {
				if err != nil && details.Max != "" {
					logger.Module("main").WithFields(logger.Fields{"max": details.Max}).Warn("Incorrect max value, ignoring")
				}
				sunTime = _set
			}
		}

		details.Time = sunTime.Format("15:04")

	} else if !_validHour.MatchString(details.Time) {
		logger.Module("main").WithFields(logger.Fields{"time": details.Time, "shutter": details.Shutters}).Warn("Incorrect time set for shutters, ignoring")
		return
	}

	logger.Module("main").WithFields(logger.Fields{"time": details.Time, "shutter": details.Shutters}).Infof("%s time set for shutters", action)
	openCloseScheduler.Every(1).Day().At(details.Time).Do(Scheduleshutters, action, shutters)
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

	eria.AddShowVersion(Version)

	logger.Module("main").Infof("Starting Scenario-Volets %s...", Version)

	// Loading config
	cm := eria.LoadConfig(configFile, &config)
	defer cm.Close()

	// Init xAAL engine
	eria.InitEngine()

	_dev = schemas.Basic("")
	setupDev(_dev)
	xaal.AddDevice(_dev)

	go xaal.Run()
	defer xaal.Stop()

	// Configure the schedulers
	gocron.Every(1).Day().At("02:00").Do(schedule) // Compute open/close time, every morning
	gocron.Start()

	openCloseScheduler = gocron.NewScheduler()
	openCloseScheduler.Start()

	schedule() // Refresh the schedulers immediately

	// Monitor for config file changes and redo the scheduling
	/*
		go func() {
			for {
				cm.Next()
				schedule()
			}
		}()
	*/
	eria.WaitForExit()
}
