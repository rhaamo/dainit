package main_test

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/rhaamo/lutrainit/lutrainit"
	"strings"
	"testing"
	//"github.com/davecgh/go-spew/spew"
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


func Test_ParseConfig(t *testing.T) {
	Convey("Basic parser without trailing newline", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService
Needs: stuff
Provides: otherstuff
Startup: foo
Shutdown: bar
Type: forking`,
				Service{
					Name:     "TestService",
					Startup:  "foo",
					Shutdown: "bar",
					Needs:    []ServiceType{"stuff"},
					Provides: []ServiceType{"otherstuff"},
					Type:     "forking",
				},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			service, err := ParseConfig(r, "test-service.service")
			So(err, ShouldBeNil)

			So(compareService(service, tc.Service), ShouldBeTrue)
		}
	})

	Convey("Basic parser with trailing newline", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService
Needs: stuff
Provides: otherstuff
Startup: foo
Shutdown: bar
Type: forking
`,
				Service{
					Name:     "TestService",
					Startup:  "foo",
					Shutdown: "bar",
					Needs:    []ServiceType{"stuff"},
					Provides: []ServiceType{"otherstuff"},
					Type:     "forking",
				},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			service, err := ParseConfig(r, "test-service.service")
			So(err, ShouldBeNil)

			So(compareService(service, tc.Service), ShouldBeTrue)
		}
	})

	Convey("Test multi-provider", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService
Needs: stuff
Provides: otherstuff
Provides: otherstuff2
Shutdown: bar
Startup: foo
Type: simple
`,
				Service{
					Name:     "TestService",
					Startup:  "foo",
					Shutdown: "bar",
					Needs:    []ServiceType{"stuff"},
					Provides: []ServiceType{"otherstuff", "otherstuff2"},
					Type:	  "simple",
				},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			service, err := ParseConfig(r, "test-service.service")
			So(err, ShouldBeNil)

			So(compareService(service, tc.Service), ShouldBeTrue)
		}
	})

	Convey("Other form of multi-provider", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
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
					Type:     "simple",
				},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			service, err := ParseConfig(r, "test-service.service")
			So(err, ShouldBeNil)

			So(compareService(service, tc.Service), ShouldBeTrue)
		}
	})

	Convey("Full config with valid strings", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService-42.0
Needs: stuff-1.0
Provides: otherstuff, foobar-1.0
Startup: foo
Shutdown: bar
Type: oneshot
`,
				Service{
					Name:     "TestService-42.0",
					Startup:  "foo",
					Shutdown: "bar",
					Needs:    []ServiceType{"stuff-1.0"},
					Provides: []ServiceType{"otherstuff", "foobar-1.0"},
					Type:     "oneshot",
				},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			service, err := ParseConfig(r, "test-service.service")
			So(err, ShouldBeNil)

			So(compareService(service, tc.Service), ShouldBeTrue)
		}
	})

	Convey("Invalid name", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestSe$@#rvice-42.0
Needs: stuff-1.0
Provides: otherstuff, foobar-1.0
Startup: foo
Shutdown: bar
Type: oneshot
`,
				Service{},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			_, err := ParseConfig(r, "test-service.service")
			So(err, ShouldNotBeNil)
		}
	})

	Convey("Invalid Provides", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService-42.0
Needs: stuff-1.0
Provides: other$#@stuff, foobar-1.0
Startup: foo
Shutdown: bar
Type: oneshot
`,
				Service{},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			_, err := ParseConfig(r, "test-service.service")
			So(err, ShouldNotBeNil)
		}
	})

	Convey("Invalid Needs", t, func() {
		testCases := []struct {
			Content string
			Service Service
		}{
			{
				`Name: TestService-42.0
Needs: stu$#@ff-1.0
Provides: otherstuff, foobar-1.0
Startup: foo
Shutdown: bar
Type: oneshot
`,
				Service{},
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			_, err := ParseConfig(r, "test-service.service")
			So(err, ShouldNotBeNil)
		}
	})
}

func Test_ParseSetupConfig(t *testing.T) {
	Convey("Basic autologin", t, func() {
		testCases := []struct {
			Content string
			Autologin []string
			Persist bool
		}{
			{
				`Autologin: testuser`,
				[]string{"testuser"},
				false,
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			autologins, persist, err := ParseSetupConfig(r)
			So(err, ShouldBeNil)

			So(len(autologins), ShouldEqual, len(tc.Autologin))

			for j := range tc.Autologin {
				So(autologins[j], ShouldEqual, tc.Autologin[j])
			}

			So(persist, ShouldEqual, tc.Persist)
		}
	})

	Convey("Trailing newline and whitespace", t, func() {
		testCases := []struct {
			Content string
			Autologin []string
			Persist bool
		}{
			{
				`Autologin: test2
	`,
				[]string{"test2"},
				false,
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			autologins, persist, err := ParseSetupConfig(r)
			So(err, ShouldBeNil)

			So(len(autologins), ShouldEqual, len(tc.Autologin))

			for j := range tc.Autologin {
				So(autologins[j], ShouldEqual, tc.Autologin[j])
			}

			So(persist, ShouldEqual, tc.Persist)
		}
	})

	Convey("Multiple tty autologin", t, func() {
		testCases := []struct {
			Content string
			Autologin []string
			Persist bool
		}{
			{
				`Autologin: test2
Autologin: foo
`,
				[]string{"test2", "foo"},
				false,
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			autologins, persist, err := ParseSetupConfig(r)
			So(err, ShouldBeNil)

			So(len(autologins), ShouldEqual, len(tc.Autologin))

			for j := range tc.Autologin {
				So(autologins[j], ShouldEqual, tc.Autologin[j])
			}

			So(persist, ShouldEqual, tc.Persist)
		}
	})

	Convey("Persist test", t, func() {
		testCases := []struct {
			Content string
			Autologin []string
			Persist bool
		}{
			{
				`Persist: true`,
				nil,
				true,
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			autologins, persist, err := ParseSetupConfig(r)
			So(err, ShouldBeNil)

			So(len(autologins), ShouldEqual, len(tc.Autologin))

			for j := range tc.Autologin {
				So(autologins[j], ShouldEqual, tc.Autologin[j])
			}

			So(persist, ShouldEqual, tc.Persist)
		}
	})

	Convey("Persist and autologin test", t, func() {
		testCases := []struct {
			Content string
			Autologin []string
			Persist bool
		}{
			{
				"Autologin: foo\nAutologin:bar\nPersist: true",
				[]string{"foo", "bar"},
				true,
			},
		}

		for _, tc := range testCases {
			r := strings.NewReader(tc.Content)
			autologins, persist, err := ParseSetupConfig(r)
			So(err, ShouldBeNil)

			So(len(autologins), ShouldEqual, len(tc.Autologin))

			for j := range tc.Autologin {
				So(autologins[j], ShouldEqual, tc.Autologin[j])
			}

			So(persist, ShouldEqual, tc.Persist)
		}
	})


}
