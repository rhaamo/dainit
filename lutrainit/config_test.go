package main

import (
	"strings"
	"testing"
)

func compareService(s1, s2 Service) bool {
	if len(s1.Provides) != len(s2.Provides) {
		return false
	}
	if len(s1.Needs) != len(s2.Needs) {
		return false
	}
	for i := range s1.Provides {
		if s1.Provides[i] != s2.Provides[i] {
			return false
		}
	}
	for i := range s1.Needs {
		if s1.Needs[i] != s2.Needs[i] {
			return false
		}
	}
	return s1.Name == s2.Name && s1.Startup == s2.Startup && s1.Shutdown == s2.Shutdown
}

func TestServiceConfigParsing(t *testing.T) {
	tcs := []struct {
		Content string
		Service Service
	}{
		{
			// Test the basic parser (with no trailing newline)
			`Name: TestService
Needs: stuff
Provides: otherstuff
Startup: foo
Shutdown: bar`,
			Service{
				Name:     "TestService",
				Startup:  "foo",
				Shutdown: "bar",
				Needs:    []ServiceType{"stuff"},
				Provides: []ServiceType{"otherstuff"},
			},
		},
		{
			// Test a multi-provider
			`Name: TestService
Needs: stuff
Provides: otherstuff
Provides: otherstuff2
Shutdown: bar
Startup: foo
`,
			Service{
				Name:     "TestService",
				Startup:  "foo",
				Shutdown: "bar",
				Needs:    []ServiceType{"stuff"},
				Provides: []ServiceType{"otherstuff", "otherstuff2"},
			},
		},
		{
			// Other form of multiprovider
			`Name: TestService
Needs: stuff
Provides: otherstuff is also, otherstuff2
Startup: foo
Shutdown: bar
`,
			Service{
				Name:     "TestService",
				Startup:  "foo",
				Shutdown: "bar",
				Needs:    []ServiceType{"stuff"},
				Provides: []ServiceType{"otherstuff is also", "otherstuff2"},
			},
		},
	}
	for i, tc := range tcs {
		r := strings.NewReader(tc.Content)
		service, err := ParseConfig(r)
		if err != nil {
			t.Errorf("Unexpected error parsing test case %d: %v", i, err)
		}
		if !compareService(service, tc.Service) {
			t.Errorf("Unexpected value parsing test case %d: got %v want %v", i, service, tc.Service)
		}
	}
}
func TestSetupConfigParsing(t *testing.T) {
	tcs := []struct {
		Content   string
		Autologin []string
		Persist   bool
	}{
		{
			// Basic autologin
			`Autologin: testuser`,
			[]string{"testuser"},
			false,
		},
		{
			// Trailing newline and whitespace
			`Autologin: test2   
`,
			[]string{"test2"},
			false,
		},
		{
			// Multiple tty autologin
			`Autologin: test2   
Autologin: foo
`,
			[]string{"test2", "foo"},
			false,
		},
		{
			// Persist test
			`Persist: true`,
			nil,
			true,
		},
		{
			// Persist and autologin test
			"Autologin: foo\nAutologin:bar\nPersist: true",
			[]string{"foo", "bar"},
			true,
		},
	}
	for i, tc := range tcs {
		r := strings.NewReader(tc.Content)
		autologins, persist, err := ParseSetupConfig(r)
		if err != nil {
			t.Fatal(err)
		}

		if len(autologins) != len(tc.Autologin) {
			t.Errorf("Incorrect number of autologins for test case %d: got %v want %v", i, len(autologins), len(tc.Autologin))
		} else {
			for j := range tc.Autologin {
				if autologins[j] != tc.Autologin[j] {
					t.Errorf("Incorrect autologin for test case %d[%d]: got %v want %v", i, j, autologins[j], tc.Autologin[j])
				}
			}
		}
		if persist != tc.Persist {
			t.Errorf("Incorrect persistence for test case %d: got %v want %v", i, persist, tc.Persist)
		}
	}
}
