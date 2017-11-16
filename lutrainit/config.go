package main

import (
	"dev.sigpipe.me/dashie/lutrainit/shared/ipc"
	"fmt"
	"github.com/go-clog/clog"
	"github.com/go-ini/ini"
	"io/ioutil"
	"strings"
	"time"
)

var (
	// MainConfig of the daemon
	MainConfig struct {
		Persist    bool
		Autologins []string

		Log struct {
			Filename  string
			Rotate    bool
			Daily     bool
			MaxSize   int
			MaxLines  int64
			MaxDays   int64
			BufferLen int64
		}

		StartedReexec bool
	}
)

/*
description
pidfile
type
autostart
execprestart...
*/

// ParseConfig a single config file into the services it provides
func ParseConfig(baseDir string, fname string) (Service, error) {
	s := Service{
		Deleted:   false,
		AutoStart: true,
		Filename:  fname,
		Name:      ServiceName(fname),
	}

	if !ipc.IsCustASCII(fname) {
		return s, fmt.Errorf("%s has invalid service name '%s', only a-Z0-9_-. allowed", fname, s.Name)
	}

	Cfg, err := ini.InsensitiveLoad(fmt.Sprintf("%s/lutra.d/%s", baseDir, fname))
	if err != nil {
		clog.Error(2, "Failed to parse '%s': %v", fname, err)
		return s, err
	}
	Cfg.NameMapper = ini.TitleUnderscore

	sec, err := Cfg.GetSection("order")
	if err != nil {
		clog.Error(2, "service %s does not contains an order section", fname)
		return s, fmt.Errorf("service %s does not contains an order section", fname)
	}

	s.Requires = sec.Key("Requires").Strings(",")
	s.Before = sec.Key("Before").Strings(",")
	s.After = sec.Key("After").Strings(",")
	if strings.HasSuffix(string(s.Name), ".target") {
		s.WantedBy = sec.Key("WantedBy").MustString("")
	} else {
		s.WantedBy = sec.Key("WantedBy").MustString("multi-user.target")
	}

	sec, err = Cfg.GetSection("service")
	if err != nil {
		clog.Error(2, "service %s does not contains an service section", fname)
		return s, fmt.Errorf("service %s does not contains an service section", fname)
	}

	s.ExecPreStart = Command(sec.Key("ExecPreStart").MustString(""))
	s.ExecStart = Command(sec.Key("ExecStart").MustString(""))
	s.ExecPostStart = Command(sec.Key("ExecPostStart").MustString(""))
	s.ExecPreStop = Command(sec.Key("ExecPreStop").MustString(""))
	s.ExecStop = Command(sec.Key("ExecStop").MustString(""))
	s.ExecPostStop = Command(sec.Key("ExecPostStop").MustString(""))

	s.Description = sec.Key("Description").MustString("")
	s.PIDFile = sec.Key("PIDFile").MustString("")
	s.AutoStart = sec.Key("Autostart").MustBool(false)

	s.Type = sec.Key("Type").MustString("forking")
	if s.Type != "forking" && s.Type != "simple" && s.Type != "oneshot" && s.Type != "virtual" {
		clog.Error(2, "service %s invalid type: %s", fname, s.Type)
		return s, fmt.Errorf("service %s invalid type: %s", fname, s.Type)
	}

	// some sanity check
	// Must have an ExecStart, execept if it's a virtual service
	if s.Type != "virtual" && s.ExecStart == "" {
		return s, fmt.Errorf("service %s does not have an ExecStart command", fname)
	}

	// forking must have pidfile
	if s.Type == "forking" && s.PIDFile == "" {
		clog.Warn("service %s does not have a PIDFile, considers setting it", fname)
	}

	return s, err
}

// ParseServiceConfigs parse all the config in directory dir return a map of
// providers of ServiceTypes from that directory.
func ParseServiceConfigs(baseDir string, reloading bool) error {
	cfgsDir := fmt.Sprintf("%s/lutra.d", baseDir)

	files, err := ioutil.ReadDir(cfgsDir)
	if err != nil {
		return err
	}
	for _, fstat := range files {
		if fstat.IsDir() {
			// Mostly to skip "." and ".."
			continue
		}

		// We only want to parse files ending with .service
		if !strings.HasSuffix(fstat.Name(), ".service") &&
			!strings.HasSuffix(fstat.Name(), ".target") {
			continue
		}

		s, err := ParseConfig(baseDir, fstat.Name())
		if err != nil {
			clog.Error(2, err.Error())
			continue
		}

		// If we are not reloading, set initial state and actions
		if !reloading {
			s.State = NotStarted
			s.LastAction = Unknown
			s.LastActionAt = time.Now().UTC().Unix()

			LoadedServices[s.Name] = &s
		} else {
			// We are reloading, AND, the init service is still present, mark it as not-deleted
			// And also not overwrite the whole service, just update what could have changed

			LoadedServices[s.Name].Deleted = false
			LoadedServices[s.Name].Description = s.Description
			LoadedServices[s.Name].AutoStart = s.AutoStart
			LoadedServices[s.Name].PIDFile = s.PIDFile
			LoadedServices[s.Name].ExecPreStart = s.ExecPreStart
			LoadedServices[s.Name].Startup = s.Startup
			LoadedServices[s.Name].ExecPostStart = s.ExecPostStart
			LoadedServices[s.Name].ExecPreStop = s.ExecPreStop
			LoadedServices[s.Name].Shutdown = s.Shutdown
			LoadedServices[s.Name].ExecPostStop = s.ExecPostStop
			LoadedServices[s.Name].Type = s.Type
		}

	}
	return nil
}

