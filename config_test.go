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
func TestConfigParsing(t *testing.T) {
	tcs := []struct {
		Content string
		Service Service
	}{
		{
			// Test the basic parser (with no trailing newline)
			`# TestService
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
			`# TestService
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
			`# TestService
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
