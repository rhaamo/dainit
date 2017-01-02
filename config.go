package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type ServiceName string
type ServiceType string

type Command string

func (c Command) String() string {
	return string(c)
}

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
	} else if strings.HasPrefix(line, "# ") {
		if s.Name == "" {
			s.Name = ServiceName(strings.TrimSpace(strings.TrimPrefix(line, "# ")))
		}
	}
	return nil
}

func parseSetupLine(line string) (autologin string, persist bool) {
	if strings.HasPrefix(line, "Autologin:") {
		return strings.TrimSpace(strings.TrimPrefix(line, "Autologin:")), false
	} else if strings.HasPrefix(line, "Persist:") {
		return "", strings.TrimSpace(strings.TrimPrefix(line, "Persist:")) == "true"
	}
	return "", false
}

// Parses a single config file into the services it provides
func ParseConfig(r io.Reader) (Service, error) {
	s := Service{}
	var line string
	var err error
	scanner := bufio.NewReader(r)

	for {
		line, err = scanner.ReadString('\n')
		switch err {
		case io.EOF:
			if err := parseLine(line, &s); err != nil {
				log.Println(err)
			}
			return s, nil
		case nil:
			if err := parseLine(line, &s); err != nil {
				log.Println(err)
			}
		default:
			return Service{}, err
		}
	}
}

// Parses all the config in directory dir return a map of
// providers of ServiceTypes from that directory.
func ParseServiceConfigs(dir string) (map[ServiceType][]*Service, error) {
	providers := make(map[ServiceType][]*Service)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fstat := range files {
		if fstat.IsDir() {
			// Mostly to skip "." and ".."
			continue
		}
		f, err := os.Open(dir + "/" + fstat.Name())
		if err != nil {
			log.Println(err)
		}
		s, err := ParseConfig(f)
		f.Close()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, t := range s.Provides {
			providers[t] = append(providers[t], &s)
		}

	}
	return providers, nil
}

// Parses the file file for "Autologin:" or "Persist:" lines.
func ParseSetupConfig(r io.Reader) (autologins []string, persist bool, err error) {
	scanner := bufio.NewReader(r)
	for {
		line, err2 := scanner.ReadString('\n')
		switch err2 {
		case io.EOF:
			autologin, persist2 := parseSetupLine(line)
			if autologin != "" {
				autologins = append(autologins, autologin)
			}
			if persist2 {
				persist = persist2
			}
			return
		case nil:
			autologin, persist2 := parseSetupLine(line)
			if autologin != "" {
				autologins = append(autologins, autologin)
			}
			if persist2 {
				persist = persist2
			}
		default:
			err = err2
			return
		}
	}
	return
}
