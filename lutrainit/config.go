package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"github.com/go-clog/clog"
	"os"
	"strings"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"time"
	"github.com/go-ini/ini"
)

// ServiceName defines the service name
type ServiceName string

// ServiceType defines the service type
type ServiceType string

// Command defines a command string used by Startup or Shutdown
type Command string

func (c Command) String() string {
	return string(c)
}

var (
	// MainConfig of the daemon
	MainConfig struct {
		Persist		bool
		Autologins	[]string

		Log struct {
			Filename 	string
			Rotate   	bool
			Daily   	bool
			MaxSize	 	int
			MaxLines 	int64
			MaxDays  	int64
			BufferLen	int64
		}
	}
)

func parseLine(line string, s *Service) error {
	if strings.HasPrefix(line, "Needs:") {
		specified := strings.Split(strings.TrimPrefix(line, "Needs:"), ",")
		for _, nd := range specified {
			s.Needs = append(s.Needs, ServiceType(strings.TrimSpace(nd)))
		}
	} else if strings.HasPrefix(line, "Provides:") {
		specified := strings.Split(strings.TrimPrefix(line, "Provides:"), ",")
		for _, nd := range specified {
			s.Provides = append(s.Provides, ServiceType(strings.TrimSpace(nd)))
		}
	} else if strings.HasPrefix(line, "Startup:") {
		if s.Startup != "" {
			return fmt.Errorf("Startup already set")
		}
		s.Startup = Command(strings.TrimSpace(strings.TrimPrefix(line, "Startup:")))
	} else if strings.HasPrefix(line, "Shutdown:") {
		if s.Shutdown != "" {
			return fmt.Errorf("Shutdown already set")
		}
		s.Shutdown = Command(strings.TrimSpace(strings.TrimPrefix(line, "Shutdown:")))
	} else if strings.HasPrefix(line, "Name:") {
		if s.Name == "" {
			s.Name = ServiceName(strings.TrimSpace(strings.TrimPrefix(line, "Name:")))
		}
	} else if strings.HasPrefix(line, "Description:") {
		if s.Description == "" {
			s.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		}
	} else if strings.HasPrefix(line, "PIDFile:") {
		if s.PIDFile == "" {
			s.PIDFile = strings.TrimSpace(strings.TrimPrefix(line, "PIDFile:"))
		}
	} else if strings.HasPrefix(line, "Type:") {
		if s.Type == "" {
			serviceType := strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
			switch serviceType {
			case "simple":
				s.Type = "simple"
			case "forking":
				s.Type = "forking"
			case "oneshot":
				s.Type = "oneshot"
			default:
				clog.Warn("[lutra] Invalid service type: %s, forcing Type=simple", serviceType)
				s.Type = "simple"
			}
		}
	}
	return nil
}

// ParseConfig a single config file into the services it provides
func ParseConfig(r io.Reader, filename string) (Service, error) {
	s := Service{}
	var line string
	var err error
	scanner := bufio.NewReader(r)

	for {
		line, err = scanner.ReadString('\n')
		switch err {
		case io.EOF:
			if err := parseLine(line, &s); err != nil {
				clog.Error(2, err.Error())
			}

			// Check for configuration sanity before returning
			if err := checkSanity(&s, filename); err != nil {
				return Service{}, err
			}

			return s, nil
		case nil:
			if err := parseLine(line, &s); err != nil {
				clog.Error(2, err.Error())
			}
		default:
			return Service{}, err
		}
	}
}

// ParseServiceConfigs parse all the config in directory dir return a map of
// providers of ServiceTypes from that directory.
func ParseServiceConfigs(dir string, reloading bool) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fstat := range files {
		if fstat.IsDir() {
			// Mostly to skip "." and ".."
			continue
		}

		// We only want to parse files ending with .service
		if !strings.HasSuffix(fstat.Name(), ".service") {
			continue
		}

		f, err := os.Open(dir + "/" + fstat.Name())
		if err != nil {
			clog.Error(2, err.Error())
			continue
		}
		s, err := ParseConfig(f, fstat.Name())
		f.Close()
		if err != nil {
			clog.Error(2, err.Error())
			continue
		}

		for _, t := range s.Provides {
			StartupServices[t] = append(StartupServices[t], &s)
		}

		// Populate the Loaded Services thingy
		ipcLoadedService :=  &ipc.LoadedService{
			Name: ipc.ServiceName(s.Name),
			Description: s.Description,
			PIDFile: s.PIDFile,
			Type: s.Type,
			Deleted: false,
		}

		// If we are not reloading, set initial state and actions
		if !reloading {
			ipcLoadedService.State = ipc.NotStarted
			ipcLoadedService.LastAction = ipc.Unknown
			ipcLoadedService.LastActionAt = time.Now().UTC().Unix()
		} else {
			// We are reloading, AND, the init service is still present, mark it as not-deleted
			ipcLoadedService.Deleted = false
		}

		LoadedServices[ipc.ServiceName(s.Name)] = ipcLoadedService

	}
	return nil
}

func checkSanity(service *Service, filename string) error {

	if ! ipc.IsCustASCII(string(service.Name)) {
		return fmt.Errorf("%s has invalid service name '%s', only a-Z0-9_-. allowed", filename, service.Name)
	}

	for _, provide := range service.Provides {
		if !ipc.IsCustASCIISpace(string(provide)) {
			return fmt.Errorf("%s has invalid provides '%s', only a-Z0-9_-. allowed", filename, provide)
		}
	}

	for _, need := range service.Needs {
		if !ipc.IsCustASCIISpace(string(need)) {
			return fmt.Errorf("%s has invalid needs '%s', only a-Z0-9_-. allowed", filename, need)
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
	MainConfig.Persist = sec.Key("Persist").MustBool(false)
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
func ReloadConfig() (err error) {
	clog.Info("Reloading configurations...")
	if err := ParseSetupConfig("/etc/lutrainit/lutra.conf"); err != nil {
		clog.Error(2,"[lutra] Failed to parse Main Configuration: %s", err.Error())
		return err
	}
	clog.Info("Main config reloaded.")

	if err = setupLogging(true); err != nil {
		clog.Error(2, "[lutra] Failed to re-setup logging: %s", err.Error())
		return err
	}
	clog.Info("Logging updated.")

	// Mark all as deleted
	for k, _ := range LoadedServices {
		LoadedServices[k].Deleted = true
	}

	// Then re-parse
	err = ParseServiceConfigs("/etc/lutrainit/lutra.d", true)
	if err != nil {
		clog.Error(2, "[lutra] Cannot re-parse service configs: %s", err.Error())
		return err
	}

	dissappeared := 0
	for k, _ := range LoadedServices {
		if LoadedServices[k].Deleted {
			dissappeared++
		}
	}
	if dissappeared > 0 {
		clog.Info("[lutra] It looks like %d Services files dissappeared :|", dissappeared)
	}
	clog.Info("Services reloaded.")

	return err
}