// ParseSetupConfig parse the main configuration
func ParseSetupConfig(fname string) (err error) {
	Cfg, err := ini.InsensitiveLoad(fname)
	if err != nil {
		clog.Error(2, "Failed to parse '%s': %v", fname, err)
		return err
	}
	Cfg.NameMapper = ini.TitleUnderscore

	sec := Cfg.Section("global")
	MainConfig.Persist = sec.Key("Persist").MustBool(true)
	MainConfig.Autologins = sec.Key("Autologin").Strings(",")

	sec = Cfg.Section("logging")
	MainConfig.Log.Filename = sec.Key("filename").MustString("/var/log/lutrainit.log")
	MainConfig.Log.Daily = sec.Key("rotate_daily").MustBool(true)
	MainConfig.Log.MaxDays = sec.Key("max_days").MustInt64(7)
	MainConfig.Log.Rotate = sec.Key("rotate").MustBool(true)
	MainConfig.Log.MaxSize = sec.Key("max_size_shift").MustInt(28)
	MainConfig.Log.MaxLines = sec.Key("max_lines").MustInt64(1000000)
	MainConfig.Log.BufferLen = sec.Key("buffer_len").MustInt64(100)

	return err
}

// ReloadConfig both Main and Services ones
func ReloadConfig(reloading bool, baseDir string, withFile bool) (err error) {
	if reloading {
		clog.Info("Parsing configurations with reloading...")
	} else {
		clog.Info("Parsing configurations...")
	}

	if err := ParseSetupConfig(fmt.Sprintf("%s/lutra.conf", baseDir)); err != nil {
		clog.Error(2, "[lutra] Failed to parse Main Configuration: %s", err.Error())
		return err
	}
	clog.Info("Main config parsed.")

	if err = setupLogging(withFile); err != nil {
		clog.Error(2, "[lutra] Failed to re-setup logging: %s", err.Error())
		return err
	}
	clog.Info("Logging updated.")

	// Mark all as deleted
	for k := range LoadedServices {
		LoadedServices[k].Deleted = true
	}

	// Then re-parse
	err = ParseServiceConfigs(baseDir, reloading)
	if err != nil {
		clog.Error(2, "[lutra] Cannot re-parse service configs: %s", err.Error())
		return err
	}

	dissappeared := 0
	for k := range LoadedServices {
		if LoadedServices[k].Deleted {
			dissappeared++
		}
	}
	if dissappeared > 0 {
		clog.Info("[lutra] It looks like %d Services files dissappeared :|", dissappeared)
	}

	// TODO: sanity check that targets: basic, disk, network and multi-user are presents
	for _, s := range LoadedServices {
		if s.WantedBy != "" {
			if _, ok := LoadedServices[ServiceName(s.WantedBy)]; !ok {
				clog.Error(2, "service %s has inexistant WantedBy: %s", s.Name, s.WantedBy)
				return fmt.Errorf("service %s has inexistant WantedBy: %s", s.Name, s.WantedBy)
			}
		}

		for _, d := range s.Requires {
			if _, ok := LoadedServices[ServiceName(d)]; !ok {
				clog.Error(2, "service %s has inexistant Requires: %s", s.Name, d)
				return fmt.Errorf("service %s has inexistant Requires: %s", s.Name, d)
			}
		}

		for _, d := range s.After {
			if _, ok := LoadedServices[ServiceName(d)]; !ok {
				clog.Error(2, "service %s has inexistant After: %s", s.Name, d)
				return fmt.Errorf("service %s has inexistant After: %s", s.Name, d)
			}
		}

		for _, d := range s.Before {
			if _, ok := LoadedServices[ServiceName(d)]; !ok {
				clog.Error(2, "service %s has inexistant Before: %s", s.Name, d)
				return fmt.Errorf("service %s has inexistant Before: %s", s.Name, d)
			}
		}
	}

	clog.Info("Services re-parsed.")

	return err
}
