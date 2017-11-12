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

		StartedReexec	bool
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
				clog.Warn("[lutra] Invalid service %s type: %s, forcing Type=simple", s.Filename, serviceType)
				s.Type = "simple"
			}
		}
	} else if strings.HasPrefix(line, "Autostart:") {
		s.AutoStart = strings.TrimSpace(strings.TrimPrefix(line, "Autostart:")) == "true"
	} else if strings.HasPrefix(line, "ExecPreStart:") {
		if s.ExecPreStart == "" {
			s.ExecPreStart = Command(strings.TrimSpace(strings.TrimPrefix(line, "ExecPreStart:")))
		}
	} else if strings.HasPrefix(line, "ExecPreStop:") {
		if s.ExecPreStop == "" {
			s.ExecPreStop = Command(strings.TrimSpace(strings.TrimPrefix(line, "ExecPreStop:")))
		}
	} else if strings.HasPrefix(line, "ExecPostStart:") {
		if s.ExecPostStart == "" {
			s.ExecPostStart = Command(strings.TrimSpace(strings.TrimPrefix(line, "ExecPostStart:")))
		}
	} else if strings.HasPrefix(line, "ExecPostStop:") {
		if s.ExecPostStop == "" {
			s.ExecPostStop = Command(strings.TrimSpace(strings.TrimPrefix(line, "ExecPostStop:")))
		}
	}
	return nil
}

// ParseConfig a single config file into the services it provides
func ParseConfig(r io.Reader, filename string) (Service, error) {
	s := Service{
		Deleted: false,
		AutoStart: true,
		Filename: filename,
	}
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
			if s.AutoStart {
				StartupServices[t] = append(StartupServices[t], &StartupService{
					Name: s.Name,
					AutoStart: s.AutoStart,

					Needs: s.Needs,
					Provides: s.Provides,
				})
			}
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
			LoadedServices[s.Name].Needs = s.Needs
			LoadedServices[s.Name].Provides = s.Provides
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

	if service.Type == "forking" && service.PIDFile == "" {
		clog.Warn("service %s is type forking without PIDFile, consider setting it", service.Name)
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
func ReloadConfig(reloading bool, withFile bool) (err error) {
	if reloading {
		clog.Info("Parsing configurations with reloading...")
	} else {
		clog.Info("Parsing configurations...")
	}

	if err := ParseSetupConfig("/etc/lutrainit/lutra.conf"); err != nil {
		clog.Error(2,"[lutra] Failed to parse Main Configuration: %s", err.Error())
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
	err = ParseServiceConfigs("/etc/lutrainit/lutra.d", reloading)
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
	clog.Info("Services re-parsed.")

	return err
}
